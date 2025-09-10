package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group otcbuyback queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryAllAuctions(),
		CmdQueryAuction(),
		CmdQueryUserPurchase(),
		CmdQueryAcceptedTokens(),
		CmdQueryAcceptedToken(),
	)

	return cmd
}

// CmdQueryAllAuctions implements the query all auctions command.
func CmdQueryAllAuctions() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auctions",
		Short: "Query all auctions",
		Long:  "Query all auctions with optional filtering to exclude completed ones",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			excludeCompleted, err := cmd.Flags().GetBool("exclude-completed")
			if err != nil {
				return err
			}

			req := &types.QueryAllAuctionsRequest{
				ExcludeCompleted: excludeCompleted,
			}

			res, err := queryClient.AllAuctions(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().Bool("exclude-completed", false, "Exclude completed auctions from results")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdQueryAuction implements the query auction command.
func CmdQueryAuction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auction [auction-id]",
		Short: "Query auction by ID",
		Long:  "Query a specific auction by ID with current discount percentage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			auctionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid auction ID: %v", err)
			}

			req := &types.QueryAuctionRequest{
				Id: auctionId,
			}

			res, err := queryClient.Auction(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdQueryUserPurchase implements the query user purchase command.
func CmdQueryUserPurchase() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user-purchase [auction-id] [user-address]",
		Short: "Query user's purchase and vesting info for an auction",
		Long:  "Query user's vesting plan and claimable amount for a specific auction",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			auctionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid auction ID: %v", err)
			}

			userAddress := args[1]

			req := &types.QueryUserPurchaseRequest{
				AuctionId: auctionId,
				User:      userAddress,
			}

			res, err := queryClient.UserPurchase(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdQueryAcceptedTokens implements the query accepted tokens command.
func CmdQueryAcceptedTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accepted-tokens",
		Short: "Query all accepted tokens",
		Long:  "Query all accepted tokens with their current prices",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			req := &types.QueryAcceptedTokensRequest{}

			res, err := queryClient.AcceptedTokens(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdQueryAcceptedToken implements the query accepted token command.
func CmdQueryAcceptedToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accepted-token [denom]",
		Short: "Query accepted token by denom",
		Long:  "Query specific accepted token data and current spot price by denomination",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			denom := args[0]

			req := &types.QueryAcceptedTokenRequest{
				Denom: denom,
			}

			res, err := queryClient.AcceptedToken(cmd.Context(), req)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
