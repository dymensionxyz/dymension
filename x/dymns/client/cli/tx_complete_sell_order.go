package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/spf13/cobra"
)

// NewCompleteSellOrderTxCmd is the CLI command for completing a Sell-Order.
func NewCompleteSellOrderTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "complete-sell-order [Name/Alias/Handle] [myname/x]",
		Aliases: []string{"complete-so"},
		Short:   "Complete a sell-order (must be expired and has at least one bid)",
		Long:    "Request to complete a sell-order (must be expired and has at least one bid). Can be submitted by either the owner or the highest bidder.",
		Example: fmt.Sprintf(
			`$ %s tx %s complete-sell-order name myname --%s owner/bidder
$ %s tx %s complete-sell-order alias x --%s owner/bidder`,
			version.AppName, dymnstypes.ModuleName, flags.FlagFrom,
			version.AppName, dymnstypes.ModuleName, flags.FlagFrom,
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			target := args[1]
			var assetType dymnstypes.AssetType

			switch strings.ToLower(args[0]) {
			case "name", "dym-name", "dymname", "n":
				assetType = dymnstypes.TypeName
				if !dymnsutils.IsValidDymName(target) {
					return fmt.Errorf("input is not a valid Dym-Name: %s", target)
				}
			case "alias", "handle", "handles", "a":
				assetType = dymnstypes.TypeAlias
				if !dymnsutils.IsValidAlias(target) {
					return fmt.Errorf("input is not a valid Alias: %s", target)
				}
			default:
				return fmt.Errorf("invalid asset type: %s, must be 'Name' or 'Alias'/'Handle'", args[0])
			}

			participant := clientCtx.GetFromAddress().String()
			if participant == "" {
				return fmt.Errorf("flag --%s is required", flags.FlagFrom)
			}

			msg := &dymnstypes.MsgCompleteSellOrder{
				AssetId:     target,
				AssetType:   assetType,
				Participant: participant,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
