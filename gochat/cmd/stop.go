package cmd

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/spf13/cobra"
)

var stopUser string

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop background deamon for specific user",
	//Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		username, err := cmd.Flags().GetString("username")
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Stopping deamon for %s\n", &username)

		pidFile := filepath.Join(getUserDir(username), "deamon.pid")

		data, err := os.ReadFile(pidFile)
		if err != nil {
			log.Fatal("Error reading pidfile:", err)
		}
		pid, err := strconv.Atoi(string(data))
		if err != nil {
			log.Fatal("Invalid PID:", err)
		}

		if err := syscall.Kill(pid, syscall.SIGTERM); err != nil {
			log.Fatal("Error killing deamon process:", err)
		}

		os.Remove(pidFile)
		log.Println("Stopped deamon process for ", username)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().StringP("username", "u", "", "Your username to be used while chatting (required)")
	stopCmd.MarkFlagRequired("username")
}