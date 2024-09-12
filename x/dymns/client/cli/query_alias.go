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

// CmdQueryAlias is the CLI command for querying Alias information
func CmdQueryAlias() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "alias [alias]",
		Short:   "Get alias (aka handle) information",
		Example: fmt.Sprintf("%s q %s alias myname", version.AppName, dymnstypes.ModuleName),
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			alias := args[0]

			if !dymnsutils.IsValidAlias(alias) {
				return fmt.Errorf("input is not a valid alias: %s", alias)
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := dymnstypes.NewQueryClient(clientCtx)

			res, err := queryClient.Alias(cmd.Context(), &dymnstypes.QueryAliasRequest{
				Alias: alias,
			})
			if err != nil {
				return fmt.Errorf("failed to fetch information of '%s': %w", alias, err)
			}

			if res == nil || res.ChainId == "" {
				fmt.Println("Alias is not registered.")
				return nil
			}

			fmt.Printf("Alias '%s' is being used by '%s'", alias, res.ChainId)
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
