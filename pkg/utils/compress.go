package utils

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func GzipFile(sourcepath string) (string, error) {
	sourceFile, err := os.Open(sourcepath)
	if err != nil {
		return "", fmt.Errorf("Failed to open the source file : %w", err)
	}

	defer sourceFile.Close()

	targetPath := sourcepath + ".gz"
	targetFile, err := os.Create(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to create target compressed file! %w", err)
	}
	defer targetFile.Close()

	gzipWriter := gzip.NewWriter(targetFile)
	defer gzipWriter.Close()

	gzipWriter.Name = filepath.Base(sourcepath)

	_, err = io.Copy(gzipWriter, sourceFile)
	if err != nil {
		return "", fmt.Errorf("failed to compress data: %w", err)
	}

	gzipWriter.Close()
	sourceFile.Close()

	if err := os.Remove(sourcepath); err != nil {
		return targetPath, fmt.Errorf("backup compressed but failed to clean %w", err)
	}
	return targetPath, nil

}
