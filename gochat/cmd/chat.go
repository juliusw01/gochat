package cmd

import (
	"gochat/chat"
	//"net/http"

	//"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var chatCmd = &cobra.Command{
    Use:   "chat",
    Short: "Chat with other user via chatrooms or message them directly",
    //Args:  cobra.MinimumNArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        chat.StartClient()
        
    },
}

func init() {
    rootCmd.AddCommand(chatCmd)
}
