package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// LocalStorage handles moving backup assets to their final destination directory
type LocalStorage struct {
	TargetDir string
}

// Upload moves the temporary compressed file into the designated backup directory
func (s *LocalStorage) Upload(tempFilePath string) (string, error) {
	// Ensure the output directory exists
	if err := os.MkdirAll(s.TargetDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create target storage directory: %w", err)
	}

	// Determine the final destination path
	fileName := filepath.Base(tempFilePath)
	finalPath := filepath.Join(s.TargetDir, fileName)

	// Open the source temporary file
	srcFile, err := os.Open(tempFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open source temp file: %w", err)
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.OpenFile(finalPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create destination backup file: %w", err)
	}
	defer dstFile.Close()

	// Stream the data across
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return "", fmt.Errorf("failed streaming data to storage destination: %w", err)
	}

	return finalPath, nil
}

// RetentionPolicy defines the configuration window rules for pruning operations
type RetentionPolicy struct {
	MaxDays   int
	KeepCount int
}

// BackupFile represents an identified historical backup target archive
type BackupFile struct {
	Path      string
	Timestamp time.Time
}

// PruneOldBackups sweeps the output folder layout and drops assets outside retention boundaries
func PruneOldBackups(outputDir string, dbName string, policy RetentionPolicy) (int, error) {
	if policy.MaxDays <= 0 && policy.KeepCount <= 0 {
		return 0, nil // No structural policy targets explicitly declared
	}

	// Dynamic match layout for matching: <database>_YYYYMMDD_HHMMSS.sql.gz
	pattern := fmt.Sprintf(`^%s_\d{8}_\d{6}\.sql\.gz$`, regexp.QuoteMeta(dbName))
	re, err := regexp.Compile(pattern)
	if err != nil {
		return 0, fmt.Errorf("failed compiling retention filter matcher: %w", err)
	}

	files, err := os.ReadDir(outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("unable to read backup directory for pruning: %w", err)
	}

	var identifiedBackups []BackupFile

	// Scan filesystem layouts matching target structures
	for _, entry := range files {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !re.MatchString(name) {
			continue
		}

		// Extract date token slices out of raw text blocks
		// Example: shop_db_20260719_183501.sql.gz -> "20260719_183501"
		parts := strings.Split(strings.TrimSuffix(name, ".sql.gz"), "_")
		if len(parts) < 3 {
			continue
		}

		timeStr := fmt.Sprintf("%s_%s", parts[len(parts)-2], parts[len(parts)-1])
		timestamp, err := time.Parse("20060102_150405", timeStr)
		if err != nil {
			continue // Skip un-parseable variants gracefully
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		path := filepath.Join(outputDir, name)
		identifiedBackups = append(identifiedBackups, BackupFile{
			Path:      path,
			Timestamp: info.ModTime(),
		})

		identifiedBackups[len(identifiedBackups)-1].Timestamp = timestamp
	}

	// Sort backups in descending order (newest first)
	for i := 0; i < len(identifiedBackups); i++ {
		for j := i + 1; j < len(identifiedBackups); j++ {
			if identifiedBackups[i].Timestamp.Before(identifiedBackups[j].Timestamp) {
				identifiedBackups[i], identifiedBackups[j] = identifiedBackups[j], identifiedBackups[i]
			}
		}
	}

	prunedCounter := 0
	now := time.Now()

	for idx, backup := range identifiedBackups {
		shouldDelete := false

		// Rule A: Evaluate Count Cap Limits
		if policy.KeepCount > 0 && idx >= policy.KeepCount {
			shouldDelete = true
		}

		// Rule B: Evaluate Calendar Age Constraints
		if policy.MaxDays > 0 && !shouldDelete {
			cutoff := now.AddDate(0, 0, -policy.MaxDays)
			if backup.Timestamp.Before(cutoff) {
				shouldDelete = true
			}
		}

		if shouldDelete {
			if err := os.Remove(backup.Path); err != nil {
				return prunedCounter, fmt.Errorf("failed clearing backup path asset %s: %w", backup.Path, err)
			}
			prunedCounter++
		}
	}

	return prunedCounter, nil
}
