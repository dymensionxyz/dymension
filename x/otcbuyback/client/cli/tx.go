package cli

import (
	"fmt"
	"strconv"
	"strings"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/otcbuyback/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdBuy(),
		CmdBuyExactSpend(),
		CmdClaimTokens(),
	)

	return cmd
}

// CmdBuy implements the buy tokens command.
func CmdBuy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buy [auction-id] [amount-to-buy] [denom-to-pay]",
		Short: "Buy a specific amount of tokens in an auction",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Buy a specific amount of tokens in an auction.

Example:
$ %s tx otcbuyback buy 1 1000000000000000000 uusdc --from mykey
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			auctionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid auction ID: %w", err)
			}

			amountToBuy, ok := math.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid amount to buy: %v", args[1])
			}
			denomToPay := args[2]

			msg := types.MsgBuy{
				Buyer:       clientCtx.GetFromAddress().String(),
				AuctionId:   auctionId,
				AmountToBuy: amountToBuy,
				DenomToPay:  denomToPay,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdBuyExactSpend implements the buy exact spend command.
func CmdBuyExactSpend() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buy-exact-spend [auction-id] [payment-coin]",
		Short: "Buy tokens with exact spend amount in an auction",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Buy tokens with an exact spend amount in an auction.

Example:
$ %s tx otcbuyback buy-exact-spend 1 100uusdc --from mykey
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			auctionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid auction ID: %w", err)
			}

			paymentCoin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return fmt.Errorf("invalid payment coin: %w", err)
			}

			msg := types.MsgBuyExactSpend{
				Buyer:       clientCtx.GetFromAddress().String(),
				AuctionId:   auctionId,
				PaymentCoin: paymentCoin,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdClaimTokens implements the claim tokens command.
func CmdClaimTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim [auction-id]",
		Short: "Claim vested tokens from an auction",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claim vested tokens from an auction.

Example:
$ %s tx otcbuyback claim 1 --from mykey
`,
				version.AppName,
			),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			auctionId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid auction ID: %w", err)
			}

			msg := types.MsgClaimTokens{
				Claimer:   clientCtx.GetFromAddress().String(),
				AuctionId: auctionId,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
