package cmd

import (
	"db-backup-cli/pkg/db"
	"db-backup-cli/pkg/storage"
	"db-backup-cli/pkg/utils"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
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
	Short: "Perform a secure full database backup with execution metrics",
	Run: func(cmd *cobra.Command, args []string) {
		startTime := time.Now()

		// 1. Establish runtime file lock
		lockFile := filepath.Join(os.TempDir(), "dbbackup_operation.lock")
		file, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
		if err != nil {
			fmt.Println("Concurrency Conflict: Another instance of this backup tool is actively executing.")
			return
		}

		// Helper cleanup function
		cleanupLock := func() {
			file.Close()
			os.Remove(lockFile)
		}
		defer cleanupLock()

		// 2. Interrupt Interception: Catch Ctrl+C and clean up the lock file immediately
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-sigChan
			fmt.Println("\n\nExecution interrupted by user. Cleaning up background file locks...")
			cleanupLock()
			os.Exit(1)
		}()

		// 3. Security: Securely resolve database credential strings
		resolvedPassword, err := db.ResolvePassword(password)
		if err != nil {
			fmt.Println("Authentication Setup Error:", err)
			return
		}

		client := &db.MysqlClient{}
		config := db.MysqlBackupConfig{
			Host:     host,
			Port:     port,
			Username: user,
			Password: resolvedPassword,
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
		_ = os.Remove(result.Path)

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