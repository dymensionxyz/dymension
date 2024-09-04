package keepers

import (
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
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward"
	packetforwardmiddleware "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcclientclient "github.com/cosmos/ibc-go/v7/modules/core/02-client/client"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/evmos/ethermint/x/evm"
	evmclient "github.com/evmos/ethermint/x/evm/client"
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
	dymnsmoduleclient "github.com/dymensionxyz/dymension/v3/x/dymns/client"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"

	appparams "github.com/dymensionxyz/dymension/v3/app/params"
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

	"github.com/dymensionxyz/dymension/v3/x/delayedack"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata"
	denommetadatamoduleclient "github.com/dymensionxyz/dymension/v3/x/denommetadata/client"
	denommetadatamoduletypes "github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc"
	eibcmoduletypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	incentivestypes "github.com/dymensionxyz/dymension/v3/x/incentives/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp"

	lightclientmodule "github.com/dymensionxyz/dymension/v3/x/lightclient"
	lightclientmoduletypes "github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollappmoduleclient "github.com/dymensionxyz/dymension/v3/x/rollapp/client"
	rollappmoduletypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer"
	streamermoduleclient "github.com/dymensionxyz/dymension/v3/x/streamer/client"
	streamermoduletypes "github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// ModuleBasics defines the module BasicManager is in charge of setting up basic,
// non-dependant module elements, such as codec registration
// and genesis verification.
var ModuleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	authzmodule.AppModuleBasic{},
	genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
	bank.AppModuleBasic{},
	capability.AppModuleBasic{},
	consensus.AppModuleBasic{},
	staking.AppModuleBasic{},
	mint.AppModuleBasic{},
	distribution.AppModuleBasic{},
	gov.NewAppModuleBasic([]client.ProposalHandler{
		paramsclient.ProposalHandler,
		upgradeclient.LegacyProposalHandler,
		upgradeclient.LegacyCancelProposalHandler,
		ibcclientclient.UpdateClientProposalHandler,
		ibcclientclient.UpgradeProposalHandler,
		streamermoduleclient.CreateStreamHandler,
		streamermoduleclient.TerminateStreamHandler,
		streamermoduleclient.ReplaceStreamHandler,
		streamermoduleclient.UpdateStreamHandler,
		rollappmoduleclient.SubmitFraudHandler,
		denommetadatamoduleclient.CreateDenomMetadataHandler,
		denommetadatamoduleclient.UpdateDenomMetadataHandler,
		dymnsmoduleclient.MigrateChainIdsProposalHandler,
		dymnsmoduleclient.UpdateAliasesProposalHandler,
		evmclient.UpdateVirtualFrontierBankContractProposalHandler,
	}),
	params.AppModuleBasic{},
	crisis.AppModuleBasic{},
	slashing.AppModuleBasic{},
	feegrantmodule.AppModuleBasic{},
	ibc.AppModuleBasic{},
	ibctm.AppModuleBasic{},
	upgrade.AppModuleBasic{},
	evidence.AppModuleBasic{},
	transfer.AppModuleBasic{},
	vesting.AppModuleBasic{},
	rollapp.AppModuleBasic{},
	sponsorship.AppModuleBasic{},
	sequencer.AppModuleBasic{},
	streamer.AppModuleBasic{},
	denommetadata.AppModuleBasic{},
	packetforward.AppModuleBasic{},
	delayedack.AppModuleBasic{},
	eibc.AppModuleBasic{},
	dymnsmodule.AppModuleBasic{},
	lightclientmodule.AppModuleBasic{},

	// Ethermint modules
	evm.AppModuleBasic{},
	feemarket.AppModuleBasic{},

	// Osmosis modules
	lockup.AppModuleBasic{},
	epochs.AppModuleBasic{},
	gamm.AppModuleBasic{},
	poolmanager.AppModuleBasic{},
	incentives.AppModuleBasic{},
	txfees.AppModuleBasic{},
)

func (a *AppKeepers) SetupModules(
	appCodec codec.Codec,
	bApp *baseapp.BaseApp,
	encodingConfig appparams.EncodingConfig,
	skipGenesisInvariants bool,
) []module.AppModule {
	return []module.AppModule{
		genutil.NewAppModule(
			a.AccountKeeper, a.StakingKeeper, bApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, a.AccountKeeper, nil, a.GetSubspace(authtypes.ModuleName)),
		authzmodule.NewAppModule(appCodec, a.AuthzKeeper, a.AccountKeeper, a.BankKeeper, encodingConfig.InterfaceRegistry),
		vesting.NewAppModule(a.AccountKeeper, a.BankKeeper),
		bank.NewAppModule(appCodec, a.BankKeeper, a.AccountKeeper, a.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *a.CapabilityKeeper, false),
		feegrantmodule.NewAppModule(appCodec, a.AccountKeeper, a.BankKeeper, a.FeeGrantKeeper, encodingConfig.InterfaceRegistry),
		crisis.NewAppModule(a.CrisisKeeper, skipGenesisInvariants, a.GetSubspace(crisistypes.ModuleName)),
		consensus.NewAppModule(appCodec, a.ConsensusParamsKeeper),
		gov.NewAppModule(appCodec, a.GovKeeper, a.AccountKeeper, a.BankKeeper, a.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, a.MintKeeper, a.AccountKeeper, nil, a.GetSubspace(minttypes.ModuleName)),
		slashing.NewAppModule(appCodec, a.SlashingKeeper, a.AccountKeeper, a.BankKeeper, a.StakingKeeper, a.GetSubspace(slashingtypes.ModuleName)),
		distr.NewAppModule(appCodec, a.DistrKeeper, a.AccountKeeper, a.BankKeeper, a.StakingKeeper, a.GetSubspace(distrtypes.ModuleName)),
		staking.NewAppModule(appCodec, a.StakingKeeper, a.AccountKeeper, a.BankKeeper, a.GetSubspace(stakingtypes.ModuleName)),
		upgrade.NewAppModule(a.UpgradeKeeper),
		evidence.NewAppModule(a.EvidenceKeeper),
		ibc.NewAppModule(a.IBCKeeper),
		params.NewAppModule(a.ParamsKeeper),
		packetforwardmiddleware.NewAppModule(a.PacketForwardMiddlewareKeeper, a.GetSubspace(packetforwardtypes.ModuleName)),
		ibctransfer.NewAppModule(a.TransferKeeper),
		rollappmodule.NewAppModule(appCodec, a.RollappKeeper, a.AccountKeeper, a.BankKeeper),
		sequencermodule.NewAppModule(appCodec, a.SequencerKeeper, a.AccountKeeper, a.BankKeeper, a.GetSubspace(sequencertypes.ModuleName)),
		sponsorship.NewAppModule(a.SponsorshipKeeper),
		streamermodule.NewAppModule(a.StreamerKeeper, a.AccountKeeper, a.BankKeeper, a.EpochsKeeper),
		delayedackmodule.NewAppModule(appCodec, a.DelayedAckKeeper),
		denommetadatamodule.NewAppModule(a.DenomMetadataKeeper, *a.EvmKeeper, a.BankKeeper),
		eibcmodule.NewAppModule(appCodec, a.EIBCKeeper, a.AccountKeeper, a.BankKeeper),
		dymnsmodule.NewAppModule(appCodec, a.DymNSKeeper),
		lightclientmodule.NewAppModule(appCodec, a.LightClientKeeper),

		// Ethermint app modules
		evm.NewAppModule(a.EvmKeeper, a.AccountKeeper, a.BankKeeper, a.GetSubspace(evmtypes.ModuleName).WithKeyTable(evmtypes.ParamKeyTable())),
		feemarket.NewAppModule(a.FeeMarketKeeper, a.GetSubspace(feemarkettypes.ModuleName).WithKeyTable(feemarkettypes.ParamKeyTable())),

		// osmosis modules
		lockup.NewAppModule(*a.LockupKeeper),
		epochs.NewAppModule(*a.EpochsKeeper),
		gamm.NewAppModule(appCodec, *a.GAMMKeeper, a.AccountKeeper, a.BankKeeper),
		poolmanager.NewAppModule(*a.PoolManagerKeeper, a.GAMMKeeper),
		incentives.NewAppModule(*a.IncentivesKeeper, a.AccountKeeper, a.BankKeeper, a.EpochsKeeper),
		txfees.NewAppModule(*a.TxFeesKeeper),
	}
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (*AppKeepers) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	// exclude the streamer as we want him to be able to get external incentives
	modAccAddrs[authtypes.NewModuleAddress(streamermoduletypes.ModuleName).String()] = false
	modAccAddrs[authtypes.NewModuleAddress(txfeestypes.ModuleName).String()] = false
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
	gammtypes.ModuleName:                               {authtypes.Minter, authtypes.Burner},
	lockuptypes.ModuleName:                             {authtypes.Minter, authtypes.Burner},
	incentivestypes.ModuleName:                         {authtypes.Minter, authtypes.Burner},
	txfeestypes.ModuleName:                             {authtypes.Burner},
	dymnstypes.ModuleName:                              {authtypes.Minter, authtypes.Burner},
}

var BeginBlockers = []string{
	epochstypes.ModuleName,
	upgradetypes.ModuleName,
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
	lightclientmoduletypes.ModuleName,
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
	lightclientmoduletypes.ModuleName,
	crisistypes.ModuleName,
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
	streamermoduletypes.ModuleName,
	denommetadatamoduletypes.ModuleName, // must after `x/bank` to trigger hooks
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
	lightclientmoduletypes.ModuleName,
	crisistypes.ModuleName,
}
