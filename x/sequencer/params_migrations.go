package sequencer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	seqkeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

type (
	ParamSet = paramtypes.ParamSet

	// Subspace defines an interface that implements the legacy x/params Subspace
	// type.
	//
	// NOTE: This is used solely for migration of x/params managed parameters.
	Subspace interface {
		Get(ctx sdk.Context, key []byte, ptr interface{})
	}
)

// Migrator is a struct for handling in-place store migrations.
type Migrator struct {
	keeper         *seqkeeper.Keeper
	legacySubspace Subspace
}

// NewMigrator returns a new Migrator.
func NewMigrator(keeper *seqkeeper.Keeper, ss Subspace) Migrator {
	return Migrator{keeper: keeper, legacySubspace: ss}
}

// Migrate2to3 migrates from version 2 to 3.
func (m Migrator) Migrate2to3(ctx sdk.Context) error {
	var currParams types.Params
	m.legacySubspace.Get(ctx, types.KeyKickThreshold, &currParams.KickThreshold)
	m.legacySubspace.Get(ctx, types.KeyNoticePeriod, &currParams.NoticePeriod)
	m.legacySubspace.Get(ctx, types.KeyLivenessSlashMinMultiplier, &currParams.LivenessSlashMinMultiplier)
	m.legacySubspace.Get(ctx, types.KeyLivenessSlashMinAbsolute, &currParams.LivenessSlashMinAbsolute)

	currParams.MinBond = types.DefaultMinBond

	if currParams.KickThreshold.Denom == "" {
		currParams.KickThreshold = types.DefaultKickThreshold
	}
	if currParams.NoticePeriod == 0 {
		currParams.NoticePeriod = types.DefaultNoticePeriod
	}
	if currParams.LivenessSlashMinMultiplier.IsNil() || currParams.LivenessSlashMinMultiplier.String() == "" {
		currParams.LivenessSlashMinMultiplier = types.DefaultLivenessSlashMultiplier
	}
	if currParams.LivenessSlashMinAbsolute.Denom == "" {
		currParams.LivenessSlashMinAbsolute = types.DefaultLivenessSlashMinAbsolute
	}

	if err := currParams.ValidateBasic(); err != nil {
		return err
	}

	m.keeper.SetParams(ctx, currParams)
	return nil
}
