package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/spf13/cobra"
)

const (
	flagWorkingChainId = "working-chain-id"
)

//goland:noinspection SpellCheckingInspection
func CmdQueryReverseResolveDymNameAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "reverse-resolve-dym-name-address [Bech32 Address/0x Address]",
		Aliases: []string{"reverse-resolve", "rr"},
		Short:   "Reverse-resolve Dym-Name-Address from bech32 or 0x address",
		Example: fmt.Sprintf(
			"%s q %s reverse-resolve dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			version.AppName, dymnstypes.ModuleName,
		),
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			workingChainId, _ := cmd.Flags().GetString(flagWorkingChainId)

			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := dymnstypes.NewQueryClient(clientCtx)

			res, err := queryClient.ReverseResolveAddress(cmd.Context(), &dymnstypes.QueryReverseResolveAddressRequest{
				Addresses:      args,
				WorkingChainId: workingChainId,
			})
			if err != nil {
				return fmt.Errorf("failed to resolve: %w", err)
			}

			if res == nil || len(res.Result) == 0 {
				return fmt.Errorf("no result, maybe not configured or expired")
			}

			fmt.Println("Working chain-id:", res.WorkingChainId)

			for query, result := range res.Result {
				fmt.Printf("%s\n", query)
				if result.Error != "" {
					fmt.Printf(" => Error: %s\n", result.Error)
				} else if len(result.Candidates) == 0 {
					fmt.Printf(" => (no result)\n")
				} else {
					fmt.Printf(" => %s\n", strings.Join(result.Candidates, ", "))
				}
			}

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	cmd.Flags().String(flagWorkingChainId, "", "Working Chain ID")

	return cmd
}
