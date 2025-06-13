package utils

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"s3manager/internal/models"
	"strings"
	"time"
)

func CreateArchive(paths []string, outputPath string) (*models.ArchiveInfo, error) {
	outFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create archive file: %w", err)
	}
	defer outFile.Close()

	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	var originalSize int64
	createdAt := time.Now()

	for _, path := range paths {
		if err := addToArchive(zipWriter, path, ""); err != nil {
			return nil, fmt.Errorf("failed to add %s to archive: %w", path, err)
		}

		size, err := getPathSize(path)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate size for %s: %w", path, err)
		}
		originalSize += size
	}

	if err := zipWriter.Close(); err != nil {
		return nil, fmt.Errorf("failed to finalize archive: %w", err)
	}

	fileInfo, err := outFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get archive info: %w", err)
	}
	compressedSize := fileInfo.Size()

	compressionRatio := 0.0
	if originalSize > 0 {
		compressionRatio = float64(compressedSize) / float64(originalSize)
	}

	return &models.ArchiveInfo{
		ArchivePath:      outputPath,
		OriginalPaths:    paths,
		CompressedSize:   compressedSize,
		OriginalSize:     originalSize,
		CompressionRatio: compressionRatio,
		CreatedAt:        createdAt,
	}, nil
}

func addToArchive(zipWriter *zip.Writer, sourcePath, basePath string) error {
	return filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		if basePath != "" {
			header.Name = filepath.Join(basePath, strings.TrimPrefix(path, sourcePath))
		} else {
			if sourcePath == path {
				header.Name = filepath.Base(path)
			} else {
				relPath, err := filepath.Rel(filepath.Dir(sourcePath), path)
				if err != nil {
					return err
				}
				header.Name = relPath
			}
		}

		header.Name = filepath.ToSlash(header.Name)

		header.Method = zip.Deflate

		if info.IsDir() {
			return nil
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}

func getPathSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

func GenerateArchiveName(paths []string, extension string) string {
	if len(paths) == 1 {
		baseName := filepath.Base(paths[0])
		if ext := filepath.Ext(baseName); ext != "" {
			baseName = strings.TrimSuffix(baseName, ext)
		}
		return fmt.Sprintf("%s_%s%s", baseName, time.Now().Format("20060102_150405"), extension)
	}

	return fmt.Sprintf("archive_%s%s", time.Now().Format("20060102_150405"), extension)
}

func ValidatePaths(paths []string) error {
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("path does not exist: %s", path)
			}
			return fmt.Errorf("cannot access path %s: %w", path, err)
		}
	}
	return nil
}

func CleanupTempFile(path string) error {
	if path == "" {
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to cleanup temporary file %s: %w", path, err)
	}
	return nil
}
