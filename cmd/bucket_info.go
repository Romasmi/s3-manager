package cmd

import (
	"context"
	"github.com/spf13/cobra"
	"s3manager/internal/s3client"
	"s3manager/pkg/utils"
	"time"
)

var bucketInfoCmd = &cobra.Command{
	Use:   "bucket-info",
	Short: "Get comprehensive bucket information",
	Long: `Get detailed information about the S3 bucket
The bucket name is taken from the configuration file unless overridden with --bucket flag.`,
	Example: `  # Get info for configured bucket
  s3manager bucket-info

  # Get info for specific bucket
  s3manager bucket-info --bucket my-other-bucket

  # Verbose output
  s3manager bucket-info --verbose`,
	Run: func(cmd *cobra.Command, args []string) {
		runBucketInfo(cmd)
	},
}

func runBucketInfo(cmd *cobra.Command) {
	client, err := s3client.New(cfg)
	if err != nil {
		utils.PrintError(err, "bucket-info")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if isVerbose(cmd) {
		cmd.Printf("Getting bucket information for: %s\n", getBucketName(cmd))
	}

	info, err := client.GetBucketInfo(ctx)
	if err != nil {
		utils.PrintError(err, "bucket-info")
		return
	}

	if bucketFlag := getBucketName(cmd); bucketFlag != cfg.BucketName {
		info.BucketName = bucketFlag
	}

	if err := utils.PrintJSON(info); err != nil {
		utils.PrintError(err, "bucket-info")
		return
	}

	if isVerbose(cmd) {
		cmd.Printf("Bucket info retrieved successfully\n")
	}
}

func init() {
	bucketInfoCmd.Flags().Int("timeout", 300, "Timeout in seconds for the operation")
}
