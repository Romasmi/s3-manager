# S3 Manager CLI Tool

A Go-based CLI tool for managing S3-compatible storage.

## Features

- 📊 **Bucket Information**: Get detailed bucket statistics including object count, total size, and metadata
- 🗂️ **File Cleanup**: Delete files older than specified days with folder-specific targeting
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
                "s3:DeleteObject"
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