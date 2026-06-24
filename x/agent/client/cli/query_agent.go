package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: "Query the agent module parameters",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdShowAgent() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show-agent [id]",
		Short:   "Query a single agent by id",
		Args:    cobra.ExactArgs(1),
		Example: "dymd query agent show-agent my-agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Agent(cmd.Context(), &types.QueryAgentRequest{Id: args[0]})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdListAgents() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-agents",
		Short: "Query all registered agents",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Agents(cmd.Context(), &types.QueryAgentsRequest{Pagination: pageReq})
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

func CmdAgentActions() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agent-actions [id]",
		Short:   "Query the attested action log for an agent (paginated)",
		Args:    cobra.ExactArgs(1),
		Example: "dymd query agent agent-actions my-agent",
		RunE: func(cmd *cobra.Command, args []string) error {
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.AgentActions(cmd.Context(), &types.QueryAgentActionsRequest{
				AgentId:    args[0],
				Pagination: pageReq,
			})
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

func CmdAgentAction() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "agent-action [id] [seq]",
		Short:   "Query a single attested action log entry",
		Args:    cobra.ExactArgs(2),
		Example: "dymd query agent agent-action my-agent 0",
		RunE: func(cmd *cobra.Command, args []string) error {
			seq, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.AgentAction(cmd.Context(), &types.QueryAgentActionRequest{
				AgentId: args[0],
				Seq:     seq,
			})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
