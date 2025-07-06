package cmd

import (
	"gochat/auth"
	//"net/http"

	//"github.com/gorilla/websocket"
	"fmt"

	"github.com/spf13/cobra"
)

var regCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user to start chatting",
	//Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username, err := cmd.Flags().GetString("username")
		if err != nil {
			fmt.Println(err)
		}
		password, err := cmd.Flags().GetString("password")
		if err != nil {
			fmt.Println(err)
		}
		//auth.UserLogin(username, password)
		auth.RegUser(username, password)
	},
}

func init() {
	rootCmd.AddCommand(regCmd)
	regCmd.Flags().StringP("username", "u", "", "Your username to be used while chatting (required)")
	regCmd.Flags().StringP("password", "p", "", "Your password for the given user (required)")
	regCmd.MarkFlagRequired("username")
	regCmd.MarkFlagRequired("password")
}
