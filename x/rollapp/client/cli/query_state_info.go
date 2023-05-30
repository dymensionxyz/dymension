package cli

import (
	"context"
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/spf13/cobra"
)

const (
	FlagStateIndex    = "index"
	FlagRollappHeight = "rollapp-height"
	FlagFinalized     = "finalized"
)

func CmdListStateInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "states [rollapp-id]",
		Short: "Query all states associated with the specified rollapp-id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			argRollappId := args[0]
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllStateInfoRequest{
				RollappId:  argRollappId,
				Pagination: pageReq,
			}

			res, err := queryClient.StateInfoAll(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdShowStateInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "state [rollapp-id]",
		Short: "Query the state associated with the specified rollapp-id and any other flags. If no flags are provided, return the latest state.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

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
				return fmt.Errorf("only one flag can be use for %s, %s or %s", FlagStateIndex, FlagRollappHeight, FlagFinalized)
			}

			params := &types.QueryGetStateInfoRequest{
				RollappId: argRollappId,
				Index:     argIndex,
				Height:    argHeight,
				Finalized: argFinalized,
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.StateInfo(context.Background(), params)
			if err != nil {
				return err
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
