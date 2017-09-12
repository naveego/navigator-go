// Copyright © 2017 Naveego

package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "subtester",
	Short: "Interactive subscriber client",
	Long:  ``,
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
	RootCmd.PersistentFlags().String("plugin", "", "optional; path to subscriber executable if it's not already running")
	RootCmd.PersistentFlags().String("addr", "", "address to use to send messages to the plugin")
	RootCmd.PersistentFlags().String("listen-addr", "tcp://:50002", "address used to listen for incoming messages from the plugin")

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
