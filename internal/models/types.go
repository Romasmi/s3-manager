package models

import "time"

type BucketInfo struct {
	BucketName     string    `json:"bucket_name"`
	Region         string    `json:"region"`
	CreationDate   time.Time `json:"creation_date"`
	ObjectCount    int64     `json:"object_count"`
	TotalSizeBytes int64     `json:"total_size_bytes"`
	TotalSizeHuman string    `json:"total_size_human"`
	LastModified   time.Time `json:"last_modified"`
	APIEndpoint    string    `json:"api_endpoint,omitempty"`
}

type ErrorResponse struct {
	Error     string `json:"error"`
	Timestamp string `json:"timestamp"`
	Command   string `json:"command"`
}

type DeleteResult struct {
	BucketName     string   `json:"bucket_name"`
	Folder         string   `json:"folder"`
	DaysOld        int      `json:"days_old"`
	DeletedFiles   []string `json:"deleted_files"`
	DeletedCount   int      `json:"deleted_count"`
	TotalSizeBytes int64    `json:"total_size_bytes"`
	TotalSizeHuman string   `json:"total_size_human"`
	OperationTime  string   `json:"operation_time"`
	CutoffDate     string   `json:"cutoff_date"`
}
