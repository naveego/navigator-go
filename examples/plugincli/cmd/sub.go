// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/naveego/navigator-go/subscribers/client"
	"github.com/naveego/navigator-go/subscribers/protocol"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var subscriberConn io.ReadWriteCloser
var subscriber protocol.Subscriber

// subCmd represents the sub command
var subCmd = &cobra.Command{
	Use:   "sub",
	Short: "CLI REPL for a subscriber.",
	Long:  `Allows you to send commands to a subscriber plugin.`,
	RunE: func(cmd *cobra.Command, args []string) error {

		var err error

		fmt.Printf("Configuration: %#v\n", viper.AllSettings())

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		subscriberPath := viper.GetString("plugin")
		if subscriberPath != "" {
			subcmd := exec.CommandContext(ctx, viper.GetString("plugin"), viper.GetString("addr"))

			subcmd.Stdout = os.Stdout

			err := subcmd.Start()
			check(err)
		}

		<-time.After(time.Second * 1)

		connectSubscriber()

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
				if err != nil {
					fmt.Println("invalid selection")
					continue
				}

				switch choice {
				case 1:
					message := protocol.TestConnectionRequest{}
					err = readMessage(&message)
					if err == nil {
						writeSubscriberResponse(subscriber.TestConnection(message))
					}
				case 2:
					message := protocol.InitRequest{}
					err = readMessage(&message)
					if err == nil {
						writeSubscriberResponse(subscriber.Init(message))
					}
				case 3:
					message := protocol.ReceiveShapeRequest{}
					err = readMessage(&message)
					if err == nil {
						writeSubscriberResponse(subscriber.ReceiveDataPoint(message))
					}
				case 4:
					message := protocol.DisposeRequest{}
					err = readMessage(&message)
					if err == nil {
						writeSubscriberResponse(subscriber.Dispose(message))
					}
				case 5:
					message := protocol.DiscoverShapesRequest{}
					err = readMessage(&message)
					if err == nil {
						writeSubscriberResponse(subscriber.DiscoverShapes(message))
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
	RootCmd.AddCommand(subCmd)

	viper.BindPFlags(RootCmd.PersistentFlags())
}

func writeSubscriberResponse(resp interface{}, err error) {
	if err != nil {
		fmt.Println("subscriber error: ", err)
		connectSubscriber()
	}

	fmt.Print("\033[32mResponse:\033[0m")
	fmt.Println()
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", " ")
	encoder.Encode(resp)

	fmt.Println()
}

func connectSubscriber() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println("Connection borked (is the plugin running?): ", err)
		}
	}()

	var err error
	fmt.Println("connecting...")

	if subscriberConn != nil {
		subscriberConn.Close()
	}

	subscriberConn, err = DefaultConnectionFactory(viper.GetString("addr"))
	check(err)

	subscriber, err = client.NewSubscriber(subscriberConn)
	check(err)

	fmt.Println("connected")
}
