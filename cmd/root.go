package cmd

import (
	"github.com/spf13/cobra"
	"s3manager/config"
)

var (
	cfg *config.Config
)

var rootCmd = &cobra.Command{
	Use:   "s3manager",
	Short: "S3 Manager tool for bucket management",
	Long: `S3 Manager is a command-line tool for managing S3 buckets and objects.
It provides functionality to get bucket information and manage old files.
Configuration is loaded from .env file or environment variables`,
}

func Execute(config *config.Config) error {
	cfg = config
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(bucketInfoCmd)
	rootCmd.AddCommand(deleteOldCmd)
	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(downloadCmd)

	rootCmd.PersistentFlags().StringP("bucket", "b", "", "Override bucket name from config")
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")
}

func getBucketName(cmd *cobra.Command) string {
	bucket, _ := cmd.Flags().GetString("bucket")
	if bucket != "" {
		return bucket
	}
	return cfg.BucketName
}

func isVerbose(cmd *cobra.Command) bool {
	verbose, _ := cmd.Flags().GetBool("verbose")
	return verbose
}
