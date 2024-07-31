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

func CmdQuerySellOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sell-order [Dym-Name]",
		Aliases: []string{"so", "sellorder"},
		Short:   "Get current active Sell Order of a Dym-Name.",
		Example: fmt.Sprintf(
			"%s q %s sell-order myname",
			version.AppName, dymnstypes.ModuleName,
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dymName := args[0]

			if !dymnsutils.IsValidDymName(dymName) {
				return fmt.Errorf("input Dym-Name '%s' is not a valid Dym-Name", dymName)
			}

			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := dymnstypes.NewQueryClient(clientCtx)

			res, err := queryClient.SellOrder(cmd.Context(), &dymnstypes.QuerySellOrderRequest{
				DymName: dymName,
			})
			if err != nil {
				return fmt.Errorf("failed to fetch Sell Order of '%s': %w", dymName, err)
			}

			if res == nil {
				return fmt.Errorf("no active Sell Order of '%s'", dymName)
			}

			return clientCtx.PrintProto(&res.Result)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
