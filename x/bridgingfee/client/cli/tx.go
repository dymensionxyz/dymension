package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	hyputil "github.com/bcp-innovations/hyperlane-cosmos/util"
	"cosmossdk.io/math"
	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"
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
		CmdCreateBridgingFeeHook(),
		CmdSetBridgingFeeHook(),
		CmdCreateAggregationHook(),
		CmdSetAggregationHook(),
	)

	return cmd
}

func CmdCreateBridgingFeeHook() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-fee-hook [token-id] [inbound-fee] [outbound-fee]",
		Short: "Create a new bridging fee hook",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			tokenId := args[0]
			inboundFee, err := math.LegacyNewDecFromStr(args[1])
			if err != nil {
				return err
			}
			outboundFee, err := math.LegacyNewDecFromStr(args[2])
			if err != nil {
				return err
			}

			msg := &types.MsgCreateBridgingFeeHook{
				Owner: clientCtx.GetFromAddress().String(),
				Fees: []types.HLAssetFee{
					{
						TokenID:     tokenId,
						InboundFee:  inboundFee,
						OutboundFee: outboundFee,
					},
				},
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdSetBridgingFeeHook() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-fee-hook [hook-id] [token-id] [inbound-fee] [outbound-fee]",
		Short: "Update an existing bridging fee hook",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			hookId, err := hyputil.HexAddressFromString(args[0])
			if err != nil {
				return err
			}

			tokenId := args[1]
			inboundFee, err := math.LegacyNewDecFromStr(args[2])
			if err != nil {
				return err
			}
			outboundFee, err := math.LegacyNewDecFromStr(args[3])
			if err != nil {
				return err
			}

			newOwner, _ := cmd.Flags().GetString("new-owner")
			renounce, _ := cmd.Flags().GetBool("renounce")

			msg := &types.MsgSetBridgingFeeHook{
				Id:    hookId,
				Owner: clientCtx.GetFromAddress().String(),
				Fees: []types.HLAssetFee{
					{
						TokenID:     tokenId,
						InboundFee:  inboundFee,
						OutboundFee: outboundFee,
					},
				},
				NewOwner:          newOwner,
				RenounceOwnership: renounce,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("new-owner", "", "New owner address")
	cmd.Flags().Bool("renounce", false, "Renounce ownership")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdCreateAggregationHook() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-aggregation-hook [hook-ids...]",
		Short: "Create a new aggregation hook",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			hookIds := make([]hyputil.HexAddress, len(args))
			for i, idStr := range args {
				hookId, err := hyputil.HexAddressFromString(idStr)
				if err != nil {
					return err
				}
				hookIds[i] = hookId
			}

			msg := &types.MsgCreateAggregationHook{
				Owner:   clientCtx.GetFromAddress().String(),
				HookIds: hookIds,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdSetAggregationHook() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-aggregation-hook [hook-id] [child-hook-ids...]",
		Short: "Update an existing aggregation hook",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			hookId, err := hyputil.HexAddressFromString(args[0])
			if err != nil {
				return err
			}

			childHookIds := make([]hyputil.HexAddress, len(args)-1)
			for i, idStr := range args[1:] {
				childHookId, err := hyputil.HexAddressFromString(idStr)
				if err != nil {
					return err
				}
				childHookIds[i] = childHookId
			}

			newOwner, _ := cmd.Flags().GetString("new-owner")
			renounce, _ := cmd.Flags().GetBool("renounce")

			msg := &types.MsgSetAggregationHook{
				Id:                hookId,
				Owner:             clientCtx.GetFromAddress().String(),
				HookIds:           childHookIds,
				NewOwner:          newOwner,
				RenounceOwnership: renounce,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String("new-owner", "", "New owner address")
	cmd.Flags().Bool("renounce", false, "Renounce ownership")
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}