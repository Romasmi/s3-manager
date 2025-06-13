package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"s3manager/internal/s3client"
	"s3manager/pkg/utils"
	"strings"
	"time"
)

var uploadCmd = &cobra.Command{
	Use:   "upload [files/folders...]",
	Short: "Upload files or folders to S3",
	Long: `Upload files or folders to S3 bucket with optional archiving.

By default, the command will create a zip archive of the specified files/folders
before uploading to S3.
You can disable archiving with the --no-archive flag to upload files individually.

The destination path in S3 can be specified with the --destination flag.
If not specified, files will be uploaded to the root of the bucket.`,
	Example: `  # Upload single file (archived by default)
  s3manager upload document.pdf

  # Upload multiple files and folders (archived)
  s3manager upload folder1/ file1.txt file2.pdf

  # Upload to specific S3 folder
  s3manager upload data/ --destination "backups/2024"

  # Upload without archiving (individual files)
  s3manager upload file1.txt file2.txt --no-archive

  # Upload with custom archive name
  s3manager upload project/ --destination "releases" --archive-name "v1.0.0"

  # Upload with different bucket
  s3manager upload data/ --bucket my-other-bucket

  # Verbose upload with progress
  s3manager upload large-folder/ --verbose`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runUpload(cmd, args)
	},
}

func runUpload(cmd *cobra.Command, args []string) {
	destination, _ := cmd.Flags().GetString("destination")
	noArchive, _ := cmd.Flags().GetBool("no-archive")
	archiveName, _ := cmd.Flags().GetString("archive-name")
	confirm, _ := cmd.Flags().GetBool("confirm")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if err := utils.ValidatePaths(args); err != nil {
		utils.PrintError(err, "upload")
		return
	}

	// Determine if we should archive (default: true, unless --no-archive is specified)
	shouldArchive := !noArchive

	if len(args) == 1 && !noArchive && !confirm {
		err := utils.ValidatePaths([]string{args[0]})
		if err == nil {
			if !isDirectory(args[0]) {
				fmt.Printf("Upload single file '%s' as archive? (y/N): ", args[0])
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "yes" && response != "Y" && response != "YES" {
					shouldArchive = false
				}
			}
		}
	}

	// Show operation summary if not in confirm mode and not dry-run
	if !confirm && !dryRun {
		bucketName := getBucketName(cmd)

		fmt.Printf("Upload operation summary:\n")
		fmt.Printf("  Bucket: %s\n", bucketName)
		fmt.Printf("  Destination: %s\n", getDestinationDisplay(destination))
		fmt.Printf("  Files/Folders: %v\n", args)
		fmt.Printf("  Archive: %t\n", shouldArchive)

		if shouldArchive && archiveName != "" {
			fmt.Printf("  Archive name: %s\n", archiveName)
		}

		fmt.Print("Continue with upload? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "yes" && response != "Y" && response != "YES" {
			fmt.Println("Upload cancelled.")
			return
		}
	}

	client, err := s3client.New(cfg)
	if err != nil {
		utils.PrintError(err, "upload")
		return
	}

	timeout, _ := cmd.Flags().GetInt("timeout")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	if isVerbose(cmd) {
		cmd.Printf("Starting upload operation...\n")
		cmd.Printf("  Paths: %v\n", args)
		cmd.Printf("  Destination: %s\n", getDestinationDisplay(destination))
		cmd.Printf("  Archive: %t\n", shouldArchive)
		if dryRun {
			cmd.Println("  DRY RUN MODE: No files will actually be uploaded")
		}
	}

	if dryRun {
		result := createDryRunResult(args, destination, shouldArchive, getBucketName(cmd))
		if err := utils.PrintJSON(result); err != nil {
			utils.PrintError(err, "upload")
			return
		}
	} else {
		result, err := client.UploadFiles(ctx, args, destination, shouldArchive)
		if err != nil {
			utils.PrintError(err, "upload")
			return
		}

		if bucketFlag := getBucketName(cmd); bucketFlag != cfg.BucketName {
			result.BucketName = bucketFlag
		}

		if err := utils.PrintJSON(result); err != nil {
			utils.PrintError(err, "upload")
			return
		}
	}

	if isVerbose(cmd) {
		cmd.Println("Upload operation completed successfully")
	}
}

func isDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

func getDestinationDisplay(destination string) string {
	if destination == "" {
		return "bucket root"
	}
	return destination
}

func createDryRunResult(paths []string, destination string, shouldArchive bool, bucketName string) interface{} {
	items := make([]interface{}, 0)

	if shouldArchive {
		archiveName := utils.GenerateArchiveName(paths, ".zip")
		remotePath := destination
		if remotePath != "" && !strings.HasSuffix(remotePath, "/") {
			remotePath += "/"
		}
		remotePath += archiveName

		items = append(items, map[string]interface{}{
			"local_path":  strings.Join(paths, ", "),
			"remote_path": remotePath,
			"size":        0,
			"is_archived": true,
		})
	} else {
		for _, path := range paths {
			remotePath := destination
			if remotePath != "" && !strings.HasSuffix(remotePath, "/") {
				remotePath += "/"
			}
			remotePath += filepath.Base(path)

			items = append(items, map[string]interface{}{
				"local_path":  path,
				"remote_path": remotePath,
				"size":        0,
				"is_archived": false,
			})
		}
	}

	return map[string]interface{}{
		"bucket_name":      bucketName,
		"destination_path": destination,
		"items":            items,
		"total_files":      len(items),
		"total_size_bytes": 0,
		"total_size_human": "0 B",
		"operation_time":   utils.FormatTime(time.Now()),
		"archive_created":  shouldArchive,
		"upload_duration":  "0s",
		"dry_run":          true,
	}
}

func init() {
	uploadCmd.Flags().StringP("destination", "d", "", "Destination folder in S3 bucket (optional)")
	uploadCmd.Flags().Bool("no-archive", false, "Upload files individually without creating archive")
	uploadCmd.Flags().StringP("archive-name", "a", "", "Custom name for the archive file (only used with archiving)")
	uploadCmd.Flags().Bool("confirm", false, "Skip confirmation prompt")
	uploadCmd.Flags().Bool("dry-run", false, "Show what would be uploaded without actually uploading")
	uploadCmd.Flags().Int("timeout", 3600, "Timeout in seconds for the operation (default: 1 hour)")

	uploadCmd.SetUsageTemplate(`Usage:{{if .Runnable}}
  {{.UseLine}}{{end}}{{if .HasAvailableSubCommands}}
  {{.CommandPath}} [command]{{end}}{{if gt (len .Aliases) 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

Flags:
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`)
}
