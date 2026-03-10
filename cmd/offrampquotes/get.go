package offrampquotes

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/internal/api"
)

var getCmd = &cobra.Command{
	Use:   "get <quoteId>",
	Short: "Get an off-ramp quote by ID",
	Long: `Retrieve the details of an existing off-ramp quote.

Examples:
  spritz off-ramp-quotes get 507f1f77bcf86cd799439011
  spritz orq get 507f1f77bcf86cd799439011`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := api.NewFromEnv()
		if err != nil {
			return err
		}

		raw, err := client.GetQuote(args[0])
		if err != nil {
			return err
		}

		return json.NewEncoder(os.Stdout).Encode(json.RawMessage(raw))
	},
}
