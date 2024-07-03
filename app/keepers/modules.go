package keepers

import (
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	authz "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	client3 "github.com/cosmos/cosmos-sdk/x/distribution/client"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	module2 "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/params"
	client2 "github.com/cosmos/cosmos-sdk/x/params/client"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	client4 "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	"github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v6/packetforward"
	"github.com/cosmos/ibc-go/v6/modules/apps/transfer"
	ibc "github.com/cosmos/ibc-go/v6/modules/core"
	client5 "github.com/cosmos/ibc-go/v6/modules/core/02-client/client"
	"github.com/evmos/ethermint/x/evm"
	client9 "github.com/evmos/ethermint/x/evm/client"
	"github.com/evmos/ethermint/x/feemarket"
	"github.com/osmosis-labs/osmosis/v15/x/epochs"
	"github.com/osmosis-labs/osmosis/v15/x/gamm"
	"github.com/osmosis-labs/osmosis/v15/x/incentives"
	"github.com/osmosis-labs/osmosis/v15/x/lockup"
	"github.com/osmosis-labs/osmosis/v15/x/poolmanager"
	"github.com/osmosis-labs/osmosis/v15/x/txfees"

	"github.com/dymensionxyz/dymension/v3/x/delayedack"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata"
	client8 "github.com/dymensionxyz/dymension/v3/x/denommetadata/client"
	"github.com/dymensionxyz/dymension/v3/x/eibc"
	"github.com/dymensionxyz/dymension/v3/x/rollapp"
	client7 "github.com/dymensionxyz/dymension/v3/x/rollapp/client"
	"github.com/dymensionxyz/dymension/v3/x/sequencer"
	"github.com/dymensionxyz/dymension/v3/x/streamer"
	client6 "github.com/dymensionxyz/dymension/v3/x/streamer/client"
)

var ModuleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	authz.AppModuleBasic{},
	genutil.AppModuleBasic{},
	bank.AppModuleBasic{},
	capability.AppModuleBasic{},
	staking.AppModuleBasic{},
	mint.AppModuleBasic{},
	distribution.AppModuleBasic{},
	gov.NewAppModuleBasic([]client.ProposalHandler{
		client2.ProposalHandler,
		client3.ProposalHandler,
		client4.LegacyProposalHandler,
		client4.LegacyCancelProposalHandler,
		client5.UpdateClientProposalHandler,
		client5.UpgradeProposalHandler,
		client6.CreateStreamHandler,
		client6.TerminateStreamHandler,
		client6.ReplaceStreamHandler,
		client6.UpdateStreamHandler,
		client7.SubmitFraudHandler,
		client8.CreateDenomMetadataHandler,
		client8.UpdateDenomMetadataHandler,
		client9.UpdateVirtualFrontierBankContractProposalHandler,
	}),
	params.AppModuleBasic{},
	crisis.AppModuleBasic{},
	slashing.AppModuleBasic{},
	module2.AppModuleBasic{},
	ibc.AppModuleBasic{},
	upgrade.AppModuleBasic{},
	evidence.AppModuleBasic{},
	transfer.AppModuleBasic{},
	vesting.AppModuleBasic{},
	rollapp.AppModuleBasic{},
	sequencer.AppModuleBasic{},
	streamer.AppModuleBasic{},
	denommetadata.AppModuleBasic{},
	packetforward.AppModuleBasic{},
	delayedack.AppModuleBasic{},
	eibc.AppModuleBasic{},

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
