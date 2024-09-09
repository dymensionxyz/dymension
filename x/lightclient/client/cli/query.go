package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/version"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group rollapp queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdGetExpectedClientState(),
		CmdGetLightClient(),
	)

	return cmd
}

func CmdGetExpectedClientState() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "expected",
		Short: "Query the expected client state - NOTE: not all returned fields are relevant",
		Long: `Query the expected client state.
Relevant fields:
	trust level
	trust period
	unbonding period
	max clock drift
	frozen height
	proof specs
	upgrade path
	
The other fields can take any value`,
		Example: fmt.Sprintf("%s query %s expected", version.AppName, ibcexported.ModuleName),
		Args:    cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryExpectedClientStateRequest{}

			clientStateRes, err := queryClient.ExpectedClientState(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(clientStateRes)
		},
	}

	return cmd
}

func CmdGetLightClient() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "light-client [rollapp-id]",
		Short: "Get canonical light client if it exists.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			argRollappId := args[0]

			req := &types.QueryGetLightClientRequest{
				RollappId: argRollappId,
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.LightClient(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}
