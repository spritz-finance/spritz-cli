package offramps

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:     "off-ramps",
	Aliases: []string{"or"},
	Short:   "View off-ramp transactions",
}

func init() {
	Cmd.AddCommand(listCmd)
}
