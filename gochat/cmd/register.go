package cmd

import (
	"fmt"
	"gochat/auth"
	"gochat/crypto"
	"log"
	"syscall"

	"golang.org/x/term"

	//"net/http"

	//"github.com/gorilla/websocket"

	"github.com/spf13/cobra"
)

var regCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new user to start chatting",
	//Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		username, err := cmd.Flags().GetString("username")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Please enter a password: ")
		bytePw, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatalf("Error reading password %v", err)
		}
		password := string(bytePw)
		fmt.Println("Confirm password: ")
		byteConfirm, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			log.Fatalf("Error reading password %v", err)
		}
		confirmPw := string(byteConfirm)
		if confirmPw != password {
			log.Fatalf("Passwords were not identical")
		}

		auth.RegUser(username, password)
		fmt.Println("Logging in...")
		auth.UserLogin(username, password)
		fmt.Println("Creating RSA key pair...")
		crypto.CreateRSAPair(username)
	},
}

func init() {
	rootCmd.AddCommand(regCmd)
	regCmd.Flags().StringP("username", "u", "", "Your username to be used while chatting (required)")
	regCmd.MarkFlagRequired("username")
}
