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

// CmdQueryDymName is the CLI command for querying Dym-Name information
func CmdQueryDymName() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "dym-name [Dym-Name]",
		Aliases: []string{"name"},
		Short:   "Get Dym-Name information",
		Example: fmt.Sprintf("%s q %s dym-name myname", version.AppName, dymnstypes.ModuleName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dymName := args[0]

			if !dymnsutils.IsValidDymName(dymName) {
				return fmt.Errorf("input is not a valid Dym-Name: %s", dymName)
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := dymnstypes.NewQueryClient(clientCtx)

			res, err := queryClient.DymName(cmd.Context(), &dymnstypes.QueryDymNameRequest{
				DymName: dymName,
			})
			if err != nil {
				return fmt.Errorf("failed to fetch information of '%s': %w", dymName, err)
			}

			if res == nil || res.DymName == nil {
				return fmt.Errorf("Dym-Name is not registered or expired: %s", dymName)
			}

			return clientCtx.PrintProto(res.DymName)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
