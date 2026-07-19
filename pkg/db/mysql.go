package db

import (
	"bytes"
	"compress/gzip" // Added this to unzip files on the fly
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type MysqlClient struct{}

func (m *MysqlClient) TestConnection() error {
	_, err := exec.LookPath("mysqldump")
	if err != nil {
		return fmt.Errorf("mysqldump not found: Ensure a mysql is indeed installed")
	}
	return nil
}

func (m *MysqlClient) CreateBackup(config BackupConfig, targetDir string) (*BackupResult, error) {
	startTime := time.Now()
	outputFile := fmt.Sprintf("%s/%s_%s.sql", targetDir, config.DBname, time.Now().Format("20060102_150405"))

	args := []string{
		"-h" + config.Host,
		"-P" + config.Port, // Uses uppercase -P
		"-u" + config.Username,
		"-p" + config.Password,
		config.DBname,
		"--result-file=" + outputFile,
	}

	cmd := exec.Command("mysqldump", args...)

	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("mysqldump failed: %w - %s", err, errBuf.String())
	}

	return &BackupResult{
		Path:     outputFile,
		Duration: time.Since(startTime),
	}, nil
}

func (m *MysqlClient) RestoreBackup(config BackupConfig, sourceFile string) error {
	_, err := exec.LookPath("mysql")
	if err != nil {
		return fmt.Errorf("mysql client utility not found")
	}

	cmd := exec.Command("mysql", "-h"+config.Host, "-P"+config.Port, "-u"+config.Username, "-p"+config.Password, config.DBname)

	fileHandle, err := os.Open(sourceFile)
	if err != nil {
		return fmt.Errorf("unable to read backup file: %w", err)
	}
	defer fileHandle.Close()

	var errBuf bytes.Buffer
	cmd.Stderr = &errBuf

	// Check if the file is a compressed .gz file
	if strings.HasSuffix(sourceFile, ".gz") {
		gzReader, err := gzip.NewReader(fileHandle)
		if err != nil {
			return fmt.Errorf("failed to initialize gzip reader: %w", err)
		}
		defer gzReader.Close()

		// Feed the uncompressed text stream into the mysql terminal
		cmd.Stdin = gzReader
	} else {
		// Fallback for regular uncompressed text .sql files
		cmd.Stdin = fileHandle
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("restore failed: %w - %s", err, errBuf.String())
	}
	return nil
}

func (m *MysqlClient) SelectiveRestore(config BackupConfig, sourceFile string, tables []string) error {
	return fmt.Errorf("Selective target restore under construction")
}
