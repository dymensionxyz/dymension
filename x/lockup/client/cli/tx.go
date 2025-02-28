package cli

import (
	"github.com/osmosis-labs/osmosis/v15/osmoutils/osmocli"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := osmocli.TxIndexCmd(types.ModuleName)
	osmocli.AddTxCmd(cmd, NewLockTokensCmd)
	osmocli.AddTxCmd(cmd, NewBeginUnlockByIDCmd)
	osmocli.AddTxCmd(cmd, NewForceUnlockByIdCmd)

	return cmd
}

func NewLockTokensCmd() (*osmocli.TxCliDesc, *types.MsgLockTokens) {
	return &osmocli.TxCliDesc{
		Use:   "lock-tokens [tokens]",
		Short: "lock tokens into lockup pool from user account",
		CustomFlagOverrides: map[string]string{
			"duration": FlagDuration,
		},
		Flags: osmocli.FlagDesc{RequiredFlags: []*pflag.FlagSet{FlagSetLockTokens()}},
	}, &types.MsgLockTokens{}
}

// NewBeginUnlockByIDCmd unlocks individual period lock by ID.
func NewBeginUnlockByIDCmd() (*osmocli.TxCliDesc, *types.MsgBeginUnlocking) {
	return &osmocli.TxCliDesc{
		Use:   "begin-unlock-by-id [id]",
		Short: "begin unlock individual period lock by ID",
		CustomFlagOverrides: map[string]string{
			"coins": FlagAmount,
		},
		Flags: osmocli.FlagDesc{OptionalFlags: []*pflag.FlagSet{FlagSetUnlockTokens()}},
	}, &types.MsgBeginUnlocking{}
}

// NewForceUnlockByIdCmd force unlocks individual period lock by ID if proper permissions exist.
func NewForceUnlockByIdCmd() (*osmocli.TxCliDesc, *types.MsgForceUnlock) {
	return &osmocli.TxCliDesc{
		Use:   "force-unlock-by-id [id]",
		Short: "force unlocks individual period lock by ID",
		Long:  "force unlocks individual period lock by ID. if no amount provided, entire lock is unlocked",
		CustomFlagOverrides: map[string]string{
			"coins": FlagAmount,
		},
		Flags: osmocli.FlagDesc{OptionalFlags: []*pflag.FlagSet{FlagSetUnlockTokens()}},
	}, &types.MsgForceUnlock{}
}
