package bankaccounts

import (
	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/internal/api"
	"github.com/spritz-finance/spritz-cli/internal/format"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all bank accounts",
	Long: `List all bank accounts saved as off-ramp payment destinations.

Returns account details including type, status, currency, and masked account numbers.

Examples:
  # List all accounts as CSV (default)
  spritz bank-accounts list

  # JSON output
  spritz bank-accounts list -o json

  # Get just the IDs (skip header, extract first column)
  spritz bank-accounts list --no-header | cut -d, -f1

  # Using the alias
  spritz ba list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewFromEnv()
		if err != nil {
			return err
		}

		accounts, err := client.ListBankAccounts()
		if err != nil {
			return err
		}

		headers := []string{
			"id", "status", "accountHolderName", "supportedRails",
			"label", "createdAt", "type", "currency",
			"accountNumberLast4", "routingNumberLast4",
		}

		rows := make([][]string, len(accounts))
		for i, a := range accounts {
			rows[i] = []string{
				a.ID,
				a.Status,
				a.AccountHolderName,
				a.SupportedRailsStr(),
				a.Label,
				a.CreatedAt,
				a.Type,
				a.Currency,
				a.AccountNumberLast4,
				a.RoutingNumberLast4,
			}
		}

		return format.Global.Write(headers, rows)
	},
}
