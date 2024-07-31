package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func CmdQueryResolveDymNameAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "resolve-dym-name-address [Dym-Name-Address...]",
		Aliases: []string{"resolve", "r"},
		Short:   "Resolve Dym-Name-Address",
		Example: fmt.Sprintf(
			"%s q %s resolve myname@dym",
			version.AppName, dymnstypes.ModuleName,
		),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := dymnstypes.NewQueryClient(clientCtx)

			res, err := queryClient.ResolveDymNameAddresses(cmd.Context(), &dymnstypes.QueryResolveDymNameAddressesRequest{
				Addresses: args,
			})
			if err != nil {
				return fmt.Errorf("failed to resolve: %w", err)
			}

			if res == nil || len(res.ResolvedAddresses) == 0 {
				return fmt.Errorf("no result, maybe not configured or expired")
			}

			for i, resolvedAddress := range res.ResolvedAddresses {
				if i > 0 {
					fmt.Println()
				}

				fmt.Printf("%s\n", resolvedAddress.Address)
				if resolvedAddress.Error != "" {
					fmt.Printf(" => Error: %s\n", resolvedAddress.Error)
				} else {
					fmt.Printf(" => %s\n", resolvedAddress.ResolvedAddress)
				}
			}

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
