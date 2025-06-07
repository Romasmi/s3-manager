package models

import "time"

type ArchiveInfo struct {
	ArchivePath      string    `json:"archive_path"`
	OriginalPaths    []string  `json:"original_paths"`
	CompressedSize   int64     `json:"compressed_size"`
	OriginalSize     int64     `json:"original_size"`
	CompressionRatio float64   `json:"compression_ratio"`
	CreatedAt        time.Time `json:"created_at"`
}
