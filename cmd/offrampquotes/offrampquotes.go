package offrampquotes

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:     "off-ramp-quotes",
	Aliases: []string{"orq"},
	Short:   "Create and manage off-ramp quotes",
}

func init() {
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(transactionCmd)
}
