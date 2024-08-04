package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        dymnstypes.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", dymnstypes.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		NewRegisterDymNameTxCmd(),
		NewUpdateResolveDymNameAddressTxCmd(),
		NewUpdateDetailsTxCmd(),
		NewPlaceDymNameSellOrderTxCmd(),
		NewPlaceBidOnOrderTxCmd(),
		NewOfferBuyDymNameTxCmd(),
		NewAcceptDymNameBuyOfferTxCmd(),
	)

	return cmd
}
