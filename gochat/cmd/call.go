//go:build darwin
// +build darwin

package cmd

import (
	"gochat/call"
	"log"
	//"net/http"

	//"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var callCmd = &cobra.Command{
    Use:   "call",
    Short: "Call other users",
    //Args:  cobra.MinimumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        username, err := cmd.Flags().GetString("username")
		if err != nil {
			log.Fatal(err)
		}
		recipient, err := cmd.Flags().GetString("recipient")
		if err != nil {
			log.Fatal(err)
		}

        call.Call(username, recipient)
    },
}

func init() {
    rootCmd.AddCommand(callCmd)
    callCmd.Flags().StringP("username", "u", "", "Your username to be used while chatting (required)")
	callCmd.MarkFlagRequired("username")
	callCmd.Flags().StringP("recipient", "r", "", "The person you wish to call (required)")
	callCmd.MarkFlagRequired("recipient")
}
