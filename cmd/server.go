package cmd

import (
	"db-backup-cli/pkg/db"
	"db-backup-cli/pkg/storage"
	"db-backup-cli/pkg/utils"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

type BackupRequest struct {
	Host         string `json:"host"`
	Port         string `json:"port"`
	User         string `json:"user"`
	Password     string `json:"password"`
	DBName       string `json:"dbname"`
	StorageDir   string `json:"storageDir"`
	SlackWebhook string `json:"slackWebhook"`
}

type RestoreRequest struct {
	Host           string `json:"host"`
	Port           string `json:"port"`
	User           string `json:"user"`
	Password       string `json:"password"`
	DBName         string `json:"dbname"`
	BackupFilePath string `json:"backupFilePath"`
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the full-stack web dashboard UI",
	Run: func(cmd *cobra.Command, args []string) {
		fs := http.FileServer(http.Dir("./web"))
		http.Handle("/", fs)

		http.HandleFunc("/api/backup", handleBackupApi)
		http.HandleFunc("/api/restore", handleRestoreApi)

		fmt.Println("Dashboard UI running at http://localhost:8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
}

func handleBackupApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req BackupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	startTime := time.Now()
	lockFile := filepath.Join(os.TempDir(), "dbbackup_operation.lock")
	file, err := os.OpenFile(lockFile, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0666)
	if err != nil {
		http.Error(w, "Concurrency Conflict: Another backup operation is active.", http.StatusConflict)
		return
	}

	defer func() {
		file.Close()
		os.Remove(lockFile)
	}()

	client := &db.MysqlClient{}
	config := db.MysqlBackupConfig{
		Host:     req.Host,
		Port:     req.Port,
		Username: req.User,
		Password: req.Password,
		DBname:   req.DBName,
	}

	if err := client.TestConnection(); err != nil {
		// Use startTime to log the duration of the failed connection check
		utils.WriteLog("BACKUP", "FAILED", time.Since(startTime), err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result, err := client.CreateBackup(config, ".")
	if err != nil {
		utils.WriteLog("BACKUP", "FAILED", time.Since(startTime), err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	compressedPath, err := utils.GzipFile(result.Path)
	if err != nil {
		utils.WriteLog("BACKUP", "FAILED", time.Since(startTime), err.Error())
		http.Error(w, "Compression failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	store := &storage.LocalStorage{TargetDir: req.StorageDir}
	finalPath, err := store.Upload(compressedPath)
	if err != nil {
		utils.WriteLog("BACKUP", "FAILED", time.Since(startTime), err.Error())
		http.Error(w, "Storage transfer failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Calculate final elapsed time
	elapsed := time.Since(startTime)
	utils.WriteLog("BACKUP", "SUCCESS", elapsed, "")

	// Trigger slack message alert asynchronously if a webhook URL was passed
	if req.SlackWebhook != "" {
		_ = utils.SendSlackNotification(req.SlackWebhook, fmt.Sprintf("✅ Web UI Database Backup Complete for %s. Target: %s", req.DBName, finalPath))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Backup stored at: " + finalPath,
		"elapsed": elapsed.String(), // Send the execution timeframe back to the UI dashboard!
	})
}

func handleRestoreApi(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.BackupFilePath == "" {
		http.Error(w, "Missing target backup file path", http.StatusBadRequest)
		return
	}

	startTime := time.Now()

	client := &db.MysqlClient{}
	config := db.MysqlBackupConfig{
		Host:     req.Host,
		Port:     req.Port,
		Username: req.User,
		Password: req.Password,
		DBname:   req.DBName,
	}

	// Test the system path binaries first
	if err := client.TestConnection(); err != nil {
		utils.WriteLog("RESTORE", "FAILED", time.Since(startTime), err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Trigger the restoration engine logic
	if err := client.RestoreBackup(config, req.BackupFilePath); err != nil {
		utils.WriteLog("RESTORE", "FAILED", time.Since(startTime), err.Error())
		http.Error(w, "Restore broken: "+err.Error(), http.StatusInternalServerError)
		return
	}

	elapsed := time.Since(startTime)
	utils.WriteLog("RESTORE", "SUCCESS", elapsed, "")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Database archive successfully restored to schema: " + req.DBName,
	})
}
