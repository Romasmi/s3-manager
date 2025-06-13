package models

import "time"

type UploadItem struct {
	LocalPath  string `json:"local_path"`
	RemotePath string `json:"remote_path"`
	Size       int64  `json:"size"`
	IsArchived bool   `json:"is_archived"`
}

type UploadResult struct {
	BucketName      string       `json:"bucket_name"`
	DestinationPath string       `json:"destination_path"`
	Items           []UploadItem `json:"items"`
	TotalFiles      int          `json:"total_files"`
	TotalSizeBytes  int64        `json:"total_size_bytes"`
	TotalSizeHuman  string       `json:"total_size_human"`
	OperationTime   string       `json:"operation_time"`
	ArchiveCreated  bool         `json:"archive_created"`
	ArchivePath     string       `json:"archive_path,omitempty"`
	UploadDuration  string       `json:"upload_duration"`
}

type ArchiveInfo struct {
	ArchivePath      string    `json:"archive_path"`
	OriginalPaths    []string  `json:"original_paths"`
	CompressedSize   int64     `json:"compressed_size"`
	OriginalSize     int64     `json:"original_size"`
	CompressionRatio float64   `json:"compression_ratio"`
	CreatedAt        time.Time `json:"created_at"`
}
