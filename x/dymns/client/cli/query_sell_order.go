package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/spf13/cobra"

	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// CmdQuerySellOrder is the CLI command for querying the current active Sell Order of a Dym-Name
func CmdQuerySellOrder() *cobra.Command {
	targetSellOrderTypeDymName := dymnstypes.NameOrder.FriendlyString()
	targetSellOrderTypeAlias := dymnstypes.AliasOrder.FriendlyString()

	cmd := &cobra.Command{
		Use:     "sell-order [Dym-Name/Alias]",
		Aliases: []string{"so", "sellorder"},
		Short:   "Get current active Sell Order of a Dym-Name/Alias.",
		Example: fmt.Sprintf(
			`%s q %s sell-order my-name --%s=%s
%s q %s sell-order dym --%s=%s`,
			version.AppName, dymnstypes.ModuleName, flagTargetType, targetSellOrderTypeDymName,
			version.AppName, dymnstypes.ModuleName, flagTargetType, targetSellOrderTypeAlias,
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := args[0]

			targetType, _ := cmd.Flags().GetString(flagTargetType)
			if targetType == "" {
				return fmt.Errorf("flag --%s is required", flagTargetType)
			}

			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := dymnstypes.NewQueryClient(clientCtx)
			switch targetType {
			case targetSellOrderTypeDymName:
				if !dymnsutils.IsValidDymName(input) {
					return fmt.Errorf("input is not a valid Dym-Name: %s", input)
				}
			case targetSellOrderTypeAlias:
				if !dymnsutils.IsValidAlias(input) {
					return fmt.Errorf("input is not a valid Alias: %s", input)
				}
			default:
				return fmt.Errorf("invalid target type: %s", targetType)
			}

			res, err := queryClient.SellOrder(cmd.Context(), &dymnstypes.QuerySellOrderRequest{
				GoodsId: input,
			})
			if err != nil {
				return fmt.Errorf("failed to fetch Sell Order of '%s': %w", input, err)
			}

			if res == nil {
				return fmt.Errorf("no active Sell Order of '%s'", input)
			}

			return clientCtx.PrintProto(&res.Result)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	cmd.Flags().String(flagTargetType, targetSellOrderTypeDymName, fmt.Sprintf("Target type to query for, one of: %s/%s", targetSellOrderTypeDymName, targetSellOrderTypeAlias))

	return cmd
}
