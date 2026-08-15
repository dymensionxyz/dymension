package cli

import (
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/spf13/cobra"
)

func CmdQueryValidationRequest() *cobra.Command {
	return validationHashQuery("validation-request [request-hash]", func(c types.QueryClient, cmd *cobra.Command, hash []byte) (any, error) {
		return c.ValidationRequest(cmd.Context(), &types.QueryValidationRequestRequest{RequestHash: hash})
	})
}

func CmdQueryValidationResponses() *cobra.Command {
	cmd := validationHashQuery("validation-responses [request-hash]", func(c types.QueryClient, cmd *cobra.Command, hash []byte) (any, error) {
		page, err := client.ReadPageRequest(cmd.Flags())
		if err != nil {
			return nil, err
		}
		return c.ValidationResponses(cmd.Context(), &types.QueryValidationResponsesRequest{RequestHash: hash, Pagination: page})
	})
	flags.AddPaginationFlagsToCmd(cmd, "validation responses")
	return cmd
}

func CmdQueryValidationRequestsByAgent() *cobra.Command {
	cmd := &cobra.Command{Use: "validation-requests-by-agent [agent-id]", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}
		page, err := client.ReadPageRequest(cmd.Flags())
		if err != nil {
			return err
		}
		res, err := types.NewQueryClient(ctx).ValidationRequestsByAgent(cmd.Context(), &types.QueryValidationRequestsByAgentRequest{AgentId: args[0], Pagination: page})
		if err != nil {
			return err
		}
		return ctx.PrintProto(res)
	}}
	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "validation requests")
	return cmd
}

func validationHashQuery(use string, query func(types.QueryClient, *cobra.Command, []byte) (any, error)) *cobra.Command {
	cmd := &cobra.Command{Use: use, Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		hash, err := hex.DecodeString(args[0])
		if err != nil {
			return err
		}
		ctx, err := client.GetClientQueryContext(cmd)
		if err != nil {
			return err
		}
		res, err := query(types.NewQueryClient(ctx), cmd, hash)
		if err != nil {
			return err
		}
		return ctx.PrintObjectLegacy(res)
	}}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
