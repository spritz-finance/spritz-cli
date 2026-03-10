package bankaccounts

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:     "bank-accounts",
	Aliases: []string{"ba"},
	Short:   "Manage bank accounts used as off-ramp destinations",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(deleteCmd)
}
