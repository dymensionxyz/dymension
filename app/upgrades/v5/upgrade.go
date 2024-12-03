package v5

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	grouptypes "github.com/cosmos/cosmos-sdk/x/group"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/dymensionxyz/dymension/v3/app/keepers"
	"github.com/dymensionxyz/dymension/v3/app/upgrades"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	lightclienttypes "github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
)

// playground only

const (
	UpgradeName = "v5"
)

var Upgrade = upgrades.Upgrade{
	Name:          UpgradeName,
	CreateHandler: CreateUpgradeHandler,
	StoreUpgrades: storetypes.StoreUpgrades{
		Added: []string{
			consensustypes.ModuleName,
			crisistypes.ModuleName,
			lightclienttypes.ModuleName,
			sponsorshiptypes.ModuleName,
			dymnstypes.ModuleName,
			irotypes.ModuleName,
			grouptypes.ModuleName,
		},
	},
}

func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	_ upgrades.BaseAppParamManager,
	keepers *keepers.AppKeepers,
) upgradetypes.UpgradeHandler {
	return func(ctx sdk.Context, _ upgradetypes.Plan, fromVM module.VersionMap) (module.VersionMap, error) {
		log := ctx.Logger().With("upgrade", UpgradeName)

		setKeyTables(keepers)

		p := rollapptypes.DefaultParams()
		p.DisputePeriodInBlocks = keepers.RollappKeeper.DisputePeriodInBlocks(ctx)
		p.LivenessSlashBlocks = keepers.RollappKeeper.LivenessSlashBlocks(ctx)
		p.LivenessSlashInterval = keepers.RollappKeeper.LivenessSlashInterval(ctx)
		p.AppRegistrationFee = keepers.RollappKeeper.AppRegistrationFee(ctx)
		p.MinSequencerBondGlobal = rollapptypes.DefaultMinSequencerBondGlobalCoin
		keepers.RollappKeeper.SetParams(ctx, p)

		rollapps := keepers.RollappKeeper.GetAllRollapps(ctx)
		for _, ra := range rollapps {
			ra.MinSequencerBond = sdk.NewCoins(rollapptypes.DefaultMinSequencerBondGlobalCoin)
			if ra.GenesisState.GetDeprecatedBridgeOpen() {
				h, ok := keepers.RollappKeeper.GetLatestHeight(ctx, ra.RollappId)
				if !ok {
					log.Error("latest height for transfer enabled not found")
				} else {
					ra.GenesisState.TransferProofHeight = h
				}
			}
			keepers.RollappKeeper.SetRollapp(ctx, ra)
		}
		return fromVM, nil
	}
}

func setKeyTables(keepers *keepers.AppKeepers) {
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
			continue

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
