package s3client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"

	appConfig "s3manager/config"
	"s3manager/internal/models"
	"s3manager/pkg/utils"
)

type Client struct {
	s3Client *s3.Client
	config   *appConfig.Config
}

func New(cfg *appConfig.Config) (*Client, error) {
	awsConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.StaticCredentialsProvider{
			Value: aws.Credentials{
				AccessKeyID:     cfg.AccessKey,
				SecretAccessKey: cfg.SecretKey,
			},
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	var s3Client *s3.Client
	if cfg.ApiURL != "" {
		s3Client = s3.NewFromConfig(awsConfig, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.ApiURL)
			o.UsePathStyle = true
		})
	} else {
		s3Client = s3.NewFromConfig(awsConfig)
	}

	return &Client{
		s3Client: s3Client,
		config:   cfg,
	}, nil
}

func (c *Client) GetBucketInfo(ctx context.Context) (*models.BucketInfo, error) {
	bucketName := c.config.BucketName

	locationResp, err := c.s3Client.GetBucketLocation(ctx, &s3.GetBucketLocationInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get bucket location: %w", err)
	}

	region := string(locationResp.LocationConstraint)
	if region == "" {
		region = c.config.Region // Use configured region as fallback
	}

	var objectCount int64
	var totalSize int64
	var lastModified time.Time

	paginator := s3.NewListObjectsV2Paginator(c.s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		objectCount += int64(len(page.Contents))
		for _, obj := range page.Contents {
			totalSize += *obj.Size
			if obj.LastModified != nil && obj.LastModified.After(lastModified) {
				lastModified = *obj.LastModified
			}
		}
	}

	bucketsResp, err := c.s3Client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("failed to list buckets: %w", err)
	}

	var creationDate time.Time
	for _, bucket := range bucketsResp.Buckets {
		if *bucket.Name == bucketName {
			creationDate = *bucket.CreationDate
			break
		}
	}

	return &models.BucketInfo{
		BucketName:     bucketName,
		Region:         region,
		CreationDate:   creationDate,
		ObjectCount:    objectCount,
		TotalSizeBytes: totalSize,
		TotalSizeHuman: utils.FormatBytes(totalSize),
		LastModified:   lastModified,
		APIEndpoint:    c.config.ApiURL,
	}, nil
}

func (c *Client) DeleteOldFiles(ctx context.Context, folder string, daysOld int) (*models.DeleteResult, error) {
	bucketName := c.config.BucketName
	cutoffDate := time.Now().AddDate(0, 0, -daysOld)

	prefix := folder
	if !strings.HasSuffix(prefix, "/") && prefix != "" {
		prefix += "/"
	}

	var toDelete []types.ObjectIdentifier
	var deletedFiles []string
	var totalSize int64

	paginator := s3.NewListObjectsV2Paginator(c.s3Client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", err)
		}

		for _, obj := range page.Contents {
			if obj.LastModified != nil && obj.LastModified.Before(cutoffDate) {
				toDelete = append(toDelete, types.ObjectIdentifier{
					Key: obj.Key,
				})
				deletedFiles = append(deletedFiles, *obj.Key)
				totalSize += *obj.Size
			}
		}
	}

	deletedCount := 0
	for i := 0; i < len(toDelete); i += 1000 {
		end := i + 1000
		if end > len(toDelete) {
			end = len(toDelete)
		}

		batch := toDelete[i:end]
		if len(batch) == 0 {
			continue
		}

		_, err := c.s3Client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &types.Delete{
				Objects: batch,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to delete objects batch: %w", err)
		}
		deletedCount += len(batch)
	}

	return &models.DeleteResult{
		BucketName:     bucketName,
		Folder:         folder,
		DaysOld:        daysOld,
		DeletedFiles:   deletedFiles,
		DeletedCount:   deletedCount,
		TotalSizeBytes: totalSize,
		TotalSizeHuman: utils.FormatBytes(totalSize),
		OperationTime:  utils.FormatTime(time.Now()),
		CutoffDate:     utils.FormatTime(cutoffDate),
	}, nil
}

func (c *Client) UploadFiles(ctx context.Context, paths []string, destinationPath string, shouldArchive bool) (*models.UploadResult, error) {
	startTime := time.Now()
	bucketName := c.config.BucketName

	if err := utils.ValidatePaths(paths); err != nil {
		return nil, fmt.Errorf("path validation failed: %w", err)
	}

	var uploadItems []models.UploadItem
	var totalSize int64
	var archivePath string
	var archiveCreated bool

	uploader := manager.NewUploader(c.s3Client)

	if shouldArchive {
		archivePath = filepath.Join(os.TempDir(), utils.GenerateArchiveName(paths, ".zip"))
		archiveInfo, err := utils.CreateArchive(paths, archivePath)
		if err != nil {
			return nil, fmt.Errorf("failed to create archive: %w", err)
		}

		archiveCreated = true
		totalSize = archiveInfo.CompressedSize

		remotePath := c.buildRemotePath(destinationPath, filepath.Base(archivePath))
		if err := c.uploadSingleFile(ctx, uploader, archivePath, remotePath); err != nil {
			utils.CleanupTempFile(archivePath)
			return nil, fmt.Errorf("failed to upload archive: %w", err)
		}

		uploadItems = append(uploadItems, models.UploadItem{
			LocalPath:  strings.Join(paths, ", "),
			RemotePath: remotePath,
			Size:       archiveInfo.CompressedSize,
			IsArchived: true,
		})

		defer utils.CleanupTempFile(archivePath)
	} else {
		for _, path := range paths {
			items, size, err := c.uploadPath(ctx, uploader, path, destinationPath)
			if err != nil {
				return nil, fmt.Errorf("failed to upload %s: %w", path, err)
			}
			uploadItems = append(uploadItems, items...)
			totalSize += size
		}
	}

	duration := time.Since(startTime)

	return &models.UploadResult{
		BucketName:      bucketName,
		DestinationPath: destinationPath,
		Items:           uploadItems,
		TotalFiles:      len(uploadItems),
		TotalSizeBytes:  totalSize,
		TotalSizeHuman:  utils.FormatBytes(totalSize),
		OperationTime:   utils.FormatTime(startTime),
		ArchiveCreated:  archiveCreated,
		ArchivePath:     archivePath,
		UploadDuration:  duration.String(),
	}, nil
}

func (c *Client) uploadPath(ctx context.Context, uploader *manager.Uploader, localPath, destinationPath string) ([]models.UploadItem, int64, error) {
	var items []models.UploadItem
	var totalSize int64

	fileInfo, err := os.Stat(localPath)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to stat %s: %w", localPath, err)
	}

	if fileInfo.IsDir() {
		err := filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				relPath, err := filepath.Rel(localPath, path)
				if err != nil {
					return err
				}

				remotePath := c.buildRemotePath(destinationPath, filepath.Join(filepath.Base(localPath), relPath))

				if err := c.uploadSingleFile(ctx, uploader, path, remotePath); err != nil {
					return err
				}

				items = append(items, models.UploadItem{
					LocalPath:  path,
					RemotePath: remotePath,
					Size:       info.Size(),
					IsArchived: false,
				})

				totalSize += info.Size()
			}
			return nil
		})

		if err != nil {
			return nil, 0, err
		}
	} else {
		remotePath := c.buildRemotePath(destinationPath, filepath.Base(localPath))

		if err := c.uploadSingleFile(ctx, uploader, localPath, remotePath); err != nil {
			return nil, 0, err
		}

		items = append(items, models.UploadItem{
			LocalPath:  localPath,
			RemotePath: remotePath,
			Size:       fileInfo.Size(),
			IsArchived: false,
		})

		totalSize = fileInfo.Size()
	}

	return items, totalSize, nil
}

func (c *Client) uploadSingleFile(ctx context.Context, uploader *manager.Uploader, localPath, remotePath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", localPath, err)
	}
	defer file.Close()

	contentType := c.detectContentType(localPath)

	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.config.BucketName),
		Key:         aws.String(remotePath),
		Body:        file,
		ContentType: aws.String(contentType),
	})

	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

func (c *Client) buildRemotePath(destinationPath, filename string) string {
	if destinationPath == "" {
		return filename
	}

	destinationPath = strings.TrimPrefix(destinationPath, "/")

	if !strings.HasSuffix(destinationPath, "/") {
		destinationPath += "/"
	}

	return destinationPath + filename
}

func (c *Client) detectContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))

	contentTypes := map[string]string{
		".txt":  "text/plain",
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".pdf":  "application/pdf",
		".zip":  "application/zip",
		".tar":  "application/x-tar",
		".gz":   "application/gzip",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".mp3":  "audio/mpeg",
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
	}

	if contentType, exists := contentTypes[ext]; exists {
		return contentType
	}

	return "application/octet-stream"
}
