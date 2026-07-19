package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Uploader interface {
	Upload(localFilePath string) (string, error)
}

type LocalStorage struct {
	TargetDir string
}

func (l *LocalStorage) Upload(localFilePath string) (string, error) {
	if err := os.MkdirAll(l.TargetDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create local storage directory: %w", err)
	}

	fileName := filepath.Base(localFilePath)
	destinationPath := filepath.Join(l.TargetDir, fileName)

	sourceFile, err := os.Open(localFilePath)
	if err != nil {
		return "", err
	}

	defer sourceFile.Close()
	destFile, err := os.Create(destinationPath)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return "", fmt.Errorf("failed to copy file to local storage : %w", err)
	}
	sourceFile.Close()
	os.Remove(localFilePath)
	return destinationPath, nil
}
