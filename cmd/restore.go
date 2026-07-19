package cmd

import (
	"db-backup-cli/pkg/db"
	"fmt"

	"github.com/spf13/cobra"
)

var (
	restoreFile string
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a databse from an uncompressed backup file",
	Run: func(cmd *cobra.Command, args []string) {
		client := &db.MysqlClient{}
		config := db.BackupConfig{
			Host:     host,
			Port:     port,
			Username: user,
			Password: password,
			DBname:   dbname,
		}

		fmt.Println("Attempting database restore...")
		err := client.RestoreBackup(config, restoreFile)
		if err != nil {
			fmt.Printf("Restore failed : %v", err)
			return
		}
		fmt.Println("Database restored sucessfully")
	},
}

func init() {
	restoreCmd.Flags().StringVarP(&host, "host", "H", "localhost", "Database host")
	restoreCmd.Flags().StringVarP(&port, "port", "P", "3306", "Database port")
	restoreCmd.Flags().StringVarP(&user, "user", "u", "", "Database user")
	restoreCmd.Flags().StringVarP(&password, "pass", "p", "", "Database password")
	restoreCmd.Flags().StringVarP(&dbname, "db", "d", "", "Database name")
	restoreCmd.Flags().StringVarP(&restoreFile, "file", "f", "", "Path to the SQL file to restore (Required)")
	restoreCmd.MarkFlagRequired("file")

	rootCmd.AddCommand(restoreCmd)
}
