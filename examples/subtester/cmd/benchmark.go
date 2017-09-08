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
	"errors"
	"fmt"
	"time"

	"github.com/naveego/api/types/pipeline"
	"github.com/naveego/navigator-go/subscribers/protocol"
	"github.com/spf13/viper"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cobra"
)

// benchmarkCmd represents the benchmark command
var benchmarkCmd = &cobra.Command{
	Use:   "benchmark",
	Short: "Benchmarks a subscriber.",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		if viper.ConfigFileUsed() == "" {
			return errors.New("you must specify a config file when benchmarking")
		}

		if !viper.IsSet("benchmark") {
			return errors.New("you must specify a 'benchmark' node in the config file")
		}

		connect()

		initRequest := protocol.InitRequest{
			Settings: viper.GetStringMap("benchmark.init"),
		}

		fmt.Println("initRequest", initRequest)

		_, err := subscriber.Init(initRequest)

		if err != nil {
			return fmt.Errorf("couldn't init subscriber: %s", err)
		}

		reps := viper.GetInt("benchmark.reps")

		fmt.Printf("about to send datapoint %d times", reps)
		fmt.Println()

		request := protocol.ReceiveShapeRequest{
			DataPoint: pipeline.DataPoint{
				Data: make(map[string]interface{}),
			},
		}

		shapeMap := viper.GetStringMap("benchmark.shape")
		mapstructure.Decode(shapeMap, &request.Shape)
		datapointValues := viper.Get("benchmark.datapointValues").([]interface{})
		// fmt.Printf("DataPointMap: %#v", datapointValues)
		// fmt.Println()

		for _, x := range datapointValues {
			kv := x.(map[interface{}]interface{})
			request.DataPoint.Data[kv["name"].(string)] = kv["value"]
		}

		// fmt.Printf("%#v", request)
		// fmt.Println()

		counterKey := request.Shape.Keys[0]
		counter := request.DataPoint.Data[counterKey].(int)
		max := counter + reps

		fmt.Printf("beginning with %s set to %d", counterKey, counter)

		startTime := time.Now()

		for ; counter < max; counter++ {
			fmt.Println(counter)
			request.DataPoint.Data[counterKey] = counter

			_, err = subscriber.ReceiveDataPoint(request)
			if err != nil {
				return fmt.Errorf("error sending datapoint: %e\r\n%#v", err, request)
			}
			request.DataPoint.Data[counterKey] = counter
		}

		elapsed := time.Since(startTime)

		totalSeconds := elapsed.Seconds()

		secondsPerDataPoint := totalSeconds / float64(reps)

		fmt.Printf("took %.3f seconds to process %d datapoints\r\n", totalSeconds, reps)
		fmt.Printf("took %.2f ms per datapoint\r\n", secondsPerDataPoint*1000)

		return nil
	},
	PostRun: func(cmd *cobra.Command, args []string) {
		if subscriber != nil {
			_, _ = subscriber.Dispose(protocol.DisposeRequest{})
		}
	},
}

func init() {
	RootCmd.AddCommand(benchmarkCmd)
}
