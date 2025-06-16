package utils

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePaths(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archieve-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	tempFile := filepath.Join(tempDir, "test-file.txt")
	if err := os.WriteFile(tempFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	tests := []struct {
		name        string
		paths       []string
		expectError bool
	}{
		{"Valid file", []string{tempFile}, false},
		{"Valid directory", []string{tempDir}, false},
		{"Multiple valid paths", []string{tempFile, tempDir}, false},
		{"Non-existent path", []string{filepath.Join(tempDir, "non-existent")}, true},
		{"Empty paths", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePaths(tt.paths)
			if (err != nil) != tt.expectError {
				t.Errorf("ValidatePaths() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestGenerateArchiveName(t *testing.T) {
	tests := []struct {
		name      string
		paths     []string
		extension string
		check     func(string) bool
	}{
		{
			name:      "Single file",
			paths:     []string{"/path/to/file.txt"},
			extension: ".zip",
			check: func(result string) bool {
				return strings.HasPrefix(result, "file_") && strings.HasSuffix(result, ".zip")
			},
		},
		{
			name:      "Multiple files",
			paths:     []string{"/path/to/file1.txt", "/path/to/file2.txt"},
			extension: ".tar",
			check: func(result string) bool {
				return strings.HasPrefix(result, "archive_") && strings.HasSuffix(result, ".tar")
			},
		},
		{
			name:      "Directory",
			paths:     []string{"/path/to/dir"},
			extension: ".zip",
			check: func(result string) bool {
				return strings.HasPrefix(result, "dir_") && strings.HasSuffix(result, ".zip")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateArchiveName(tt.paths, tt.extension)
			if !tt.check(result) {
				t.Errorf("GenerateArchiveName() = %s, doesn't match expected pattern", result)
			}
		})
	}
}

func TestCleanupTempFile(t *testing.T) {
	tempFile, err := os.CreateTemp("", "cleanup-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tempFile.Close()
	tempPath := tempFile.Name()

	err = CleanupTempFile(tempPath)
	if err != nil {
		t.Errorf("CleanupTempFile() error = %v", err)
	}

	_, err = os.Stat(tempPath)
	if !os.IsNotExist(err) {
		t.Errorf("File was not removed: %v", err)
	}

	err = CleanupTempFile(tempPath)
	if err != nil {
		t.Errorf("CleanupTempFile() on non-existent file error = %v", err)
	}

	err = CleanupTempFile("")
	if err != nil {
		t.Errorf("CleanupTempFile() with empty path error = %v", err)
	}
}

func TestCreateArchive(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "archive-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	file1Path := filepath.Join(tempDir, "file1.txt")
	file2Path := filepath.Join(tempDir, "file2.txt")

	if err := os.WriteFile(file1Path, []byte("test content 1"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2Path, []byte("test content 2"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	subDir := filepath.Join(tempDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	file3Path := filepath.Join(subDir, "file3.txt")
	if err := os.WriteFile(file3Path, []byte("test content 3"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	archivePath := filepath.Join(tempDir, "test-archive.zip")

	archiveInfo, err := CreateArchive([]string{file1Path, file2Path}, archivePath)
	if err != nil {
		t.Fatalf("CreateArchive() error = %v", err)
	}

	if _, err := os.Stat(archivePath); err != nil {
		t.Errorf("Archive file was not created: %v", err)
	}

	if archiveInfo.ArchivePath != archivePath {
		t.Errorf("ArchivePath = %s, want %s", archiveInfo.ArchivePath, archivePath)
	}

	if len(archiveInfo.OriginalPaths) != 2 {
		t.Errorf("OriginalPaths length = %d, want 2", len(archiveInfo.OriginalPaths))
	}

	if archiveInfo.CompressedSize <= 0 {
		t.Errorf("CompressedSize = %d, want > 0", archiveInfo.CompressedSize)
	}

	if archiveInfo.OriginalSize <= 0 {
		t.Errorf("OriginalSize = %d, want > 0", archiveInfo.OriginalSize)
	}

	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		t.Fatalf("Failed to open archive: %v", err)
	}
	defer reader.Close()

	if len(reader.File) != 2 {
		t.Errorf("Archive contains %d files, want 2", len(reader.File))
	}

	archivePath2 := filepath.Join(tempDir, "test-archive2.zip")
	_, err = CreateArchive([]string{tempDir}, archivePath2)
	if err != nil {
		t.Fatalf("CreateArchive() with directory error = %v", err)
	}

	if _, err := os.Stat(archivePath2); err != nil {
		t.Errorf("Archive file was not created: %v", err)
	}

	reader2, err := zip.OpenReader(archivePath2)
	if err != nil {
		t.Fatalf("Failed to open archive: %v", err)
	}
	defer reader2.Close()

	if len(reader2.File) < 3 {
		t.Errorf("Archive contains %d files, want at least 3", len(reader2.File))
	}

	_, err = CreateArchive([]string{filepath.Join(tempDir, "non-existent")}, archivePath)
	if err == nil {
		t.Errorf("CreateArchive() with invalid path should return error")
	}
}

func TestGetPathSize(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "pathsize-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	file1Path := filepath.Join(tempDir, "file1.txt")
	file2Path := filepath.Join(tempDir, "file2.txt")

	file1Content := []byte("test content 1")
	file2Content := []byte("test content 2 with more data")

	if err := os.WriteFile(file1Path, file1Content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2Path, file2Content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	size1, err := getPathSize(file1Path)
	if err != nil {
		t.Errorf("getPathSize() error = %v", err)
	}
	if size1 != int64(len(file1Content)) {
		t.Errorf("getPathSize() = %d, want %d", size1, len(file1Content))
	}

	expectedSize := int64(len(file1Content) + len(file2Content))
	size2, err := getPathSize(tempDir)
	if err != nil {
		t.Errorf("getPathSize() error = %v", err)
	}
	if size2 != expectedSize {
		t.Errorf("getPathSize() = %d, want %d", size2, expectedSize)
	}

	_, err = getPathSize(filepath.Join(tempDir, "non-existent"))
	if err == nil {
		t.Errorf("getPathSize() with invalid path should return error")
	}
}
