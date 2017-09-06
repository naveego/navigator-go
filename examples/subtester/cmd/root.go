// Copyright © 2017 Naveego

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/naveego/navigator-go/subscribers/protocol"

	"github.com/naveego/navigator-go/subscribers/client"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "subtester",
	Short: "Interactive subscriber client",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {

		fmt.Printf("Configuration: %#v\n", viper.AllSettings())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		subcmd := exec.CommandContext(ctx, viper.GetString("subscriber-path"), viper.GetString("addr"))

		subcmd.Stdout = os.Stdout

		err := subcmd.Start()
		check(err)

		<-time.After(time.Second * 1)

		conn, err := DefaultConnectionFactory(viper.GetString("addr"))
		check(err)
		defer conn.Close()

		subscriber, err := client.NewSubscriber(conn)
		check(err)

		go func() {
			for {
				fmt.Fprintln(os.Stdout, "Choose Method:")
				fmt.Fprintln(os.Stdout, " 1: TestConnection")
				fmt.Fprintln(os.Stdout, " 2: Init")
				fmt.Fprintln(os.Stdout, " 3: ReceiveShape")
				fmt.Fprintln(os.Stdout, " 4: Dispose")
				fmt.Fprintln(os.Stdout, " 5: DiscoverShapes")
				fmt.Print("\033[32mmethod:\033[0m ")
				choice := 0

				_, err = fmt.Fscanf(os.Stdin, "%d", &choice)
				check(err)

				switch choice {
				case 1:
					message := protocol.TestConnectionRequest{}
					err = readMessage(&message)
					if err == nil {
						writeResponse(subscriber.TestConnection(message))
					}
				case 2:
					message := protocol.InitRequest{}
					err = readMessage(&message)
					if err == nil {
						writeResponse(subscriber.Init(message))
					}
				case 3:
					message := protocol.ReceiveShapeRequest{}
					err = readMessage(&message)
					if err == nil {
						writeResponse(subscriber.ReceiveDataPoint(message))
					}
				case 4:
					message := protocol.DisposeRequest{}
					err = readMessage(&message)
					if err == nil {
						writeResponse(subscriber.Dispose(message))
					}
				case 5:
					message := protocol.DiscoverShapesRequest{}
					err = readMessage(&message)
					if err == nil {
						writeResponse(subscriber.DiscoverShapes(message))
					}
				default:
					fmt.Println("\033[31mnot understood\033[0m")
					_, _ = fmt.Scanln()
				}
			}
		}()

		done := awaitShutdown()

		<-done
	},
}

func readMessage(message interface{}) error {
	fmt.Println("Message structure:")
	encoder := json.NewEncoder(os.Stdout)
	encoder.Encode(message)
	fmt.Print("Message: ")
	decoder := json.NewDecoder(os.Stdin)
	err := decoder.Decode(message)

	if err != nil {
		fmt.Println("input invalid: ", err)
	}
	return err
}

func writeResponse(resp interface{}, err error) {
	if err != nil {
		fmt.Println("subscriber error: ", err)
	}

	fmt.Print("\033[32mResponse:\033[0m")
	fmt.Println()
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", " ")
	encoder.Encode(resp)

	fmt.Println()
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func awaitShutdown() <-chan bool {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	// `signal.Notify` registers the given channel to
	// receive notifications of the specified signals.
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// This goroutine executes a blocking receive for
	// signals. When it gets one it'll print it out
	// and then notify the program that it can finish.
	go func() {
		sig := <-sigs
		fmt.Println()
		fmt.Println(sig)
		done <- true
	}()

	fmt.Println("CTRL-C to quit")

	return done
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.subtester.yaml)")
	RootCmd.PersistentFlags().String("subscriber-path", "", "path to subscriber executable")
	RootCmd.PersistentFlags().String("addr", "", "address to use")

	viper.BindPFlags(RootCmd.PersistentFlags())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {

		// Search config in home directory with name ".subtester" (without extension).
		viper.AddConfigPath(".")
		viper.SetConfigName(".subtester")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
