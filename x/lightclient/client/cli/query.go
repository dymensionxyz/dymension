package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
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
		Use:     "expected",
		Short:   "Query the expected client state",
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

			clientState, err := clienttypes.UnpackClientState(clientStateRes.ClientState)
			if err != nil {
				return err
			}

			tm, ok := clientState.(*ibctm.ClientState)
			if !ok {
				return fmt.Errorf("expected tendermint client state")
			}

			relevant := types.ExpectedClientFieldMask(*tm)
			return clientCtx.PrintObjectLegacy(relevant)

			// return clientCtx.PrintProto(clientStateRes)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetLightClient() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "light-client [rollapp-id]",
		Short: "Get canonical light client if it exists.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argRollappId := args[0]

			req := &types.QueryGetLightClientRequest{
				RollappId: argRollappId,
			}

			res, err := queryClient.LightClient(context.Background(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	return cmd
}
