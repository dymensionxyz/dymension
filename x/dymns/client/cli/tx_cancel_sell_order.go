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

// NewCancelSellOrderTxCmd is the CLI command for close a Sell-Order.
func NewCancelSellOrderTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cancel-sell-order [Name/Alias/Handle] [myname/x]",
		Aliases: []string{"cancel-so"},
		Short:   "Cancel a sell-order (only when no bid placed)",
		Example: fmt.Sprintf(
			`$ %s tx %s cancel-sell-order name myname --%s owner
$ %s tx %s cancel-sell-order alias x --%s owner`,
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

			owner := clientCtx.GetFromAddress().String()
			if owner == "" {
				return fmt.Errorf("flag --%s is required", flags.FlagFrom)
			}

			msg := &dymnstypes.MsgCancelSellOrder{
				AssetId:   target,
				AssetType: assetType,
				Owner:     owner,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
