package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"

	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// CmdQuerySellOrder is the CLI command for querying the current active Sell Order of a Dym-Name
func CmdQuerySellOrder() *cobra.Command {
	targetSellOrderAssetTypeDymName := dymnstypes.TypeName.PrettyName()
	targetSellOrderAssetTypeAlias := dymnstypes.TypeAlias.PrettyName()

	cmd := &cobra.Command{
		Use:     "sell-order [Dym-Name/Alias]",
		Aliases: []string{"so", "sellorder"},
		Short:   "Get current active Sell Order of a Dym-Name/Alias.",
		Example: fmt.Sprintf(
			`%s q %s sell-order my-name --%s=%s
%s q %s sell-order dym --%s=%s`,
			version.AppName, dymnstypes.ModuleName, flagTargetType, targetSellOrderAssetTypeDymName,
			version.AppName, dymnstypes.ModuleName, flagTargetType, targetSellOrderAssetTypeAlias,
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			input := args[0]

			targetType, err := cmd.Flags().GetString(flagTargetType)
			if err != nil {
				return err
			}
			if targetType == "" {
				return fmt.Errorf("flag --%s is required", flagTargetType)
			}

			switch targetType {
			case targetSellOrderAssetTypeDymName:
				if !dymnsutils.IsValidDymName(input) {
					return fmt.Errorf("input is not a valid Dym-Name: %s", input)
				}
			case targetSellOrderAssetTypeAlias:
				if !dymnsutils.IsValidAlias(input) {
					return fmt.Errorf("input is not a valid Alias: %s", input)
				}
			default:
				return fmt.Errorf("invalid target type: %s", targetType)
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := dymnstypes.NewQueryClient(clientCtx)

			res, err := queryClient.SellOrder(cmd.Context(), &dymnstypes.QuerySellOrderRequest{
				AssetId:   input,
				AssetType: targetType,
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

	cmd.Flags().String(flagTargetType, targetSellOrderAssetTypeDymName, fmt.Sprintf("Target type to query for, one of: %s/%s", targetSellOrderAssetTypeDymName, targetSellOrderAssetTypeAlias))

	return cmd
}
