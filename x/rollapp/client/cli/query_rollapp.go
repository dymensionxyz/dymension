package cli

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/gogo/protobuf/proto"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func CmdListRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Query all rollapps currently registered in the hub",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllRollappRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.RollappAll(context.Background(), params)
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

func CmdShowRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [rollapp-id] [by-alias]",
		Short:   "Query the rollapp associated with the specified rollapp-id or alias",
		Args:    cobra.ExactArgs(1),
		Example: "dymd query rollapp show ROLLAPP_CHAIN_ID, dymd query rollapp show ROLLAPP_ALIAS --by-alias",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)
			argRollapp := args[0]

			var (
				res proto.Message
				err error
			)
			if byAlias, _ := cmd.Flags().GetBool(FlagByAlias); byAlias {
				res, err = rollappByAlias(cmd.Context(), queryClient, argRollapp)
			} else {
				res, err = rollappByID(cmd.Context(), queryClient, argRollapp)
			}
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().BoolP(FlagByAlias, "a", false, "Query the rollapp by alias")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func rollappByID(ctx context.Context, queryClient types.QueryClient, rollappID string) (*types.QueryGetRollappResponse, error) {
	params := &types.QueryGetRollappRequest{
		RollappId: rollappID,
	}

	res, err := queryClient.Rollapp(ctx, params)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func rollappByAlias(ctx context.Context, queryClient types.QueryClient, alias string) (*types.QueryGetRollappResponse, error) {
	params := &types.QueryGetRollappByAliasRequest{
		Alias: alias,
	}

	res, err := queryClient.RollappByAlias(ctx, params)
	if err != nil {
		return nil, err
	}
	return res, nil
}
