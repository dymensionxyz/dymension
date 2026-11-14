package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

// Migrator handles in-place x/agent store migrations.
type Migrator struct {
	k Keeper
}

// NewMigrator returns an x/agent store migrator.
func NewMigrator(k Keeper) Migrator {
	return Migrator{k: k}
}

// Migrate1to2 initializes fields added to the agent params at version 2.
func (m Migrator) Migrate1to2(ctx sdk.Context) error {
	params, err := m.k.GetParams(ctx)
	if err != nil {
		return err
	}
	defaults := types.DefaultParams()
	if params.ValidationRequestFee.Denom == "" {
		params.ValidationRequestFee = defaults.ValidationRequestFee
	}
	if params.ValidationTagMaxBytes == 0 {
		params.ValidationTagMaxBytes = defaults.ValidationTagMaxBytes
	}
	if params.ValidationUriMaxBytes == 0 {
		params.ValidationUriMaxBytes = defaults.ValidationUriMaxBytes
	}
	params.SpendRecipientAllowlistMax = types.DefaultSpendRecipientAllowlistMax
	return m.k.SetParams(ctx, params)
}

// Migrate2to3 initializes fields added to the agent params at version 3.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	params, err := m.k.GetParams(ctx)
	if err != nil {
		return err
	}
	params.ValidationMaxResponsesPerRequest = types.DefaultValidationMaxResponsesPerRequest
	return m.k.SetParams(ctx, params)
}

// Migrate3to4 initializes the liveness window (disabled) and backfills every
// agent's last action height to the upgrade height, so enabling the window
// later never retroactively makes pre-existing agents stale.
func (m Migrator) Migrate3to4(ctx sdk.Context) error {
	params, err := m.k.GetParams(ctx)
	if err != nil {
		return err
	}
	params.AgentLivenessWindowBlocks = types.DefaultAgentLivenessWindowBlocks
	if err := m.k.SetParams(ctx, params); err != nil {
		return err
	}
	iter, err := m.k.agents.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	agents, err := iter.Values()
	if err != nil {
		return err
	}
	for _, a := range agents {
		a.LastActionHeight = ctx.BlockHeight()
		if err := m.k.SetAgent(ctx, a); err != nil {
			return err
		}
	}
	return nil
}
