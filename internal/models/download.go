package models

type DownloadItem struct {
	RemotePath   string `json:"remote_path"`
	LocalPath    string `json:"local_path"`
	Size         int64  `json:"size"`
	LastModified string `json:"last_modified"`
}

type DownloadResult struct {
	BucketName       string         `json:"bucket_name"`
	SourcePath       string         `json:"source_path"`
	Items            []DownloadItem `json:"items"`
	TotalFiles       int            `json:"total_files"`
	TotalSizeBytes   int64          `json:"total_size_bytes"`
	TotalSizeHuman   string         `json:"total_size_human"`
	OperationTime    string         `json:"operation_time"`
	DownloadDuration string         `json:"download_duration"`
}
