package cli

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	iroQueryCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	iroQueryCmd.AddCommand(
		CmdQueryParams(),
		CmdQueryPlans(),
		CmdQueryPlan(),
		CmdQueryPlanByRollapp(),
		CmdQuerySpotPrice(),
		CmdQueryCost(),
		CmdQueryClaimed(),
	)

	return iroQueryCmd
}

func CmdQueryPlans() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plans",
		Short: "Query all IRO plans",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.QueryPlans(cmd.Context(), &types.QueryPlansRequest{})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryPlan() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan [plan-id]",
		Short: "Query a specific IRO plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.QueryPlan(cmd.Context(), &types.QueryPlanRequest{PlanId: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryPlanByRollapp() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plan-by-rollapp [rollapp-id]",
		Short: "Query IRO plan by rollapp ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.QueryPlanByRollapp(cmd.Context(), &types.QueryPlanByRollappRequest{RollappId: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQuerySpotPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price [plan-id]",
		Short: "Query the current price for 1 IRO token for a specific IRO plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.QuerySpotPrice(cmd.Context(), &types.QuerySpotPriceRequest{PlanId: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryCost() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cost [plan-id] [amount]",
		Short: "Query the expected cost for buying or selling a specified amount of tokens",
		Long: `Query the expected cost for buying or selling a specified amount of tokens.
A positive amount indicates a buy operation, while a negative amount indicates a sell operation.`,
		Example: `
  dymd query iro cost plan1 1000000 
  # Query the cost of buying 1000000 tokens from plan1

  dymd query iro cost plan1 -500000
  # Query the cost of selling 500000 tokens from plan1`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			planId := args[0]
			amount, ok := math.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid amount: %s", args[1])
			}

			sell := amount.IsNegative()
			if sell {
				amount = amount.Abs()
			}

			res, err := queryClient.QueryCost(cmd.Context(), &types.QueryCostRequest{
				PlanId: planId,
				Amt:    amount,
				Sell:   sell,
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

func CmdQueryClaimed() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claimed [plan-id]",
		Short: "Query the claimed amount for a specific plan",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.QueryClaimed(cmd.Context(), &types.QueryClaimedRequest{PlanId: args[0]})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
