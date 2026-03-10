package offrampquotes

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/internal/api"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an off-ramp quote",
	Long: `Create a new off-ramp quote for converting crypto to fiat.

The quote contains pricing, fees, and fulfillment instructions.
Use the returned quote ID to generate a transaction for signing.

Examples:
  # Basic quote for $100 USD output
  spritz off-ramp-quotes create \
    --account-id 507f1f77bcf86cd799439011 \
    --amount 100 --chain ethereum

  # Specify input amount mode (pay exactly X in crypto)
  spritz off-ramp-quotes create \
    --account-id 507f1f77bcf86cd799439011 \
    --amount 100 --chain polygon --amount-mode input

  # Specify rail and token
  spritz off-ramp-quotes create \
    --account-id 507f1f77bcf86cd799439011 \
    --amount 50 --chain base --rail rtp \
    --token-address 0x833589fCD6eDb6E08f4c7C32D4f71b54bdA02913`,
	RunE: func(cmd *cobra.Command, args []string) error {
		accountID, _ := cmd.Flags().GetString("account-id")
		amount, _ := cmd.Flags().GetString("amount")
		chain, _ := cmd.Flags().GetString("chain")
		amountMode, _ := cmd.Flags().GetString("amount-mode")
		tokenAddr, _ := cmd.Flags().GetString("token-address")
		rail, _ := cmd.Flags().GetString("rail")
		memo, _ := cmd.Flags().GetString("memo")

		client, err := api.NewFromEnv()
		if err != nil {
			return err
		}

		raw, err := client.CreateQuote(&api.CreateQuoteRequest{
			AccountID:    accountID,
			Amount:       amount,
			Chain:        chain,
			AmountMode:   amountMode,
			TokenAddress: tokenAddr,
			Rail:         rail,
			Memo:         memo,
		})
		if err != nil {
			return err
		}

		return json.NewEncoder(os.Stdout).Encode(json.RawMessage(raw))
	},
}

func init() {
	f := createCmd.Flags()
	f.String("account-id", "", "Destination bank account ID (required)")
	f.String("amount", "", "Amount as decimal string (required)")
	f.String("chain", "", "Blockchain network (required)")
	f.String("amount-mode", "", "Amount mode: input, output (default: output)")
	f.String("token-address", "", "Token contract address or symbol")
	f.String("rail", "", "Payment rail: ach_standard, rtp, wire, eft, sepa, push_to_debit, bill_pay")
	f.String("memo", "", "Payment memo/note")

	_ = createCmd.MarkFlagRequired("account-id")
	_ = createCmd.MarkFlagRequired("amount")
	_ = createCmd.MarkFlagRequired("chain")
}
