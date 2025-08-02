package cmd

import (
	"gochat/chat"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/andybrewer/mack"
)

var (
	detach bool
)

var deamonCmd = &cobra.Command{
	Use:   "deamon",
	Short: "Starts deamon process to connect to chatserver for chat client",
	//Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username, err := cmd.Flags().GetString("username")
		if err != nil {
			log.Fatal(err)
		}

		if detach {
			relaunchInBackground(username)
		}

		log.Printf("Starting deamon for %s\n", &username)

		pid := os.Getpid()
		pidFile := filepath.Join(getUserDir(username), "deamon.pid")
		os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)

		chat.StartClient(username)
	},
}

func init() {
	rootCmd.AddCommand(deamonCmd)
	deamonCmd.Flags().StringP("username", "u", "", "Your username to be used while chatting (required)")
	//daemonCmd.Flags().StringVarP(&user, "user", "u", "", "Username to authenticate")
	deamonCmd.MarkFlagRequired("username")
	deamonCmd.Flags().BoolVar(&detach, "detach", false, "Run in background")
}

func relaunchInBackground(user string) {
	exe, err := os.Executable()
	if err != nil {
		log.Fatal("Error getting executable path", err)
	}

	args := []string{"deamon", "-u", user}
	cmd := exec.Command(exe, args...)

	if runtime.GOOS != "windows" {
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setsid: true,
		}
	} else {
		//cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	}

	if err := cmd.Start(); err != nil {
		log.Fatal("Error starting background deamon:", err)
	}

	mack.Notify("gochat deamon process started", "gochat")

	os.Exit(0)
}

func getUserDir(user string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Error getting user home directory:", err)
	}
	return filepath.Join(homeDir, ".gochat", user)
}
