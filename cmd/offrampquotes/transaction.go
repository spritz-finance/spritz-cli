package offrampquotes

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/internal/api"
)

var transactionCmd = &cobra.Command{
	Use:   "transaction <quoteId>",
	Short: "Get transaction parameters for a quote",
	Long: `Generate blockchain transaction parameters for an off-ramp quote.

Returns either EVM calldata or a Solana serialized transaction,
depending on the quote's chain.

Examples:
  # EVM transaction (Ethereum, Polygon, Base, etc.)
  spritz off-ramp-quotes transaction 507f1f77bcf86cd799439011 \
    --sender-address 0x1234...

  # Solana transaction
  spritz off-ramp-quotes transaction 507f1f77bcf86cd799439011 \
    --sender-address So1ana... --fee-payer FeePayerAddr...`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		senderAddr, _ := cmd.Flags().GetString("sender-address")
		feePayer, _ := cmd.Flags().GetString("fee-payer")

		client, err := api.NewFromEnv()
		if err != nil {
			return err
		}

		raw, err := client.CreateTransaction(args[0], &api.CreateTransactionRequest{
			SenderAddress: senderAddr,
			FeePayer:      feePayer,
		})
		if err != nil {
			return err
		}

		return json.NewEncoder(os.Stdout).Encode(json.RawMessage(raw))
	},
}

func init() {
	f := transactionCmd.Flags()
	f.String("sender-address", "", "Wallet address of the sender (required)")
	f.String("fee-payer", "", "Fee payer address (Solana only)")

	_ = transactionCmd.MarkFlagRequired("sender-address")
}
