package fileutils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// CalculateCRC computes a simple checksum for data validation
func CalculateCRC(data []byte) uint32 {
	var crc uint32 = 0
	for _, b := range data {
		crc = crc*31 + uint32(b)
	}
	return crc
}

// EnsureDirectory ensures that the directory for the given file path exists
func EnsureDirectory(path string) error {
	dir := filepath.Dir(path)
	if dir == "." {
		return nil
	}

	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// Create directory with default permissions
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to stat directory %s: %w", dir, err)
	}

	return nil
}

// AtomicWriteFile writes data to a file atomically using a temporary file
func AtomicWriteFile(filename string, data []byte, perm os.FileMode) error {
	// Ensure directory exists
	if err := EnsureDirectory(filename); err != nil {
		return err
	}

	// Create temp file in same directory
	dir := filepath.Dir(filename)
	tempFile, err := os.CreateTemp(dir, filepath.Base(filename)+".tmp")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tempPath := tempFile.Name()

	// Clean up on any error
	defer func() {
		if err != nil {
			os.Remove(tempPath)
		}
	}()

	// Write data to temp file
	if _, err = tempFile.Write(data); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to write to temp file: %w", err)
	}

	// Sync to ensure data is written to disk
	if err = tempFile.Sync(); err != nil {
		tempFile.Close()
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	// Close the file before rename
	if err = tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Set permissions
	if err = os.Chmod(tempPath, perm); err != nil {
		return fmt.Errorf("failed to set file permissions: %w", err)
	}

	// Rename the temp file (atomic on most filesystems)
	if err = os.Rename(tempPath, filename); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}

// ReadFileWithLimit reads a file with a size limit
func ReadFileWithLimit(path string, maxSize int64) ([]byte, error) {
	// Open file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat file: %w", err)
	}

	// Check file size
	if maxSize > 0 && stat.Size() > maxSize {
		return nil, fmt.Errorf("file size %d exceeds maximum allowed size %d", stat.Size(), maxSize)
	}

	// Read file with limit
	data := make([]byte, stat.Size())
	if _, err := io.ReadFull(file, data); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}