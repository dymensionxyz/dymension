package keeper

import (
	"cosmossdk.io/math"
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

// Migrate3to4 persists the per-transfer cap added at version 4 as zero (no
// cap) on existing agents, whose records predate the field and decode it nil.
func (m Migrator) Migrate3to4(ctx sdk.Context) error {
	var agents []types.Agent
	if err := m.k.agents.Walk(ctx, nil, func(_ string, a types.Agent) (bool, error) {
		agents = append(agents, a)
		return false, nil
	}); err != nil {
		return err
	}
	for _, a := range agents {
		a.SpendMaxPerTransfer = math.ZeroInt()
		if err := m.k.SetAgent(ctx, a); err != nil {
			return err
		}
	}
	return nil
}
