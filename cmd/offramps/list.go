package offramps

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/internal/api"
	"github.com/spritz-finance/spritz-cli/internal/format"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List off-ramp transactions",
	Long: `List off-ramp transactions with optional filtering and pagination.

Examples:
  # List recent off-ramps
  spritz off-ramps list

  # Filter by status
  spritz off-ramps list --status completed

  # Filter by chain and account
  spritz off-ramps list --chain ethereum --account-id 507f1f77bcf86cd799439011

  # Paginate through results
  spritz off-ramps list --limit 10
  spritz off-ramps list --limit 10 --cursor <nextCursor>

  # JSON output
  spritz off-ramps list -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewFromEnv()
		if err != nil {
			return err
		}

		limit, _ := cmd.Flags().GetInt("limit")
		cursor, _ := cmd.Flags().GetString("cursor")
		status, _ := cmd.Flags().GetString("status")
		chain, _ := cmd.Flags().GetString("chain")
		accountID, _ := cmd.Flags().GetString("account-id")
		sort, _ := cmd.Flags().GetString("sort")

		result, err := client.ListOffRamps(&api.ListOffRampsParams{
			Limit:     limit,
			Cursor:    cursor,
			Status:    status,
			Chain:     chain,
			AccountID: accountID,
			Sort:      sort,
		})
		if err != nil {
			return err
		}

		headers := []string{"id", "status", "chain", "amount", "currency", "accountId", "createdAt"}
		rows := make([][]string, len(result.Data))
		for i, r := range result.Data {
			rows[i] = []string{
				r.ID,
				r.Status,
				r.Chain,
				r.Payout.Amount,
				r.Payout.Currency,
				r.AccountID,
				r.CreatedAt,
			}
		}

		if err := format.Global.Write(headers, rows); err != nil {
			return err
		}

		if result.HasMore && result.NextCursor != "" {
			fmt.Fprintf(os.Stderr, "nextCursor: %s\n", result.NextCursor)
		}

		return nil
	},
}

func init() {
	f := listCmd.Flags()
	f.Int("limit", 50, "Maximum number of results (1-100)")
	f.String("cursor", "", "Pagination cursor from previous response")
	f.String("status", "", "Filter by status: awaiting_funding, queued, in_flight, completed, canceled, failed, reversed, refunded")
	f.String("chain", "", "Filter by blockchain network")
	f.String("account-id", "", "Filter by destination account ID")
	f.String("sort", "", "Sort order: asc, desc")
}
