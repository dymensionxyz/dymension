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
	FlagRollappId     = "rollapp-id"
	FlagStateIndex    = "index"
	FlagRollappHeight = "rollapp-height"
)

func CmdListStateInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-rollapp-state-info",
		Short: "list all state_info",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllStateInfoRequest{
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
		Use:   fmt.Sprintf("show-rollapp-state-info [%s] [%s|%s]", FlagRollappId, FlagStateIndex, FlagRollappHeight),
		Short: "shows a state_info by index or by height",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			flagSet := cmd.Flags()
			argRollappId, err := flagSet.GetString(FlagRollappId)
			if err != nil {
				return err
			}
			argIndex, err := flagSet.GetUint64(FlagStateIndex)
			if err != nil {
				return err
			}
			argHeight, err := flagSet.GetUint64(FlagRollappHeight)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			byIndex := argIndex != 0
			byHeight := argHeight != 0
			if byIndex && byHeight {
				return fmt.Errorf("only one flag can be use for %s or %s", FlagStateIndex, FlagRollappHeight)
			}
			if byIndex {
				params := &types.QueryGetStateInfoRequest{RollappId: argRollappId, Index: argIndex}

				res, err := queryClient.StateInfo(context.Background(), params)
				if err != nil {
					return err
				}

				return clientCtx.PrintProto(res)
			} else if byHeight {

				params := &types.QueryGetStateInfoByHeightRequest{

					RollappId: argRollappId,
					Height:    argHeight,
				}

				res, err := queryClient.GetStateInfoByHeight(cmd.Context(), params)
				if err != nil {
					return err
				}
				return clientCtx.PrintProto(res)
			} else {
				return fmt.Errorf("please choose one flag for %s or %s", FlagStateIndex, FlagRollappHeight)
			}
		},
	}

	cmd.Flags().String(FlagRollappId, "", "rollapp-id to query for state-info")
	cmd.Flags().Uint64(FlagStateIndex, 0, "Use a specific state-index to query state-info at")
	cmd.Flags().Uint64(FlagRollappHeight, 0, "Use a specific height of the rollapp to query state-info at")
	cmd.MarkFlagRequired(FlagRollappId)

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
