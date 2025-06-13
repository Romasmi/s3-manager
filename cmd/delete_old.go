package cmd

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"s3manager/internal/s3client"
	"s3manager/pkg/utils"
	"time"
)

var deleteOldCmd = &cobra.Command{
	Use:   "delete-old",
	Short: "Delete files older than specified days",
	Long: `Delete files in the S3 bucket that are older than the specified number of days.

The command will:
- List all objects in the specified folder (or entire bucket if no folder specified)
- Filter objects older than the cutoff date
- Delete matching objects in batches
- Return detailed information about the deletion operation

WARNING: This operation is irreversible. Deleted files cannot be recovered.`,
	Example: `  # Delete files older than 30 days from entire bucket
  s3manager delete-old --days 30

  # Delete files older than 7 days from specific folder
  s3manager delete-old --days 7 --folder "logs/2025"

  # Delete with confirmation and verbose output
  s3manager delete-old --days 30 --folder "temp" --confirm --verbose

  # Use different bucket
  s3manager delete-old --days 30 --bucket my-other-bucket`,
	Run: func(cmd *cobra.Command, args []string) {
		runDeleteOld(cmd)
	},
}

func runDeleteOld(cmd *cobra.Command) {
	days, _ := cmd.Flags().GetInt("days")
	folder, _ := cmd.Flags().GetString("folder")
	confirm, _ := cmd.Flags().GetBool("confirm")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if days <= 0 {
		err := fmt.Errorf("days must be greater than 0")
		utils.PrintError(err, "delete-old")
		return
	}

	// Show confirmation prompt if not in confirm mode and not dry-run
	if !confirm && !dryRun {
		cutoffDate := time.Now().AddDate(0, 0, -days)
		bucketName := getBucketName(cmd)

		fmt.Printf("WARNING: This will permanently delete files older than %d days (%s) from bucket '%s'",
			days, cutoffDate.Format("2006-01-02"), bucketName)

		if folder != "" {
			fmt.Printf(" in folder '%s'", folder)
		}
		fmt.Println()
		fmt.Print("Are you sure? (yes/no): ")

		var response string
		fmt.Scanln(&response)
		if response != "yes" && response != "y" && response != "YES" {
			fmt.Println("Operation cancelled.")
			return
		}
	}

	client, err := s3client.New(cfg)
	if err != nil {
		utils.PrintError(err, "delete-old")
		return
	}

	timeout, _ := cmd.Flags().GetInt("timeout")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	if isVerbose(cmd) {
		cmd.Printf("Deleting files older than %d days from bucket: %s\n", days, getBucketName(cmd))
		if folder != "" {
			cmd.Printf("Folder: %s\n", folder)
		}
		if dryRun {
			cmd.Println("DRY RUN MODE: No files will actually be deleted")
		}
	}

	result, err := client.DeleteOldFiles(ctx, folder, days, dryRun)
	if err != nil {
		utils.PrintError(err, "delete-old")
		return
	}

	if err := utils.PrintJSON(result); err != nil {
		utils.PrintError(err, "delete-old")
		return
	}

	if isVerbose(cmd) {
		cmd.Println("Delete operation completed successfully")
	}
}

func init() {
	deleteOldCmd.Flags().IntP("days", "d", 0, "Delete files older than this many days (required)")
	err := deleteOldCmd.MarkFlagRequired("days")
	if err != nil {
		utils.PrintError(err, "delete-old")
		return
	}

	deleteOldCmd.Flags().StringP("folder", "f", "", "Folder/prefix to search in (optional, searches entire bucket if not specified)")
	deleteOldCmd.Flags().Bool("confirm", false, "Skip confirmation prompt")
	deleteOldCmd.Flags().Bool("dry-run", false, "Show what would be deleted without actually deleting")
	deleteOldCmd.Flags().Int("timeout", 1800, "Timeout in seconds for the operation (default: 30 minutes)")

	deleteOldCmd.SetUsageTemplate(`Usage:{{if .Runnable}}
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
