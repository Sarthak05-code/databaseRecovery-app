package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)


var rootCmd = &cobra.Command {
	Use: "dbBackup",
	Short: "dbBackup is a versatile database backup and restore cli tool",
	Long: "A robust cli utility designed for database nackup and restoration",
}

func Execute() {
	if err := rootCmd.Execute() ; err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}