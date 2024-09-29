package main

import (
	"fmt"
	"os"

	"github.com/tadeasf/go_stream/src/commands"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "go_stream",
	Short:   "A simple video streaming server",
	Version: "1.0.0",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(commands.ServeCmd)
	rootCmd.AddCommand(commands.BasicAuthCmd)
}
