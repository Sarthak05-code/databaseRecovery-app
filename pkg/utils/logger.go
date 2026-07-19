package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type LogEvent struct {
	Timestamp string `json:"timestamp"`
	Action    string `json:"action"` // "BACKUP" or "RESTORE"
	Status    string `json:"status"` // "SUCCESS" or "FAILED"
	Duration  string `json:"duration"`
	Error     string `json:"error,omitempty"`
}

func WriteLog(action, status string, duraion time.Duration, err string) {
	logFile, _ := os.OpenFile("backup_activity.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	defer logFile.Close()

	event := LogEvent{
		Timestamp: time.Now().Format(time.RFC3339),
		Action:    action,
		Status:    status,
		Duration:  duraion.String(),
		Error:     err,
	}

	jsonData, _ := json.Marshal(event)
	fmt.Fprintln(logFile, string(jsonData))
}
