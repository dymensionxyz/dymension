package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

var DefaultRelativePacketTimeoutTimestamp = uint64((time.Duration(10) * time.Minute).Nanoseconds())

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(NewFullfilOrderTxCmd())

	return cmd
}

func NewFullfilOrderTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fulfill-order [order-id]",
		Short:   "Fullfil a new eibc order",
		Example: "dymd tx eibc fulfill-order <order-id> <expected-fee-amount>",
		Args:    cobra.ExactArgs(2),
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

func NewUpdateDemandOrderTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update-demand-order [order-id]",
		Short:   "Update a demand order",
		Example: "dymd tx eibc update-demand-order <order-id> <fee-amount>",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			orderId := args[0]
			newFee := args[1]

			msg := types.NewMsgFulfillOrder(
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
