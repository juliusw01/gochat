package cmd

import (
	//"gochat/crypto"
	"gochat/auth"
	"log"
	"syscall"

	"golang.org/x/term"

	//"net/http"

	//"github.com/gorilla/websocket"
	"fmt"

	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "authenticate",
	Short: "Authenticate yourself as a user with username & password to get an access token",
	//Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		username, err := cmd.Flags().GetString("username")
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Please enter your password: ")
		bytePw, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatalf("Error reading password %v", err)
		}
		password := string(bytePw)
		auth.UserLogin(username, password)
	},
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.Flags().StringP("username", "u", "", "Your username to be used while chatting (required)")
	authCmd.MarkFlagRequired("username")
}
