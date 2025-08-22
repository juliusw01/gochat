package cmd

import (
	"gochat/chat"
	"gochat/misc"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	daemonized bool
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Starts daemon process to connect to chatserver for chat client",
	Run: func(cmd *cobra.Command, args []string) {
		username, err := cmd.Flags().GetString("username")
		if err != nil {
			log.Fatal(err)
		}

		// If not already daemonized, relaunch and exit
		if !daemonized {
			relaunchInBackground(username)
			return
		}

		// Only the daemonized process reaches here
		log.Printf("Starting daemon for %s\n", username)

		pid := os.Getpid()
		pidFile := filepath.Join(getUserDir(username), "daemon.pid")
		if err := os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644); err != nil {
			log.Printf("Error writing pid file: %v", err)
		}

		chat.StartClient(username, true)
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
	daemonCmd.Flags().StringP("username", "u", "", "Your username to be used while chatting (required)")
	daemonCmd.MarkFlagRequired("username")

	// Hidden flag, only used internally
	daemonCmd.Flags().BoolVar(&daemonized, "daemonized", false, "Internal flag (do not use directly)")
	err := daemonCmd.Flags().MarkHidden("daemonized")
	if err != nil {
		log.Println(err)
	}
}

func relaunchInBackground(user string) {
	exe, err := os.Executable()
	if err != nil {
		log.Fatal("Error getting executable path", err)
	}

	args := []string{"daemon", "-u", user, "--daemonized"}
	cmd := exec.Command(exe, args...)

	misc.SetBackgroundAttributes(cmd)

	if err := cmd.Start(); err != nil {
		log.Fatal("Error starting background daemon:", err)
	}

	misc.Notify("daemon process started", "gochat", "", "Blow.aiff")

	os.Exit(0)
}

func getUserDir(user string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error getting user home directory:", err)
	}
	return filepath.Join(homeDir, ".gochat", user)
}
