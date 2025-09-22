package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"cosmossdk.io/math"
	hyputil "github.com/bcp-innovations/hyperlane-cosmos/util"
	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"
)

const (
	FlagSetFees           = "hook-fees"
	FlagNewOwner          = "new-owner"
	FlagRenounceOwnership = "renounce-ownership"
	FlagHookIds           = "hook-ids"
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

// CmdCreateBridgingFeeHook returns a CLI command for creating a bridging fee hook
func CmdCreateBridgingFeeHook() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-fee-hook [fees-json...]",
		Short: "Create a new bridging fee hook",
		Args:  cobra.MinimumNArgs(1),
		Long: `Create a new fee hook that charges fees for token transfers across bridges.

Fees should be provided as JSON objects as positional arguments.

Example:
dymd tx bridgingfee create-fee-hook '{"token_id":"0x1234567890abcdef1234567890abcdef12345678","inbound_fee":"0.01","outbound_fee":"0.02"}' --from mykey`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var fees []types.HLAssetFee
			for _, feeJSON := range args {
				var feeInput struct {
					TokenID     string `json:"token_id"`
					InboundFee  string `json:"inbound_fee"`
					OutboundFee string `json:"outbound_fee"`
				}

				if err := json.Unmarshal([]byte(feeJSON), &feeInput); err != nil {
					return fmt.Errorf("failed to parse fee JSON %q: %w", feeJSON, err)
				}

				// Parse token ID as HexAddress
				tokenID, err := hyputil.DecodeHexAddress(feeInput.TokenID)
				if err != nil {
					return fmt.Errorf("invalid token_id %q: %w", feeInput.TokenID, err)
				}

				// Parse inbound fee
				inboundFee, err := math.LegacyNewDecFromStr(feeInput.InboundFee)
				if err != nil {
					return fmt.Errorf("invalid inbound_fee %q: %w", feeInput.InboundFee, err)
				}

				// Parse outbound fee
				outboundFee, err := math.LegacyNewDecFromStr(feeInput.OutboundFee)
				if err != nil {
					return fmt.Errorf("invalid outbound_fee %q: %w", feeInput.OutboundFee, err)
				}

				fee := types.HLAssetFee{
					TokenId:     tokenID,
					InboundFee:  inboundFee,
					OutboundFee: outboundFee,
				}
				fees = append(fees, fee)
			}

			msg := &types.MsgCreateBridgingFeeHook{
				Owner: clientCtx.GetFromAddress().String(),
				Fees:  fees,
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

// CmdSetBridgingFeeHook returns a CLI command for updating a bridging fee hook
func CmdSetBridgingFeeHook() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-fee-hook [hook-id]",
		Short: "Update an existing bridging fee hook",
		Long: `Update the configuration of an existing fee hook, including fees, ownership, or other settings.

Note that old values will be overwritten by new values. All fee objects must be supplied otherwise they will be removed.

Examples:
# Update fees
dymd tx bridgingfee set-fee-hook 0x1234... --hook-fees '{"token_id":"0x1234567890abcdef1234567890abcdef12345678","inbound_fee":"0.01","outbound_fee":"0.02"}' --from mykey

# Transfer ownership
dymd tx bridgingfee set-fee-hook 0x1234... --new-owner dym1newowner... --from mykey

# Renounce ownership
dymd tx bridgingfee set-fee-hook 0x1234... --renounce-ownership --from mykey`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			hookID, err := hyputil.DecodeHexAddress(args[0])
			if err != nil {
				return fmt.Errorf("invalid hook ID %q: %w", args[0], err)
			}

			feesJSONStr, err := cmd.Flags().GetString(FlagSetFees)
			if err != nil {
				return err
			}

			var feesJSON []string
			if feesJSONStr != "" {
				feesJSON = []string{feesJSONStr}
			}

			newOwner, err := cmd.Flags().GetString(FlagNewOwner)
			if err != nil {
				return err
			}

			renounceOwnership, err := cmd.Flags().GetBool(FlagRenounceOwnership)
			if err != nil {
				return err
			}

			var fees []types.HLAssetFee
			for _, feeJSON := range feesJSON {
				var feeInput struct {
					TokenID     string `json:"token_id"`
					InboundFee  string `json:"inbound_fee"`
					OutboundFee string `json:"outbound_fee"`
				}

				if err := json.Unmarshal([]byte(feeJSON), &feeInput); err != nil {
					return fmt.Errorf("failed to parse fee JSON %q: %w", feeJSON, err)
				}

				// Parse token ID as HexAddress
				tokenID, err := hyputil.DecodeHexAddress(feeInput.TokenID)
				if err != nil {
					return fmt.Errorf("invalid token_id %q: %w", feeInput.TokenID, err)
				}

				// Parse inbound fee
				inboundFee, err := math.LegacyNewDecFromStr(feeInput.InboundFee)
				if err != nil {
					return fmt.Errorf("invalid inbound_fee %q: %w", feeInput.InboundFee, err)
				}

				// Parse outbound fee
				outboundFee, err := math.LegacyNewDecFromStr(feeInput.OutboundFee)
				if err != nil {
					return fmt.Errorf("invalid outbound_fee %q: %w", feeInput.OutboundFee, err)
				}

				fee := types.HLAssetFee{
					TokenId:     tokenID,
					InboundFee:  inboundFee,
					OutboundFee: outboundFee,
				}
				fees = append(fees, fee)
			}

			msg := &types.MsgSetBridgingFeeHook{
				Owner:             clientCtx.GetFromAddress().String(),
				Id:                hookID,
				Fees:              fees,
				NewOwner:          newOwner,
				RenounceOwnership: renounceOwnership,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagSetFees, "", "Fee configuration for token (JSON format: {\"token_id\":\"0x...\",\"inbound_fee\":\"0.01\",\"outbound_fee\":\"0.02\"})")
	cmd.Flags().String(FlagNewOwner, "", "Transfer ownership to this address")
	cmd.Flags().Bool(FlagRenounceOwnership, false, "Renounce ownership of the hook")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdCreateAggregationHook returns a CLI command for creating an aggregation hook
func CmdCreateAggregationHook() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-aggregation-hook [hook-ids...]",
		Short: "Create a new aggregation hook",
		Args:  cobra.MinimumNArgs(1),
		Long: `Create a new aggregation hook that combines multiple sub-hooks to execute them sequentially.

Note that old values will be overwritten by new values. All hook IDs must be supplied otherwise they will be removed.

Hook IDs should be provided as positional arguments (comma-separated or space-separated).

Example:
dymd tx bridgingfee create-aggregation-hook 0x1234...,0x5678... --from mykey
dymd tx bridgingfee create-aggregation-hook 0x1234... 0x5678... --from mykey`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var hookIds []hyputil.HexAddress
			for _, arg := range args {
				// Handle both comma-separated and space-separated hook IDs
				hookIdStrs := strings.Split(arg, ",")
				for _, hookIdStr := range hookIdStrs {
					hookIdStr = strings.TrimSpace(hookIdStr)
					if hookIdStr == "" {
						continue
					}
					hookId, err := hyputil.DecodeHexAddress(hookIdStr)
					if err != nil {
						return fmt.Errorf("invalid hook ID %q: %w", hookIdStr, err)
					}
					hookIds = append(hookIds, hookId)
				}
			}

			msg := &types.MsgCreateAggregationHook{
				Owner:   clientCtx.GetFromAddress().String(),
				HookIds: hookIds,
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

// CmdSetAggregationHook returns a CLI command for updating an aggregation hook
func CmdSetAggregationHook() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-aggregation-hook [hook-id]",
		Short: "Update an existing aggregation hook",
		Long: `Update the configuration of an existing aggregation hook, including the list of sub-hooks or ownership settings.

Examples:
# Update hook IDs
dymd tx bridgingfee set-aggregation-hook 0x1234... --hook-ids 0x1234...,0x5678... --from mykey

# Transfer ownership
dymd tx bridgingfee set-aggregation-hook 0x1234... --new-owner dym1newowner... --from mykey

# Renounce ownership
dymd tx bridgingfee set-aggregation-hook 0x1234... --renounce-ownership --from mykey`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			hookID, err := hyputil.DecodeHexAddress(args[0])
			if err != nil {
				return fmt.Errorf("invalid hook ID %q: %w", args[0], err)
			}

			hookIdsStr, err := cmd.Flags().GetString(FlagHookIds)
			if err != nil {
				return err
			}

			newOwner, err := cmd.Flags().GetString(FlagNewOwner)
			if err != nil {
				return err
			}

			renounceOwnership, err := cmd.Flags().GetBool(FlagRenounceOwnership)
			if err != nil {
				return err
			}

			var hookIds []hyputil.HexAddress
			if hookIdsStr != "" {
				for _, hookIdStr := range strings.Split(hookIdsStr, ",") {
					hookId, err := hyputil.DecodeHexAddress(strings.TrimSpace(hookIdStr))
					if err != nil {
						return fmt.Errorf("invalid hook ID %q: %w", hookIdStr, err)
					}
					hookIds = append(hookIds, hookId)
				}
			}

			msg := &types.MsgSetAggregationHook{
				Owner:             clientCtx.GetFromAddress().String(),
				Id:                hookID,
				HookIds:           hookIds,
				NewOwner:          newOwner,
				RenounceOwnership: renounceOwnership,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagHookIds, "", "Comma-separated list of hook IDs to aggregate")
	cmd.Flags().String(FlagNewOwner, "", "Transfer ownership to this address")
	cmd.Flags().Bool(FlagRenounceOwnership, false, "Renounce ownership of the hook")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
