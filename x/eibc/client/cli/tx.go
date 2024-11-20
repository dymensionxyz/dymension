package cli

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

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

			amountInt, ok := sdk.NewIntFromString(amountStr)
			if !ok {
				return fmt.Errorf("invalid amount")
			}

			amount := sdk.IntProto{Int: amountInt}

			oepratorFeeShareStr, err := cmd.Flags().GetString(FlagOperatorFeeShare)
			if err != nil {
				return fmt.Errorf("fulfiller fee part is required")
			}
			operatorFeeShareDec, err := sdk.NewDecFromStr(oepratorFeeShareStr)
			if err != nil {
				return fmt.Errorf("invalid fulfiller fee part: %w", err)
			}
			operatorFeeShare := sdk.DecProto{Dec: operatorFeeShareDec}

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
