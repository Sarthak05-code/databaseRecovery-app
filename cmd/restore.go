package cmd

import (
	"db-backup-cli/pkg/db"
	"fmt"
	"github.com/spf13/cobra"
)

var backupFilePath string

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore an archive directly into target schema",
	Run: func(cmd *cobra.Command, args []string) {
		resolvedPassword, err := db.ResolvePassword(password)
		if err != nil {
			fmt.Println("Authentication Setup Error:", err)
			return
		}

		client := &db.MysqlClient{}
		// Updated to use MysqlBackupConfig
		config := db.MysqlBackupConfig{
			Host:     host,
			Port:     port,
			Username: user,
			Password: resolvedPassword,
			DBname:   dbname,
		}

		fmt.Println("Attempting database restore...")
		if err := client.RestoreBackup(config, backupFilePath); err != nil {
			fmt.Println("Restore broken:", err)
			return
		}
		fmt.Println("Database restored successfully")
	},
}

func init() {
	restoreCmd.Flags().StringVarP(&host, "host", "H", "localhost", "Database host")
	restoreCmd.Flags().StringVarP(&port, "port", "P", "3306", "Database port")
	restoreCmd.Flags().StringVarP(&user, "user", "u", "", "Database user")
	restoreCmd.Flags().StringVarP(&password, "pass", "p", "", "Database password")
	restoreCmd.Flags().StringVarP(&dbname, "db", "d", "", "Database name")
	restoreCmd.Flags().StringVarP(&backupFilePath, "file", "f", "", "Target backup path (.sql.gz)")
	
	restoreCmd.MarkFlagRequired("file")
	rootCmd.AddCommand(restoreCmd)
}