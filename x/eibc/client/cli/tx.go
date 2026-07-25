package cli

import (
	"fmt"
	"strconv"

	math "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
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

	cmd.AddCommand(NewFulfillOrderTxCmd())
	cmd.AddCommand(NewFulfillOrderAuthorizedTxCmd())
	cmd.AddCommand(NewUpdateDemandOrderTxCmd())
	cmd.AddCommand(NewCmdGrantAuthorization())
	cmd.AddCommand(NewCmdTryFulfillOnDemand())
	cmd.AddCommand(NewCmdCreateOnDemandLP())
	cmd.AddCommand(NewCmdDeleteOnDemandLP())
	return cmd
}

func NewFulfillOrderTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fulfill-order [order-id] [expected-fee-amount]",
		Short:   "Fulfill a new eibc order",
		Example: "dymd tx eibc fulfill-order <order-id> <expected-fee-amount>",
		Long: `Fulfill a new eibc order by providing the order ID and the expected fee amount.
		The expected fee amount is the amount of fee that the user expects to pay for fulfilling the order.
		`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			orderId := args[0]
			fee := args[1]

			msg := types.NewMsgFulfillOrder(
				clientCtx.GetFromAddress().String(),
				orderId,
				fee,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

const (
	FlagOperatorFeeAddress = "operator-fee-address"
	FlagRollappId          = "rollapp-id"
	FlagPrice              = "price"
	FlagAmount             = "amount"
	FlagValidUntilHeight   = "valid-until-height"
	FlagRateLimitAmount    = "rate-limit-amount"
	FlagRateLimitBlocks    = "rate-limit-blocks"
	FlagMinFeeAbsolute     = "min-fee-absolute"
)

func NewFulfillOrderAuthorizedTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fulfill-order-authorized [order-id] [expected-fee-amount]",
		Short:   "Fulfill a new eibc order with authorization",
		Example: "dymd tx eibc fulfill-order-authorized <order-id> <expected-fee-amount>",
		Long: `Fulfill a new eibc order by providing the order ID and the expected fee amount.
		The expected fee amount is the amount of fee that the user expects to pay for fulfilling the order.
		`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			orderId := args[0]
			fee := args[1]

			rollappId, err := cmd.Flags().GetString(FlagRollappId)
			if err != nil {
				return fmt.Errorf("rollapp ID is required")
			}

			operatorFeeAddress, err := cmd.Flags().GetString(FlagOperatorFeeAddress)
			if err != nil {
				return fmt.Errorf("operator fee address is required")
			}

			priceStr, err := cmd.Flags().GetString(FlagPrice)
			if err != nil {
				return fmt.Errorf("price is required")
			}

			price, err := sdk.ParseCoinsNormalized(priceStr)
			if err != nil {
				return fmt.Errorf("invalid price: %w", err)
			}

			amountStr, err := cmd.Flags().GetString(FlagAmount)
			if err != nil {
				return fmt.Errorf("amount is required")
			}

			amountInt, ok := math.NewIntFromString(amountStr)
			if !ok {
				return fmt.Errorf("invalid amount")
			}

			amount := amountInt

			oepratorFeeShareStr, err := cmd.Flags().GetString(FlagOperatorFeeShare)
			if err != nil {
				return fmt.Errorf("fulfiller fee part is required")
			}
			operatorFeeShareDec, err := math.LegacyNewDecFromStr(oepratorFeeShareStr)
			if err != nil {
				return fmt.Errorf("invalid fulfiller fee part: %w", err)
			}
			operatorFeeShare := operatorFeeShareDec

			settlementValidated, err := cmd.Flags().GetBool(FlagSettlementValidated)
			if err != nil {
				return fmt.Errorf("settlement validated flag")
			}

			msg := types.NewMsgFulfillOrderAuthorized(
				orderId,
				rollappId,
				clientCtx.GetFromAddress().String(),
				operatorFeeAddress,
				fee,
				price,
				amount,
				operatorFeeShare,
				settlementValidated,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().Bool(FlagSettlementValidated, false, "Settlement validated flag")
	cmd.Flags().String(FlagRollappId, "", "Rollapp ID")
	cmd.Flags().String(FlagPrice, "", "Price")
	cmd.Flags().String(FlagAmount, "", "Amount")
	cmd.Flags().String(FlagOperatorFeeShare, "", "Operator fee share")
	cmd.Flags().String(FlagOperatorFeeAddress, "", "Operator fee address")
	return cmd
}

func NewUpdateDemandOrderTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-demand-order [order-id] [new-fee-amount]",
		Short:   "Update a demand order",
		Example: "dymd tx eibc update-demand-order <order-id> <new-fee-amount>",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			orderId := args[0]
			newFee := args[1]

			msg := types.NewMsgUpdateDemandOrder(
				clientCtx.GetFromAddress().String(),
				orderId,
				newFee,
			)

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewCmdTryFulfillOnDemand() *cobra.Command {
	short := "Try to find a fulfiller for a given order and fulfill on the spot"
	cmd := &cobra.Command{
		Use:   "try-fulfill-on-demand [order-id] [rng]",
		Short: short,
		Long:  short + " Can provide rng to avoid choosing same fulfiller multiple times (number). ",

		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			orderId := args[0]
			rng := 0
			if len(args) > 1 {
				rng, err = strconv.Atoi(args[1])
				if err != nil {
					return err
				}
			}

			msg := &types.MsgTryFulfillOnDemand{
				Signer:  clientCtx.GetFromAddress().String(),
				OrderId: orderId,
				Rng:     int64(rng),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func NewCmdCreateOnDemandLP() *cobra.Command {
	short := "Create on demand lp - FUNDS AT RISK - use with caution"
	long := short + "Create on demand lp - anyone can fill an order through your lp with your funds"
	cmd := &cobra.Command{
		Use:     "create-demand-lp [rollapp] [denom] [max-price] [min-fee] [spend-limit] [order-min-age-blocks]",
		Short:   short,
		Long:    long,
		Example: "dymd tx eibc create-demand-lp rollapp1 foo 1000 0.005 500 100",

		Args: cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			rollapp := args[0]
			denom := args[1]

			maxPrice, ok := math.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("invalid max price")
			}

			minFee, err := math.LegacyNewDecFromStr(args[3])
			if err != nil {
				return fmt.Errorf("invalid min fee: %w", err)
			}

			spendLimit, ok := math.NewIntFromString(args[4])
			if !ok {
				return fmt.Errorf("invalid spend limit")
			}

			orderMinAgeBlocks, err := strconv.ParseUint(args[5], 10, 64)
			if err != nil {
				return fmt.Errorf("invalid order min age blocks: %w", err)
			}

			validUntilHeight, err := cmd.Flags().GetUint64(FlagValidUntilHeight)
			if err != nil {
				return err
			}

			rateLimitBlocks, err := cmd.Flags().GetUint64(FlagRateLimitBlocks)
			if err != nil {
				return err
			}

			rateLimitAmount := math.ZeroInt()
			rateLimitAmountStr, err := cmd.Flags().GetString(FlagRateLimitAmount)
			if err != nil {
				return err
			}
			if rateLimitAmountStr != "" {
				rateLimitAmount, ok = math.NewIntFromString(rateLimitAmountStr)
				if !ok {
					return fmt.Errorf("invalid rate limit amount")
				}
			}

			minFeeAbsoluteStr, err := cmd.Flags().GetString(FlagMinFeeAbsolute)
			if err != nil {
				return err
			}
			minFeeAbsolute, ok := math.NewIntFromString(minFeeAbsoluteStr)
			if !ok {
				return fmt.Errorf("invalid min fee absolute")
			}

			msg := &types.MsgCreateOnDemandLP{
				Lp: &types.OnDemandLP{
					FundsAddr:         clientCtx.GetFromAddress().String(),
					Rollapp:           rollapp,
					Denom:             denom,
					MaxPrice:          maxPrice,
					MinFee:            minFee,
					SpendLimit:        spendLimit,
					OrderMinAgeBlocks: orderMinAgeBlocks,
					ValidUntilHeight:  validUntilHeight,
					RateLimitAmount:   rateLimitAmount,
					RateLimitBlocks:   rateLimitBlocks,
					MinFeeAbsolute:    minFeeAbsolute,
				},
				Signer: clientCtx.GetFromAddress().String(),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().Uint64(FlagValidUntilHeight, 0, "LP stops accepting orders at heights >= this value (0 = no expiry)")
	cmd.Flags().String(FlagRateLimitAmount, "", "Max amount spendable within one rate window (empty = disabled)")
	cmd.Flags().Uint64(FlagRateLimitBlocks, 0, "Length in blocks of the rate window")
	cmd.Flags().String(FlagMinFeeAbsolute, "0", "Absolute minimum fee (in the LP denom) an order must offer (0 = disabled)")

	return cmd
}

func NewCmdDeleteOnDemandLP() *cobra.Command {
	short := "Delete on demand lp"
	cmd := &cobra.Command{
		Use:   "delete-demand-lp [id]",
		Short: short,

		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			parse, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return err
			}

			msg := &types.MsgDeleteOnDemandLP{
				Signer: clientCtx.GetFromAddress().String(),
				Ids:    []uint64{parse},
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
