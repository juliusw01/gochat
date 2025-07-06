package cmd

import (
    "fmt"
    "github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
    Use:   "gochat",
    Short: "gochat is a simple chat room application",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("Run `gochar help` for available commands")
    },
}

func Execute() {
    cobra.CheckErr(rootCmd.Execute())
}