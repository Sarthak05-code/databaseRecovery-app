package db

import (
	"time"
)

// Struct that holds the connection details.
type BackupConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBname   string
}

// Struct that tracks the performance and status
type BackupResult struct {
	Path       string
	Duration   time.Duration
	SizeInByte int64
}

// interface Client thatr defines the common capabilities for any database engine we support
type Client interface {
	TestConnection() error
	CreateBack(config BackupConfig, targetDir string) (*BackupResult, error)
	RestoreBackup(config BackupConfig, sourceFile string) error
	SelectiveRestore(config BackupConfig, sourceFile string, tables []string) error
}
