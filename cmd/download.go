package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"s3manager/internal/s3client"
	"s3manager/pkg/utils"
	"slices"
	"strings"
	"time"
)

var downloadCmd = &cobra.Command{
	Use:   "download [folder]",
	Short: "Download the latest file from a specific folder",
	Long: `Download the latest file from a specific folder in an S3 bucket.

This command lists all files in the specified folder, sorts them by last modified date,
and downloads the most recent file to the specified destination path.

If no destination is specified, the file will be downloaded to the current directory.`,
	Example: `  # Download the latest file from a folder
  s3manager download backups/

  # Download to a specific destination
  s3manager download logs/ --destination /tmp/downloads/

  # Download from a different bucket
  s3manager download data/ --bucket my-other-bucket

  # Verbose download with progress
  s3manager download archives/ --verbose`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runDownload(cmd, args)
	},
}

func runDownload(cmd *cobra.Command, args []string) {
	folder := args[0]
	destination, _ := cmd.Flags().GetString("destination")
	confirm, _ := cmd.Flags().GetBool("confirm")

	// If destination is empty, use current directory
	if destination == "" {
		destination = "."
	}

	// Show operation summary if not in confirm mode
	if !confirm {
		bucketName := getBucketName(cmd)

		fmt.Printf("Download operation summary:\n")
		fmt.Printf("Bucket: %s\n", bucketName)
		fmt.Printf("Folder: %s\n", folder)
		fmt.Printf("Destination: %s\n", destination)

		fmt.Print("Continue with download? (y/N): ")
		var response string
		_, err := fmt.Scanln(&response)
		if err != nil {
			utils.PrintError(err, "download")
			return
		}
		if !slices.Contains([]string{"y", "yes"}, strings.ToLower(response)) {
			fmt.Println("Download cancelled.")
			return
		}
	}

	client, err := s3client.New(cfg)
	if err != nil {
		utils.PrintError(err, "download")
		return
	}

	timeout, _ := cmd.Flags().GetInt("timeout")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	if isVerbose(cmd) {
		cmd.Printf("Starting download operation...\n")
		cmd.Printf("  Folder: %s\n", folder)
		cmd.Printf("  Destination: %s\n", destination)
	}

	result, err := client.DownloadLatestFile(ctx, folder, destination)
	if err != nil {
		utils.PrintError(err, "download")
		return
	}

	if bucketFlag := getBucketName(cmd); bucketFlag != cfg.BucketName {
		result.BucketName = bucketFlag
	}

	if err := utils.PrintJSON(result); err != nil {
		utils.PrintError(err, "download")
		return
	}

	if isVerbose(cmd) {
		cmd.Println("Download operation completed successfully")
		cmd.Printf("Downloaded file: %s\n", result.Items[0].LocalPath)
	}
}

func init() {
	downloadCmd.Flags().StringP("destination", "d", "", "Local destination path (default: current directory)")
	downloadCmd.Flags().Bool("confirm", false, "Skip confirmation prompt")
	downloadCmd.Flags().Int("timeout", 3600, "Timeout in seconds for the operation (default: 1 hour)")

	downloadCmd.SetUsageTemplate(`Usage:{{if .Runnable}}
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
