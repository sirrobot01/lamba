package common

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func ExtractZip(zipFile io.ReaderAt, size int64, destPath string) error {
	reader, err := zip.NewReader(zipFile, size)
	if err != nil {
		return fmt.Errorf("failed to create zip reader: %w", err)
	}

	// Create destination directory if it doesn't exist
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	for _, file := range reader.File {

		if strings.HasPrefix(file.Name, "__MACOSX") || strings.HasPrefix(filepath.Base(file.Name), ".") {
			continue
		}
		// Construct destination path
		path := filepath.Join(destPath, file.Name)

		// Check for ZipSlip vulnerability
		if !strings.HasPrefix(path, filepath.Clean(destPath)+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if file.FileInfo().IsDir() {
			os.MkdirAll(path, file.Mode())
			continue
		}

		// Create directory for file if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}

		// Create file
		dstFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file: %w", err)
		}

		// Open file in zip
		srcFile, err := file.Open()
		if err != nil {
			dstFile.Close()
			return fmt.Errorf("failed to open file in zip: %w", err)
		}

		// Copy file contents
		_, err = io.Copy(dstFile, srcFile)
		srcFile.Close()
		dstFile.Close()
		if err != nil {
			return fmt.Errorf("failed to copy file contents: %w", err)
		}
	}

	return nil
}
