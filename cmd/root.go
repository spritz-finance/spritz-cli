package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/cmd/bankaccounts"
	"github.com/spritz-finance/spritz-cli/internal/api"
	"github.com/spritz-finance/spritz-cli/internal/auth"
	"github.com/spritz-finance/spritz-cli/internal/config"
	"github.com/spritz-finance/spritz-cli/internal/format"
	"github.com/spritz-finance/spritz-cli/internal/update"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
	api.Version = v
}

var rootCmd = &cobra.Command{
	Use:   "spritz",
	Short: "CLI for the Spritz payments API",
	Long:  "spritz is a fast, agent-optimized CLI for interacting with the Spritz platform.",
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		config.Init()

		outputFlag, _ := cmd.Flags().GetString("output")
		noHeader, _ := cmd.Flags().GetBool("no-header")
		resolved := format.ResolveFormat(outputFlag)
		format.Global = format.New(resolved, noHeader, os.Stdout)

		// Non-blocking update check (skip for version/update commands)
		name := cmd.Name()
		if name != "version" && name != "update" && name != "login" && name != "logout" {
			go func() {
				if latest := update.Check(version); latest != "" {
					updateNotice = latest
				}
			}()
		}

		return nil
	},
}

var updateNotice string

func init() {
	rootCmd.PersistentFlags().StringP("output", "o", "",
		"Output format: csv, json, table (default: csv, or $SPRITZ_OUTPUT)")
	rootCmd.PersistentFlags().Bool("no-header", false,
		"Omit header row from CSV output (useful for piping)")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(updateCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
	rootCmd.AddCommand(whoamiCmd)
	rootCmd.AddCommand(bankaccounts.Cmd)
}

func Execute() error {
	err := rootCmd.Execute()
	update.PrintNotice(version, updateNotice)

	if err != nil {
		exitCode := 1
		var apiErr *api.APIError
		if errors.As(err, &apiErr) {
			exitCode = apiErr.ExitCode()
		} else if errors.Is(err, auth.ErrNotAuthenticated) {
			exitCode = 2
		}
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(exitCode)
	}
	return nil
}
