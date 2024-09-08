package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/dymensionxyz/dymension/v3/x/iro/types"
	"github.com/spf13/cobra"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdCreateIRO())
	cmd.AddCommand(CmdBuy())
	cmd.AddCommand(CmdSell())
	cmd.AddCommand(CmdClaim())

	return cmd
}
