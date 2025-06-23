# S3 Manager CLI Tool

A Go-based CLI tool for managing S3-compatible storage.

## Features

- 📊 **Bucket Information**: Get detailed bucket statistics including object count, total size, and metadata
- 🗂️ **File Cleanup**: Delete files older than specified days with folder-specific targeting
- 📤 **File Upload**: Upload files and folders with automatic archiving options
- 🔧 **Flexible Configuration**: Support for custom S3 endpoints (MinIO, DigitalOcean Spaces, etc.)
- 🛡️ **Safety Features**: Confirmation prompts and dry-run mode for delete operations
- ⚡ **Performance**: Efficient batch operations for large buckets

## Installation

### Prerequisites

- Go 1.21 or later
- S3 credentials 

### Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd s3cli

# Install dependencies
go mod tidy

# Build the binary
go build -o s3manager

# Optional: Install globally
go install
```

### Build for Linux x64
``` bash
GOOS=linux GOARCH=amd64 go build -o s3manager
```

## Configuration

### Environment Variables

Create a `.env` file in the project root (see `.env.example`):

```bash
# Copy the example file
cp .env.example .env

# Edit with your configuration
nano .env
```

### Required Configuration

| Variable      | Description            | Example     |
|---------------|------------------------|-------------|
| `ACCESS_KEY`  | AWS Access Key ID      | `AKIA...`   |
| `SECRET_KEY`  | AWS Secret Access Key  | `wJalr...`  |
| `BUCKET_NAME` | Default S3 bucket name | `my-bucket` |
| `REGION`      | AWS region             | `us-east-1` |

### Optional Configuration

| Variable  | Description          | Example                 |
|-----------|----------------------|-------------------------|
| `API_URL` | Custom S3 endpoint   | `http://localhost:9000` |
| `TOKEN`   | Authentication token | `token123`              |

## Usage

### Get Bucket Information

Retrieve comprehensive bucket statistics:

```bash
# Use configured bucket
./s3manager bucket-info

# Override bucket name
./s3manager bucket-info --bucket my-other-bucket

# Verbose output
./s3manager bucket-info --verbose
```

**Example Output:**
```json
{
  "bucket_name": "my-bucket",
  "region": "us-west-2",
  "creation_date": "2023-01-15T10:30:00Z",
  "object_count": 1250,
  "total_size_bytes": 524288000,
  "total_size_human": "500.0 MB",
  "last_modified": "2025-03-15T14:22:33Z",
  "api_endpoint": "http://localhost:9000"
}
```

### Delete Old Files

Remove files older than specified days:

```bash
# Delete files older than 30 days (with confirmation prompt)
./s3manager delete-old --days 30

# Delete from specific folder
./s3manager delete-old --days 7 --folder "logs/2023"

# Skip confirmation prompt
./s3manager delete-old --days 30 --confirm

# Dry run (see what would be deleted)
./s3manager delete-old --days 30 --dry-run

# Use different bucket
./s3manager delete-old --days 30 --bucket my-other-bucket
```

**Example Output:**
```json
{
  "bucket_name": "my-bucket",
  "folder": "logs/2023",
  "days_old": 30,
  "deleted_files": [
    "logs/2023/app-2023-01-01.log",
    "logs/2023/app-2023-01-02.log"
  ],
  "deleted_count": 2,
  "total_size_bytes": 2048576,
  "total_size_human": "2.0 MB",
  "operation_time": "2024-03-15T14:22:33Z",
  "cutoff_date": "2024-02-14T14:22:33Z"
}
```

### Upload Files and Folders

Upload files or folders to S3 with optional archiving:

```bash
# Upload a single file (will prompt for archiving)
./s3manager upload document.pdf

# Upload multiple files and folders (archived by default)
./s3manager upload folder1/ file1.txt file2.pdf

# Upload without archiving (individual files)
./s3manager upload file1.txt file2.txt --no-archive

# Upload to specific folder in S3
./s3manager upload data/ --destination "backups/2024"

# Upload with custom archive name
./s3manager upload project/ --archive-name "v1.0.0"
```

**Example Output:**
```json
{
  "bucket_name": "my-bucket",
  "destination_path": "backups/2024",
  "items": [
    {
      "local_path": "data/file1.txt",
      "remote_path": "backups/2024/archive-20240315-142233.zip",
      "size": 1048576,
      "is_archived": true
    }
  ],
  "total_files": 1,
  "total_size_bytes": 1048576,
  "total_size_human": "1.0 MB",
  "operation_time": "2024-03-15T14:22:33Z",
  "archive_created": true,
  "upload_duration": "2.5s"
}
```

## Command Reference

### Global Flags

| Flag            | Description                      | Default     |
|-----------------|----------------------------------|-------------|
| `--bucket, -b`  | Override bucket name from config | From config |
| `--verbose, -v` | Enable verbose output            | `false`     |
| `--help, -h`    | Show help information            |             |

### `bucket-info` Command

Get comprehensive bucket information.

**Flags:**
- `--timeout`: Operation timeout in seconds (default: 300)

### `delete-old` Command

Delete files older than specified days.

**Required Flags:**
- `--days, -d`: Number of days (files older than this will be deleted)

**Optional Flags:**
- `--folder, -f`: Specific folder/prefix to search in
- `--confirm`: Skip confirmation prompt
- `--dry-run`: Show what would be deleted without actually deleting
- `--timeout`: Operation timeout in seconds (default: 1800)


## AWS Permissions

Your AWS credentials need the following permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:ListBucket",
                "s3:GetBucketLocation",
                "s3:ListAllMyBuckets",
                "s3:DeleteObject",
                "s3:PutObject"
            ],
            "Resource": [
                "arn:aws:s3:::*",
                "arn:aws:s3:::*/*"
            ]
        }
    ]
}
```

## Security Considerations

- Never commit `.env` files to version control
- Review delete operations carefully - deletions are permanent

## Testing

The project includes both unit tests and integration tests.

### Running Unit Tests

Unit tests can be run without any external dependencies:

```bash
# Run all unit tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./pkg/utils
```

### Running Integration Tests

Integration tests require a real S3 connection and are skipped by default. To run them:

1. Set up environment variables for testing:

```bash
# Create a .env.test file
cp .env.example .env.test
nano .env.test
```

2. Configure the test environment variables:

```
# S3 connection details for testing
TEST_BUCKET_NAME=your-test-bucket
TEST_REGION=your-test-region
TEST_API_URL=your-test-api-url
TEST_ACCESS_KEY=your-test-access-key
TEST_SECRET_KEY=your-test-secret-key

# Enable integration tests
S3_INTEGRATION_TEST=true
```

3. Run the integration tests:

```bash
# Load test environment variables and run integration tests
export $(grep -v '^#' .env.test | xargs) && go test -v ./...
```

**Note**: Integration tests will create and delete files in your test bucket. Make sure to use a dedicated test bucket, not a production bucket.
