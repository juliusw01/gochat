package cmd

import (
	"gochat/auth"
	//"net/http"

	//"github.com/gorilla/websocket"
	"fmt"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "authenticate",
	Short: "Authenticate yourself as a user with username & password to get an access token",
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
		auth.UserLogin(username, password)
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.Flags().StringP("username", "u", "", "Your username to be used while chatting (required)")
	authCmd.Flags().StringP("password", "p", "", "Your password for the given user (required)")
	authCmd.MarkFlagRequired("username")
	authCmd.MarkFlagRequired("password")
}
