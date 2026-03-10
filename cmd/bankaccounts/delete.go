package bankaccounts

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/internal/api"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <accountId>",
	Short: "Delete a bank account",
	Long: `Delete a bank account by its ID.

Examples:
  spritz bank-accounts delete 507f1f77bcf86cd799439011
  spritz ba delete 507f1f77bcf86cd799439011`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewFromEnv()
		if err != nil {
			return err
		}

		accountID := args[0]
		if err := client.DeleteBankAccount(accountID); err != nil {
			return err
		}

		jsonOut, _ := cmd.Flags().GetBool("json")
		if jsonOut {
			return json.NewEncoder(os.Stdout).Encode(map[string]interface{}{
				"deleted":   true,
				"accountId": accountID,
			})
		}

		fmt.Fprintf(os.Stderr, "Deleted bank account %s\n", accountID)
		return nil
	},
}

func init() {
	deleteCmd.Flags().Bool("json", false, "Print structured JSON output")
}
