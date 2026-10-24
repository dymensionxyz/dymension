package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
$ %s tx otcbuyback buy 1 1000000000000000000 uusdc --vesting-period 720h --from mykey
$ %s tx otcbuyback buy 1 1000000000000000000 uusdc --max-cost 260000 --from mykey

The price is derived from the live AMM spot price at execution time. Programmatic
callers should set --max-cost to bound slippage.
`,
				version.AppName,
				version.AppName,
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

			// Parse optional vesting period
			vestingPeriodStr, err := cmd.Flags().GetString("vesting-period")
			if err != nil {
				return err
			}
			var vestingPeriod time.Duration
			if vestingPeriodStr != "" {
				duration, err := time.ParseDuration(vestingPeriodStr)
				if err != nil {
					return fmt.Errorf("invalid vesting period: %w", err)
				}
				vestingPeriod = duration
			}

			maxCost, err := getOptionalIntFlag(cmd, "max-cost")
			if err != nil {
				return err
			}

			msg := types.MsgBuy{
				Buyer:         clientCtx.GetFromAddress().String(),
				AuctionId:     auctionId,
				AmountToBuy:   amountToBuy,
				DenomToPay:    denomToPay,
				VestingPeriod: vestingPeriod,
				MaxCost:       maxCost,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String("vesting-period", "", "Optional vesting period (e.g., '720h' for 30 days). Required for fixed discount auctions, ignored for linear discount auctions")
	cmd.Flags().String("max-cost", "", "Optional maximum payment (in denom-to-pay base units); the tx fails if the execution price exceeds it. Unset disables the check")

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
$ %s tx otcbuyback buy-exact-spend 1 100uusdc --vesting-period 720h --from mykey
$ %s tx otcbuyback buy-exact-spend 1 100uusdc --min-tokens-out 400000000000000 --from mykey

The price is derived from the live AMM spot price at execution time. Programmatic
callers should set --min-tokens-out to bound slippage.
`,
				version.AppName,
				version.AppName,
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

			// Parse optional vesting period
			vestingPeriodStr, err := cmd.Flags().GetString("vesting-period")
			if err != nil {
				return err
			}
			var vestingPeriod time.Duration
			if vestingPeriodStr != "" {
				duration, err := time.ParseDuration(vestingPeriodStr)
				if err != nil {
					return fmt.Errorf("invalid vesting period: %w", err)
				}
				vestingPeriod = duration
			}

			minTokensOut, err := getOptionalIntFlag(cmd, "min-tokens-out")
			if err != nil {
				return err
			}

			msg := types.MsgBuyExactSpend{
				Buyer:         clientCtx.GetFromAddress().String(),
				AuctionId:     auctionId,
				PaymentCoin:   paymentCoin,
				VestingPeriod: vestingPeriod,
				MinTokensOut:  minTokensOut,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String("vesting-period", "", "Optional vesting period (e.g., '720h' for 30 days). Required for fixed discount auctions, ignored for linear discount auctions")
	cmd.Flags().String("min-tokens-out", "", "Optional minimum tokens to receive (base units); the tx fails if the execution price yields fewer. Unset disables the check")

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

func getOptionalIntFlag(cmd *cobra.Command, name string) (math.Int, error) {
	str, err := cmd.Flags().GetString(name)
	if err != nil {
		return math.Int{}, err
	}
	if str == "" {
		return math.ZeroInt(), nil
	}
	v, ok := math.NewIntFromString(str)
	if !ok {
		return math.Int{}, fmt.Errorf("invalid %s: %s", name, str)
	}
	return v, nil
}
