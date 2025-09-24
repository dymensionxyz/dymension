package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	hyputil "github.com/bcp-innovations/hyperlane-cosmos/util"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"
	"github.com/spf13/cobra"
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
		Use:   "create-fee-hook [fees-json-array]",
		Short: "Create a new bridging fee hook",
		Args:  cobra.ExactArgs(1),
		Long: `Create a new fee hook that charges fees for token transfers across bridges.

Fees should be provided as a JSON array of fee objects.

Examples:
# Single fee
dymd tx bridgingfee create-fee-hook '[{"token_id":"0x1234567890abcdef1234567890abcdef12345678","inbound_fee":"0.01","outbound_fee":"0.02"}]'

# Multiple fees
dymd tx bridgingfee create-fee-hook '[{"token_id":"0x1234...","inbound_fee":"0.01","outbound_fee":"0.02"},{"token_id":"0x5678...","inbound_fee":"0.05","outbound_fee":"0.03"}]'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var fees []types.HLAssetFee

			if err := json.Unmarshal([]byte(args[0]), &fees); err != nil {
				return fmt.Errorf("parse fees JSON array: %w", err)
			}

			msg := &types.MsgCreateBridgingFeeHook{
				Owner: clientCtx.GetFromAddress().String(),
				Fees:  fees,
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
# Update fees (single fee)
dymd tx bridgingfee set-fee-hook 0x1234... --hook-fees '[{"token_id":"0x1234567890abcdef1234567890abcdef12345678","inbound_fee":"0.01","outbound_fee":"0.02"}]'

# Update fees (multiple fees)
dymd tx bridgingfee set-fee-hook 0x1234... --hook-fees '[{"token_id":"0x1234...","inbound_fee":"0.01","outbound_fee":"0.02"},{"token_id":"0x5678...","inbound_fee":"0.05","outbound_fee":"0.03"}]'

# Transfer ownership
dymd tx bridgingfee set-fee-hook 0x1234... --new-owner dym1newowner...

# Renounce ownership
dymd tx bridgingfee set-fee-hook 0x1234... --renounce-ownership`,
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

			newOwner, err := cmd.Flags().GetString(FlagNewOwner)
			if err != nil {
				return err
			}

			renounceOwnership, err := cmd.Flags().GetBool(FlagRenounceOwnership)
			if err != nil {
				return err
			}

			var fees []types.HLAssetFee
			if feesJSONStr != "" {
				if err := json.Unmarshal([]byte(feesJSONStr), &fees); err != nil {
					return fmt.Errorf("parse fees JSON array: %w", err)
				}
			}

			msg := &types.MsgSetBridgingFeeHook{
				Owner:             clientCtx.GetFromAddress().String(),
				Id:                hookID,
				Fees:              fees,
				NewOwner:          newOwner,
				RenounceOwnership: renounceOwnership,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagSetFees, "", "Fee configuration as JSON array (format: [{\"token_id\":\"0x...\",\"inbound_fee\":\"0.01\",\"outbound_fee\":\"0.02\"}])")
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

Hook IDs should be provided as positional arguments (comma-separated or space-separated).

Example:
dymd tx bridgingfee create-aggregation-hook 0x1234...,0x5678...`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var hookIds []hyputil.HexAddress
			for _, arg := range args {
				// Split by comma first, then handle each piece
				parts := strings.Split(arg, ",")
				for _, part := range parts {
					part = strings.TrimSpace(part)
					if part == "" {
						continue
					}
					hookId, err := hyputil.DecodeHexAddress(part)
					if err != nil {
						return fmt.Errorf("invalid hook ID %q: %w", part, err)
					}
					hookIds = append(hookIds, hookId)
				}
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

// CmdSetAggregationHook returns a CLI command for updating an aggregation hook
func CmdSetAggregationHook() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-aggregation-hook [hook-id]",
		Short: "Update an existing aggregation hook",
		Long: `Update the configuration of an existing aggregation hook, including the list of sub-hooks or ownership settings.

Examples:
# Update hook IDs (multiple)
dymd tx bridgingfee set-aggregation-hook 0x1234... --hook-ids "0x1234567890abcdef1234567890abcdef12345678,0x5678901234567890abcdef1234567890abcdef56"

# Update hook IDs (single)
dymd tx bridgingfee set-aggregation-hook 0x1234... --hook-ids "0x1234567890abcdef1234567890abcdef12345678"

# Transfer ownership
dymd tx bridgingfee set-aggregation-hook 0x1234... --new-owner dym1newowner...

# Renounce ownership
dymd tx bridgingfee set-aggregation-hook 0x1234... --renounce-ownership`,
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
				parts := strings.Split(hookIdsStr, ",")
				for _, part := range parts {
					part = strings.TrimSpace(part)
					if part == "" {
						continue
					}
					hookId, err := hyputil.DecodeHexAddress(part)
					if err != nil {
						return fmt.Errorf("invalid hook ID %q: %w", part, err)
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

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(FlagHookIds, "", "Comma-separated list of hook IDs to aggregate (format: \"0x...,0x...\")")
	cmd.Flags().String(FlagNewOwner, "", "Transfer ownership to this address")
	cmd.Flags().Bool(FlagRenounceOwnership, false, "Renounce ownership of the hook")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
