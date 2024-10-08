package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

const (
	FlagStateIndex    = "index"
	FlagRollappHeight = "rollapp-height"
	FlagFinalized     = "finalized"
)

func CmdShowStateInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state [rollapp-id]",
		Short: "Query the state associated with the specified rollapp-id and any other flags. If no flags are provided, return the latest state.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			argRollappId := args[0]

			flagSet := cmd.Flags()
			argIndex, err := flagSet.GetUint64(FlagStateIndex)
			if err != nil {
				return err
			}
			argHeight, err := flagSet.GetUint64(FlagRollappHeight)
			if err != nil {
				return err
			}
			argFinalized, err := flagSet.GetBool(FlagFinalized)
			if err != nil {
				return err
			}

			if (argHeight != 0 && argIndex != 0) || (argHeight != 0 && argFinalized) || (argIndex != 0 && argFinalized) {
				return status.Error(codes.InvalidArgument, fmt.Sprintf("only one flag can be use for %s, %s or %s", FlagStateIndex, FlagRollappHeight, FlagFinalized))
			}

			params := &types.QueryGetStateInfoRequest{
				RollappId: argRollappId,
				Index:     argIndex,
				Height:    argHeight,
				Finalized: argFinalized,
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.StateInfo(cmd.Context(), params)
			if err != nil {
				return fmt.Errorf("state info: %w", err)
			}
			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Uint64(FlagStateIndex, 0, "Use a specific state-index to query state-info at")
	cmd.Flags().Uint64(FlagRollappHeight, 0, "Use a specific height of the rollapp to query state-info at")
	cmd.Flags().Bool(FlagFinalized, false, "Indicates whether to return the latest finalized state")

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
