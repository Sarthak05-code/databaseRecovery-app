package cmd

import (
	"db-backup-cli/pkg/db"
	"db-backup-cli/pkg/storage"
	"db-backup-cli/pkg/utils"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var (
	host, port, user, password, dbname string
	storageDir                         string
	slackWebhook                       string
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Perform a full database backup with execution metrics",
	Run: func(cmd *cobra.Command, args []string) {
		startTime := time.Now()
		client := &db.MysqlClient{}
		config := db.BackupConfig{
			Host:     host,
			Port:     port,
			Username: user,
			Password: password,
			DBname:   dbname,
		}

		if err := client.TestConnection(); err != nil {
			utils.WriteLog("BACKUP", "FAILED", time.Since(startTime), err.Error())
			fmt.Println("Connection verification failed:", err)
			return
		}

		result, err := client.CreateBackup(config, ".")
		if err != nil {
			utils.WriteLog("BACKUP", "FAILED", time.Since(startTime), err.Error())
			fmt.Println("Dump execution failed:", err)
			return
		}

		compressedPath, err := utils.GzipFile(result.Path)
		if err != nil {
			utils.WriteLog("BACKUP", "FAILED", time.Since(startTime), err.Error())
			fmt.Println("Compression failed:", err)
			return
		}

		store := &storage.LocalStorage{TargetDir: storageDir}
		finalPath, err := store.Upload(compressedPath)
		if err != nil {
			utils.WriteLog("BACKUP", "FAILED", time.Since(startTime), err.Error())
			fmt.Println("Storage tracking error:", err)
			return
		}

		elapsed := time.Since(startTime)
		utils.WriteLog("BACKUP", "SUCCESS", elapsed, "")

		successMsg := fmt.Sprintf("Success! Backup archive safely stored at: %s (Time taken: %s)", finalPath, elapsed)
		fmt.Println(successMsg)

		// Optional Slack hook notification integration execution
		if slackWebhook != "" {
			utils.SendSlackNotification(slackWebhook, fmt.Sprintf("✅ Database Backup Job Complete for %s. Destination: %s", dbname, finalPath))
		}
	},
}

func init() {
	backupCmd.Flags().StringVarP(&host, "host", "H", "localhost", "Database host")
	backupCmd.Flags().StringVarP(&port, "port", "P", "3306", "Database port")
	backupCmd.Flags().StringVarP(&user, "user", "u", "", "Database user")
	backupCmd.Flags().StringVarP(&password, "pass", "p", "", "Database password")
	backupCmd.Flags().StringVarP(&dbname, "db", "d", "", "Database name")
	backupCmd.Flags().StringVarP(&storageDir, "output", "o", "./backups", "Local target backup directory")
	backupCmd.Flags().StringVarP(&slackWebhook, "slack", "s", "", "Slack App incoming Webhook URL link (optional)")

	rootCmd.AddCommand(backupCmd)
}
