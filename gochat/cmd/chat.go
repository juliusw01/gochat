package cmd

import (
	"gochat/chat"
	"log"
	//"net/http"

	//"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
    Use:   "chat",
    Short: "Chat with other user via chatrooms or message them directly",
    //Args:  cobra.MinimumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        username, err := cmd.Flags().GetString("username")
		if err != nil {
			log.Fatal(err)
		}

        chat.StartClient(username, false)
        
    },
}

func init() {
    rootCmd.AddCommand(chatCmd)
    chatCmd.Flags().StringP("username", "u", "", "Your username to be used while chatting (required)")
	chatCmd.MarkFlagRequired("username")
}
