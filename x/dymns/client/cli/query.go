package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"

	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group DymNS queries under a subcommand
	cmd := &cobra.Command{
		Use:                        dymnstypes.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", dymnstypes.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryParams(),
		CmdQueryDymName(),
		CmdQuerySellOrder(),
		CmdQueryOfferToBuy(),
		CmdQueryResolveDymNameAddress(),
		CmdQueryReverseResolveDymNameAddress(),
	)

	return cmd
}
