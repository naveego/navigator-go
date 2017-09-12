package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/naveego/api/types/pipeline"

	"github.com/naveego/navigator-go/publishers/client"
	"github.com/naveego/navigator-go/publishers/protocol"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var publisherConn io.ReadWriteCloser
var publisher client.PublisherProxy
var datapointCollector *client.DataPointCollector
var publishedDataPoints chan []pipeline.DataPoint

// subCmd represents the sub command
var pubCmd = &cobra.Command{
	Use:   "pub",
	Short: "CLI REPL for a publisher.",
	Long:  `Allows you to send commands to a publisher plugin.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		var err error

		fmt.Printf("Configuration: %#v\n", viper.AllSettings())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		publisherPath := viper.GetString("plugin")
		if publisherPath != "" {
			pubcmd := exec.CommandContext(ctx, viper.GetString("plugin"), viper.GetString("addr"))

			pubcmd.Stdout = os.Stdout

			err := pubcmd.Start()
			check(err)
		}

		<-time.After(time.Second * 1)

		connectPublisher()

		go func() {
			for {
				fmt.Fprintln(os.Stdout, " 1: TestConnection")
				fmt.Fprintln(os.Stdout, " 2: Init")
				fmt.Fprintln(os.Stdout, " 3: Publish")
				fmt.Fprintln(os.Stdout, " 4: Dispose")
				fmt.Fprintln(os.Stdout, " 5: DiscoverShapes")
				fmt.Print("\033[32mmethod:\033[0m ")
				choice := 0

				_, err = fmt.Fscanf(os.Stdin, "%d", &choice)
				if err != nil {
					fmt.Println("invalid selection")
					continue
				}

				switch choice {
				case 1:
					message := protocol.TestConnectionRequest{}
					err = readMessage(&message)
					if err == nil {
						writePublisherResponse(publisher.TestConnection(message))
					}
				case 2:
					message := protocol.InitRequest{}
					err = readMessage(&message)
					if err == nil {
						writePublisherResponse(publisher.Init(message))
					}
				case 3:
					message := protocol.PublishRequest{}
					err = readMessage(&message)
					if err == nil {
						writePublisherResponse(publisher.Publish(message))
					}
				case 4:
					message := protocol.DisposeRequest{}
					err = readMessage(&message)
					if err == nil {
						writePublisherResponse(publisher.Dispose(message))
					}
				case 5:
					message := protocol.DiscoverShapesRequest{}
					err = readMessage(&message)
					if err == nil {
						writePublisherResponse(publisher.DiscoverShapes(message))
					}
				default:
					fmt.Println("\033[31mnot understood\033[0m")
					_, _ = fmt.Scanln()
				}
			}
		}()

		done := awaitShutdown()

		<-done

		return err
	},
}

func init() {
	RootCmd.AddCommand(pubCmd)

}

func writePublisherResponse(resp interface{}, err error) {
	if err != nil {
		fmt.Println("publisher error: ", err)
		connectSubscriber()
	}

	fmt.Print("\033[32mResponse:\033[0m")
	fmt.Println()
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", " ")
	encoder.Encode(resp)

	fmt.Println()
}

func connectPublisher() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Connection borked (is the plugin running?): ", err)
		}
	}()

	var err error
	fmt.Println("connecting...")

	if publisherConn != nil {
		publisherConn.Close()

	}

	if datapointCollector != nil {
		datapointCollector.Stop()
	}

	listenAddr := viper.GetString("listen-addr")

	datapointCollector, err := client.NewDataPointCollector(listenAddr)
	check(err)

	publishedDataPoints = make(chan []pipeline.DataPoint, 100)
	err = datapointCollector.Start(publishedDataPoints)
	check(err)

	go func() {
		for msg := range publishedDataPoints {
			fmt.Println("Got message: ", msg)
		}
	}()

	publisherConn, err = DefaultConnectionFactory(viper.GetString("addr"))
	check(err)

	publisher, err = client.NewPublisher(publisherConn)
	check(err)

	fmt.Println("connected")
}

type dataPointHandler struct {
	dataPoints chan pipeline.DataPoint
}
