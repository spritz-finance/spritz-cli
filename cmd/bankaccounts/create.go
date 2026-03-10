package bankaccounts

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/internal/api"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new bank account",
	Long: `Create a new bank account as an off-ramp payment destination.

The --type flag determines which additional flags are required:

  US:   --routing-number, --account-number, --ownership
  CA:   --institution-number, --transit-number, --account-number, --ownership
  UK:   --sort-code, --account-number, --ownership
  IBAN: --iban, --ownership (--bic optional)

Examples:
  # Create a US checking account
  spritz bank-accounts create --type us \
    --routing-number 021000021 --account-number 123456789 \
    --account-subtype checking --ownership personal

  # Create an IBAN account
  spritz bank-accounts create --type iban \
    --iban DE89370400440532013000 --ownership personal

  # Using the alias
  spritz ba create --type us --routing-number 021000021 \
    --account-number 123456789 --ownership personal`,
	RunE: runCreate,
}

func init() {
	f := createCmd.Flags()
	f.String("type", "", "Account type: us, ca, uk, iban (required)")
	f.String("ownership", "", "Account ownership: personal, thirdParty (required)")
	f.String("routing-number", "", "ABA routing number (US)")
	f.String("account-number", "", "Bank account number (US, CA, UK)")
	f.String("account-subtype", "", "Account subtype: checking, savings (US, CA)")
	f.String("institution-number", "", "Institution number (CA)")
	f.String("transit-number", "", "Transit number (CA)")
	f.String("sort-code", "", "Sort code (UK)")
	f.String("iban", "", "IBAN (IBAN type)")
	f.String("bic", "", "BIC/SWIFT code (IBAN type, optional)")
	f.String("label", "", "Friendly label for the account (optional)")

	_ = createCmd.MarkFlagRequired("type")
	_ = createCmd.MarkFlagRequired("ownership")
}

func runCreate(cmd *cobra.Command, args []string) error {
	acctType, _ := cmd.Flags().GetString("type")
	ownership, _ := cmd.Flags().GetString("ownership")
	label, _ := cmd.Flags().GetString("label")

	body := map[string]interface{}{
		"type":      acctType,
		"ownership": ownership,
	}
	if label != "" {
		body["label"] = label
	}

	switch acctType {
	case "us":
		if err := requireFlags(cmd, "routing-number", "account-number"); err != nil {
			return err
		}
		body["routingNumber"], _ = cmd.Flags().GetString("routing-number")
		body["accountNumber"], _ = cmd.Flags().GetString("account-number")
		if v, _ := cmd.Flags().GetString("account-subtype"); v != "" {
			body["accountSubtype"] = v
		}

	case "ca":
		if err := requireFlags(cmd, "institution-number", "transit-number", "account-number"); err != nil {
			return err
		}
		body["institutionNumber"], _ = cmd.Flags().GetString("institution-number")
		body["transitNumber"], _ = cmd.Flags().GetString("transit-number")
		body["accountNumber"], _ = cmd.Flags().GetString("account-number")
		if v, _ := cmd.Flags().GetString("account-subtype"); v != "" {
			body["accountSubtype"] = v
		}

	case "uk":
		if err := requireFlags(cmd, "sort-code", "account-number"); err != nil {
			return err
		}
		body["sortCode"], _ = cmd.Flags().GetString("sort-code")
		body["accountNumber"], _ = cmd.Flags().GetString("account-number")

	case "iban":
		if err := requireFlags(cmd, "iban"); err != nil {
			return err
		}
		body["iban"], _ = cmd.Flags().GetString("iban")
		if v, _ := cmd.Flags().GetString("bic"); v != "" {
			body["bic"] = v
		}

	default:
		return fmt.Errorf("unsupported account type %q (must be us, ca, uk, or iban)", acctType)
	}

	client, err := api.NewFromEnv()
	if err != nil {
		return err
	}

	account, err := client.CreateBankAccount(body)
	if err != nil {
		return err
	}

	return json.NewEncoder(os.Stdout).Encode(account)
}

func requireFlags(cmd *cobra.Command, names ...string) error {
	for _, name := range names {
		v, _ := cmd.Flags().GetString(name)
		if v == "" {
			acctType, _ := cmd.Flags().GetString("type")
			return fmt.Errorf("--%s is required for account type %q", name, acctType)
		}
	}
	return nil
}
