package app

import (
	"slices"

	"cosmossdk.io/x/circuit"
	circuittypes "cosmossdk.io/x/circuit/types"
	"cosmossdk.io/x/evidence"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	feegrantmodule "cosmossdk.io/x/feegrant/module"
	"cosmossdk.io/x/upgrade"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	grouptypes "github.com/cosmos/cosmos-sdk/x/group"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	packetforwardmiddleware "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v8/packetforward/types"
	ratelimit "github.com/cosmos/ibc-apps/modules/rate-limiting/v8"
	ratelimittypes "github.com/cosmos/ibc-apps/modules/rate-limiting/v8/types"
	"github.com/cosmos/ibc-go/modules/capability"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	ibctransfer "github.com/cosmos/ibc-go/v8/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v8/modules/core"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v8/modules/light-clients/07-tendermint"
	"github.com/evmos/ethermint/x/evm"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/evmos/ethermint/x/feemarket"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
	"github.com/osmosis-labs/osmosis/v15/x/epochs"
	epochstypes "github.com/osmosis-labs/osmosis/v15/x/epochs/types"
	"github.com/osmosis-labs/osmosis/v15/x/gamm"
	gammtypes "github.com/osmosis-labs/osmosis/v15/x/gamm/types"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager"
	poolmanagertypes "github.com/osmosis-labs/osmosis/v15/x/poolmanager/types"
	"github.com/osmosis-labs/osmosis/v15/x/txfees"
	txfeestypes "github.com/osmosis-labs/osmosis/v15/x/txfees/types"

	dymnsmodule "github.com/dymensionxyz/dymension/v3/x/dymns"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/dymensionxyz/dymension/v3/x/kas"
	kastypes "github.com/dymensionxyz/dymension/v3/x/kas/types"

	delayedackmodule "github.com/dymensionxyz/dymension/v3/x/delayedack"
	denommetadatamodule "github.com/dymensionxyz/dymension/v3/x/denommetadata"
	eibcmodule "github.com/dymensionxyz/dymension/v3/x/eibc"
	"github.com/dymensionxyz/dymension/v3/x/incentives"
	"github.com/dymensionxyz/dymension/v3/x/lockup"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
	rollappmodule "github.com/dymensionxyz/dymension/v3/x/rollapp"
	sequencermodule "github.com/dymensionxyz/dymension/v3/x/sequencer"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship"
	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
	streamermodule "github.com/dymensionxyz/dymension/v3/x/streamer"

	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	denommetadatamoduletypes "github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	eibcmoduletypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/dymensionxyz/dymension/v3/x/iro"
	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	lightclientmodule "github.com/dymensionxyz/dymension/v3/x/lightclient"
	lightclientmoduletypes "github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollappmoduletypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	streamermoduletypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"

	hypercore "github.com/bcp-innovations/hyperlane-cosmos/x/core"
	hypertypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	hyperwarp "github.com/bcp-innovations/hyperlane-cosmos/x/warp"
	hyperwarptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
)

func (app *App) SetupModules(
	appCodec codec.Codec,
	bApp *baseapp.BaseApp,
	skipGenesisInvariants bool,
) []module.AppModule {
	return []module.AppModule{
		genutil.NewAppModule(
			app.AccountKeeper, app.StakingKeeper, app, app.txConfig,
		),
		auth.NewAppModule(appCodec, app.AccountKeeper, nil, app.GetSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, false),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil, app.GetSubspace(minttypes.ModuleName)),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName), app.interfaceRegistry),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(app.UpgradeKeeper, app.AccountKeeper.AddressCodec()),
		evidence.NewAppModule(app.EvidenceKeeper),
		params.NewAppModule(app.ParamsKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		circuit.NewAppModule(appCodec, app.CircuitBreakerKeeper),

		// IBC modules
		ratelimit.NewAppModule(appCodec, app.RateLimitingKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		packetforwardmiddleware.NewAppModule(app.PacketForwardMiddlewareKeeper, app.GetSubspace(packetforwardtypes.ModuleName)),
		ibctransfer.NewAppModule(app.TransferKeeper),
		ibctm.NewAppModule(),

		rollappmodule.NewAppModule(appCodec, app.RollappKeeper),
		iro.NewAppModule(appCodec, *app.IROKeeper, app.AccountKeeper, app.BankKeeper),
		sequencermodule.NewAppModule(appCodec, app.SequencerKeeper),
		sponsorship.NewAppModule(app.SponsorshipKeeper, app.AccountKeeper, app.BankKeeper, app.IncentivesKeeper, app.StakingKeeper),
		streamermodule.NewAppModule(app.StreamerKeeper, app.AccountKeeper, app.BankKeeper, app.EpochsKeeper),
		delayedackmodule.NewAppModule(appCodec, app.DelayedAckKeeper, app.DelayedAckMiddleware),
		denommetadatamodule.NewAppModule(app.DenomMetadataKeeper, *app.EvmKeeper, app.BankKeeper),
		eibcmodule.NewAppModule(appCodec, app.EIBCKeeper, app.AccountKeeper, app.BankKeeper),
		dymnsmodule.NewAppModule(appCodec, app.DymNSKeeper),
		lightclientmodule.NewAppModule(appCodec, app.LightClientKeeper),

		// Ethermint app modules
		evm.NewAppModule(app.EvmKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(evmtypes.ModuleName)),
		feemarket.NewAppModule(app.FeeMarketKeeper, app.GetSubspace(feemarkettypes.ModuleName)),

		// osmosis modules
		lockup.NewAppModule(*app.LockupKeeper),
		epochs.NewAppModule(*app.EpochsKeeper),
		gamm.NewAppModule(appCodec, *app.GAMMKeeper, app.AccountKeeper, app.BankKeeper),
		poolmanager.NewAppModule(*app.PoolManagerKeeper, app.GAMMKeeper),
		incentives.NewAppModule(*app.IncentivesKeeper, app.AccountKeeper, app.BankKeeper, app.EpochsKeeper),
		txfees.NewAppModule(*app.TxFeesKeeper),

		// Hyperlane modules
		hypercore.NewAppModule(appCodec, &app.HyperCoreKeeper),
		hyperwarp.NewAppModule(appCodec, app.HyperWarpKeeper),
		kas.NewAppModule(appCodec, app.KasKeeper),
	}
}

// ModuleAccountAddrs returns all the app's module account addresses.
func ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	// set false not-blocked addresses
	// exclude the streamer as we want him to be able to get external incentives
	modAccAddrs[authtypes.NewModuleAddress(streamermoduletypes.ModuleName).String()] = false
	modAccAddrs[authtypes.NewModuleAddress(txfeestypes.ModuleName).String()] = false
	modAccAddrs[authtypes.NewModuleAddress(irotypes.ModuleName).String()] = false

	return modAccAddrs
}

// module account permissions
var maccPerms = map[string][]string{
	authtypes.FeeCollectorName:                         nil,
	distrtypes.ModuleName:                              nil,
	minttypes.ModuleName:                               {authtypes.Minter},
	stakingtypes.BondedPoolName:                        {authtypes.Burner, authtypes.Staking},
	stakingtypes.NotBondedPoolName:                     {authtypes.Burner, authtypes.Staking},
	govtypes.ModuleName:                                {authtypes.Burner},
	ibctransfertypes.ModuleName:                        {authtypes.Minter, authtypes.Burner},
	sequencertypes.ModuleName:                          {authtypes.Minter, authtypes.Burner, authtypes.Staking},
	rollappmoduletypes.ModuleName:                      {authtypes.Burner},
	sponsorshiptypes.ModuleName:                        nil,
	streamermoduletypes.ModuleName:                     nil,
	evmtypes.ModuleName:                                {authtypes.Minter, authtypes.Burner}, // used for secure addition and subtraction of balance using module account.
	evmtypes.ModuleVirtualFrontierContractDeployerName: nil,                                  // used for deploying virtual frontier bank contract.
	grouptypes.ModuleName:                              nil,
	gammtypes.ModuleName:                               {authtypes.Minter, authtypes.Burner},
	lockuptypes.ModuleName:                             {authtypes.Minter, authtypes.Burner},
	incentivestypes.ModuleName:                         {authtypes.Minter, authtypes.Burner},
	txfeestypes.ModuleName:                             {authtypes.Burner},
	dymnstypes.ModuleName:                              {authtypes.Minter, authtypes.Burner},
	irotypes.ModuleName:                                {authtypes.Minter, authtypes.Burner},
	hypertypes.ModuleName:                              nil,
	hyperwarptypes.ModuleName:                          {authtypes.Minter, authtypes.Burner},
	kastypes.ModuleName:                                nil,
	ratelimittypes.ModuleName:                          nil,
}

var PreBlockers = []string{
	upgradetypes.ModuleName,
}

// TODO: can be cleaned up. only those modules are needed in BeginBlockers now
var BeginBlockers = []string{
	epochstypes.ModuleName,
	capabilitytypes.ModuleName,
	minttypes.ModuleName,
	distrtypes.ModuleName,
	slashingtypes.ModuleName,
	evidencetypes.ModuleName,
	stakingtypes.ModuleName,
	vestingtypes.ModuleName,
	feemarkettypes.ModuleName,
	evmtypes.ModuleName,
	ibcexported.ModuleName,
	ibctransfertypes.ModuleName,
	packetforwardtypes.ModuleName,
	authtypes.ModuleName,
	authz.ModuleName,
	banktypes.ModuleName,
	govtypes.ModuleName,
	crisistypes.ModuleName,
	genutiltypes.ModuleName,
	feegrant.ModuleName,
	paramstypes.ModuleName,
	rollappmoduletypes.ModuleName,
	sequencertypes.ModuleName,
	sponsorshiptypes.ModuleName,
	streamermoduletypes.ModuleName,
	denommetadatamoduletypes.ModuleName,
	delayedacktypes.ModuleName,
	eibcmoduletypes.ModuleName,
	dymnstypes.ModuleName,
	lockuptypes.ModuleName,
	gammtypes.ModuleName,
	poolmanagertypes.ModuleName,
	incentivestypes.ModuleName,
	txfeestypes.ModuleName,
	consensusparamtypes.ModuleName,
	irotypes.ModuleName,
	lightclientmoduletypes.ModuleName,
	grouptypes.ModuleName,
	hypertypes.ModuleName,
	hyperwarptypes.ModuleName,
	kastypes.ModuleName,
	ratelimittypes.ModuleName,
}

var EndBlockers = []string{
	govtypes.ModuleName,
	stakingtypes.ModuleName,
	capabilitytypes.ModuleName,
	authtypes.ModuleName,
	authz.ModuleName,
	banktypes.ModuleName,
	distrtypes.ModuleName,
	feemarkettypes.ModuleName,
	evmtypes.ModuleName,
	slashingtypes.ModuleName,
	vestingtypes.ModuleName,
	minttypes.ModuleName,
	genutiltypes.ModuleName,
	evidencetypes.ModuleName,
	feegrant.ModuleName,
	paramstypes.ModuleName,
	upgradetypes.ModuleName,
	ibcexported.ModuleName,
	ibctransfertypes.ModuleName,
	packetforwardtypes.ModuleName,
	rollappmoduletypes.ModuleName,
	sequencertypes.ModuleName,
	sponsorshiptypes.ModuleName,
	streamermoduletypes.ModuleName,
	denommetadatamoduletypes.ModuleName,
	delayedacktypes.ModuleName,
	eibcmoduletypes.ModuleName,
	dymnstypes.ModuleName,
	epochstypes.ModuleName,
	lockuptypes.ModuleName,
	gammtypes.ModuleName,
	poolmanagertypes.ModuleName,
	incentivestypes.ModuleName,
	txfeestypes.ModuleName,
	consensusparamtypes.ModuleName,
	irotypes.ModuleName,
	lightclientmoduletypes.ModuleName,
	crisistypes.ModuleName,
	grouptypes.ModuleName,
	hypertypes.ModuleName,
	hyperwarptypes.ModuleName,
	kastypes.ModuleName,
	ratelimittypes.ModuleName,
}

var InitGenesis = []string{
	capabilitytypes.ModuleName,
	authtypes.ModuleName,
	authz.ModuleName,
	banktypes.ModuleName,
	distrtypes.ModuleName,
	stakingtypes.ModuleName,
	vestingtypes.ModuleName,
	slashingtypes.ModuleName,
	feemarkettypes.ModuleName,
	evmtypes.ModuleName,
	govtypes.ModuleName,
	minttypes.ModuleName,
	ibcexported.ModuleName,
	genutiltypes.ModuleName,
	evidencetypes.ModuleName,
	paramstypes.ModuleName,
	upgradetypes.ModuleName,
	ibctransfertypes.ModuleName,
	packetforwardtypes.ModuleName,
	feegrant.ModuleName,
	rollappmoduletypes.ModuleName,
	sequencertypes.ModuleName,
	sponsorshiptypes.ModuleName,
	denommetadatamoduletypes.ModuleName, // must after `x/bank` to trigger hooks
	delayedacktypes.ModuleName,
	eibcmoduletypes.ModuleName,
	dymnstypes.ModuleName,
	epochstypes.ModuleName,
	streamermoduletypes.ModuleName, // must be after x/epochs to fill epoch pointers
	lockuptypes.ModuleName,
	gammtypes.ModuleName,
	poolmanagertypes.ModuleName,
	incentivestypes.ModuleName,
	txfeestypes.ModuleName,
	consensusparamtypes.ModuleName,
	irotypes.ModuleName,
	lightclientmoduletypes.ModuleName,
	crisistypes.ModuleName,
	grouptypes.ModuleName,
	hypertypes.ModuleName,
	hyperwarptypes.ModuleName,
	circuittypes.ModuleName,
	kastypes.ModuleName,
	ratelimittypes.ModuleName,
}

// We have custom migration order to make sure we run txfees first (we need it for iro migrations)
func CustomMigrationOrder(modules []string) []string {
	defaultOrder := module.DefaultMigrationsOrder(modules)

	// run txfees first (we need it for iro migrations)
	txfeesIndex := slices.Index(defaultOrder, txfeestypes.ModuleName)
	defaultOrder = append(defaultOrder[:txfeesIndex], defaultOrder[txfeesIndex+1:]...)
	defaultOrder = append([]string{txfeestypes.ModuleName}, defaultOrder...)

	return defaultOrder
}
