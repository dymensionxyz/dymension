package v4

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"

	"github.com/dymensionxyz/dymension/v3/app/keepers"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	incentiveskeeper "github.com/dymensionxyz/dymension/v3/x/incentives/keeper"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencerkeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// LoadDeprecatedParamsSubspaces loads the deprecated param subspaces for each module
// used to support the migration from x/params to each module's own store
func LoadDeprecatedParamsSubspaces(keepers *keepers.AppKeepers) {
	for _, subspace := range keepers.ParamsKeeper.GetSubspaces() {
		var keyTable paramstypes.KeyTable
		switch subspace.Name() {
		// Cosmos SDK modules
		case authtypes.ModuleName:
			keyTable = authtypes.ParamKeyTable()
		case banktypes.ModuleName:
			keyTable = banktypes.ParamKeyTable()
		case stakingtypes.ModuleName:
			keyTable = stakingtypes.ParamKeyTable()
		case minttypes.ModuleName:
			keyTable = minttypes.ParamKeyTable()
		case distrtypes.ModuleName:
			keyTable = distrtypes.ParamKeyTable()
		case slashingtypes.ModuleName:
			keyTable = slashingtypes.ParamKeyTable()
		case govtypes.ModuleName:
			keyTable = govv1.ParamKeyTable()
		case crisistypes.ModuleName:
			keyTable = crisistypes.ParamKeyTable()

		// Dymension modules
		case rollapptypes.ModuleName:
			keyTable = rollapptypes.ParamKeyTable()
		case sequencertypes.ModuleName:
			keyTable = sequencertypes.ParamKeyTable()

		// Ethermint  modules
		case evmtypes.ModuleName:
			keyTable = evmtypes.ParamKeyTable()
		case feemarkettypes.ModuleName:
			keyTable = feemarkettypes.ParamKeyTable()
		default:
			continue
		}

		if !subspace.HasKeyTable() {
			subspace.WithKeyTable(keyTable)
		}
	}
}

//nolint:staticcheck
func migrateModuleParams(ctx sdk.Context, keepers *keepers.AppKeepers) {
	// Migrate Tendermint consensus parameters from x/params module to a dedicated x/consensus module.
	baseAppLegacySS := keepers.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable())
	baseapp.MigrateParams(ctx, baseAppLegacySS, &keepers.ConsensusParamsKeeper)
}

func migrateDelayedAckParams(ctx sdk.Context, delayedAckKeeper delayedackkeeper.Keeper) {
	// overwrite params for delayedack module due to added parameters
	params := delayedacktypes.DefaultParams()
	delayedAckKeeper.SetParams(ctx, params)
}

func migrateRollappParams(ctx sdk.Context, rollappkeeper *rollappkeeper.Keeper) {
	// overwrite params for rollapp module due to proto change
	params := rollapptypes.DefaultParams()
	params.DisputePeriodInBlocks = rollappkeeper.DisputePeriodInBlocks(ctx)
	rollappkeeper.SetParams(ctx, params)
}

func migrateSequencerParams(ctx sdk.Context, sequencerKeeper *sequencerkeeper.Keeper) {
	// overwrite params for sequencer module due to proto change
	params := sequencertypes.DefaultParams()
	params.MinBond = sequencerKeeper.GetParams(ctx).MinBond
	sequencerKeeper.SetParams(ctx, params)
}

func migrateIncentivesParams(ctx sdk.Context, ik *incentiveskeeper.Keeper) {
	params := ik.GetParams(ctx)
	defaultParams := incentivestypes.DefaultParams()
	params.CreateGaugeBaseFee = defaultParams.CreateGaugeBaseFee
	params.AddToGaugeBaseFee = defaultParams.AddToGaugeBaseFee
	params.AddDenomFee = defaultParams.AddDenomFee
	ik.SetParams(ctx, params)
}
