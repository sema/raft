package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var flagRemoteAddress string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "srctl",
	Short: "SR command line tool",
	Long:  `Command line tool for the SR key value store.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagRemoteAddress, "server", "", "address of remote server, e.g. 127.0.0.1:8000")

	rootCmd.MarkFlagRequired("server")
}
