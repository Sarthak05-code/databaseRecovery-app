package db

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

type MysqlClient struct{}

type MysqlBackupConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	DBname   string
}

// MysqlBackupResult represents result info for MySQL backups.
type MysqlBackupResult struct {
	Path     string
	Duration time.Duration
}

// ResolvePassword checks environment variables first, then prompts securely if empty
func ResolvePassword(explicitPass string) (string, error) {
	if explicitPass != "" {
		return explicitPass, nil
	}
	if envPass := os.Getenv("DB_PASSWORD"); envPass != "" {
		return envPass, nil
	}

	fmt.Print("Enter Database Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println() // Print clean line break after hidden input
	if err != nil {
		return "", fmt.Errorf("failed to read secure password: %w", err)
	}
	return strings.TrimSpace(string(bytePassword)), nil
}

func (m *MysqlClient) TestConnection() error {
	_, err := exec.LookPath("mysqldump")
	if err != nil {
		return fmt.Errorf("mysqldump utility binary not found in system PATH")
	}
	_, err = exec.LookPath("mysql")
	if err != nil {
		return fmt.Errorf("mysql client utility binary not found in system PATH")
	}
	return nil
}

// executeWithRetry wraps commands with basic transient fault handling
func executeWithRetry(attempts int, delay time.Duration, operation func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		if err = operation(); err == nil {
			return nil
		}
		if i < attempts-1 {
			time.Sleep(delay)
		}
	}
	return fmt.Errorf("operation failed after %d attempts: %w", attempts, err)
}

// Fixed function signature: changed BackupConfig to MysqlBackupConfig
func (m *MysqlClient) CreateBackup(config MysqlBackupConfig, targetDir string) (*MysqlBackupResult, error) {
	startTime := time.Now()
	outputFile := fmt.Sprintf("%s/%s_%s.sql", targetDir, config.DBname, time.Now().Format("20060102_150405"))

	args := []string{
		"-h" + config.Host,
		"-P" + config.Port,
		"-u" + config.Username,
		"-p" + config.Password,
		config.DBname,
		"--result-file=" + outputFile,
	}

	var errBuf bytes.Buffer
	err := executeWithRetry(3, 2*time.Second, func() error {
		errBuf.Reset()
		cmd := exec.Command("mysqldump", args...)
		cmd.Stderr = &errBuf
		return cmd.Run()
	})

	if err != nil {
		return nil, fmt.Errorf("mysqldump failed: %w - %s", err, errBuf.String())
	}

	return &MysqlBackupResult{
		Path:     outputFile,
		Duration: time.Since(startTime),
	}, nil
}

// Fixed function signature: changed BackupConfig to MysqlBackupConfig
func (m *MysqlClient) RestoreBackup(config MysqlBackupConfig, sourceFile string) error {
	// Auto-provision target schema if missing before restoring data
	provisionArgs := []string{
		"-h" + config.Host,
		"-P" + config.Port,
		"-u" + config.Username,
		"-p" + config.Password,
		"-e", fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s;", config.DBname),
	}

	var provErrBuf bytes.Buffer
	provCmd := exec.Command("mysql", provisionArgs...)
	provCmd.Stderr = &provErrBuf
	if err := provCmd.Run(); err != nil {
		return fmt.Errorf("failed to auto-provision target database: %w - %s", err, provErrBuf.String())
	}

	cmd := exec.Command("mysql", "-h"+config.Host, "-P"+config.Port, "-u"+config.Username, "-p"+config.Password, config.DBname)

	var errBuf bytes.Buffer
	err := executeWithRetry(3, 2*time.Second, func() error {
		errBuf.Reset()

		fileHandle, err := os.Open(sourceFile)
		if err != nil {
			return fmt.Errorf("unable to open file: %w", err)
		}
		defer fileHandle.Close()

		if strings.HasSuffix(sourceFile, ".gz") {
			gzReader, err := gzip.NewReader(fileHandle)
			if err != nil {
				return fmt.Errorf("failed to init gzip stream: %w", err)
			}
			defer gzReader.Close()
			cmd.Stdin = gzReader
		} else {
			cmd.Stdin = fileHandle
		}

		cmd.Stderr = &errBuf
		return cmd.Run()
	})

	if err != nil {
		return fmt.Errorf("restore failed: %w - %s", err, errBuf.String())
	}
	return nil
}