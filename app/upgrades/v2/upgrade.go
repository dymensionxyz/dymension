package rc

import (
	"fmt"
	"math/big"

	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingKeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	clientkeeper "github.com/cosmos/ibc-go/v6/modules/core/02-client/keeper"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	dmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/01-dymint/types"
	tmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	rollappKeeper "github.com/dymensionxyz/dymension/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/x/rollapp/types"
)

// CreateUpgradeHandler creates an SDK upgrade handler for v2
func CreateUpgradeHandler(
	mm *module.Manager,
	configurator module.Configurator,
	bankKeeper bankkeeper.Keeper,
	clientKeeper clientkeeper.Keeper,
	rollappKeeper rollappKeeper.Keeper,
	stakingKeeper stakingKeeper.Keeper,
	cdc codec.BinaryCodec,

) upgradetypes.UpgradeHandler {

	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {

		logger := ctx.Logger().With("upgrade", UpgradeName)

		// Create a map from client id to status where the statuses will be saved
		clientStatuses := updateClientStates(ctx, clientKeeper, logger, cdc)

		// Update the consensus states
		updateConsensusStates(ctx, clientKeeper, logger)

		// Verify the client statuses are the same after the upgrade
		verifyClientStatus(ctx, clientKeeper, clientStatuses, logger, cdc)

		// Update delegations
		updateDelegation(ctx, stakingKeeper, logger)

		// Update the denom metadata for DYM token
		bankKeeper.SetDenomMetaData(ctx, DYMTokenMetata)

		// Update the params for rolllapp to init the new param enable_rollapps
		rollappKeeper.SetParams(ctx, rollapptypes.DefaultParams())

		// Iterate over all rollapps and set PermissionedAddresses to empty as the migration
		// Caused the PermissionedAddresses to be set to ""
		rollappsList := rollappKeeper.GetAllRollapp(ctx)
		for _, rollapp := range rollappsList {
			rollapp.PermissionedAddresses = []string{}
			rollappKeeper.SetRollapp(ctx, rollapp)
		}

		// Start running the module migrations
		logger.Debug("running module migrations ...")
		return mm.RunMigrations(ctx, configurator, vm)
	}
}

// updateClientState updates the client state from Dymint client state to Tendermint client state
func updateClientStates(ctx sdk.Context, clientKeeper clientkeeper.Keeper, logger log.Logger, cdc codec.BinaryCodec) map[string]exported.Status {
	// Create a map from client id to status where the statuses will be saved
	clientStatuses := make(map[string]exported.Status)

	// Iterate over all clients and change their client state to Tendermint client state
	clientKeeper.IterateClients(ctx, func(clientID string, clientState exported.ClientState) bool {
		logger.Info("Trying to upgrade client", "clientID", clientID, "clientState", clientState)
		dmClientState, ok := clientState.(*dmtypes.ClientState)
		if !ok {
			logger.Info("Not a Dymint client state. Skipping", "clientID", clientID, "clientState", clientState)
			return false
		}
		logger.Info("Found dymint state. Upgrading client", "clientID", clientID, "clientState", clientState)
		prevStatus := dmClientState.Status(ctx, clientKeeper.ClientStore(ctx, clientID), cdc)
		clientStatuses[clientID] = prevStatus
		newClientState := tmtypes.NewClientState(
			dmClientState.ChainId,
			tmtypes.DefaultTrustLevel,
			dmClientState.TrustingPeriod,
			dmClientState.UnbondingPeriod,
			dmClientState.MaxClockDrift,
			dmClientState.LatestHeight,
			dmClientState.GetProofSpecs(),
			dmClientState.UpgradePath,
			true,
			true,
		)
		clientKeeper.SetClientState(ctx, clientID, newClientState)

		return false
	})

	logger.Info("Upgraded all clients Successfully")
	return clientStatuses
}

func updateConsensusStates(ctx sdk.Context, clientKeeper clientkeeper.Keeper, logger log.Logger) {
	// Iterate over all consensus states and change them to Tendermint consensus state
	clientKeeper.IterateConsensusStates(ctx, func(clientID string, consensusStateWithHeight clienttypes.ConsensusStateWithHeight) bool {
		// Get the consensus state
		logger.Info("Trying to upgrade consensus state", "clientID", clientID)
		exportedConsensusState, found := clientKeeper.GetClientConsensusState(ctx, clientID, consensusStateWithHeight.Height)
		dmConsensusState, ok := exportedConsensusState.(*dmtypes.ConsensusState)
		if !ok {
			logger.Info("Not a Dymint consensus state. Skipping", "clientID", clientID)
			return false
		}
		if !found {
			logger.Info("Consensus state not found. Skipping", "clientID", clientID)
			return false
		}
		// Convert to Tendermint consensus state
		tmConsensusState := &tmtypes.ConsensusState{
			Timestamp:          dmConsensusState.Timestamp,
			Root:               dmConsensusState.Root,
			NextValidatorsHash: dmConsensusState.NextValidatorsHash,
		}

		clientKeeper.SetClientConsensusState(ctx, clientID, consensusStateWithHeight.Height, tmConsensusState)
		logger.Info("Upgraded consensus state", "clientID", clientID, "consensusStateHeight", consensusStateWithHeight.Height)
		return false
	})

	logger.Info("Upgraded all consensus states successfully")

}

func verifyClientStatus(ctx sdk.Context, clientKeeper clientkeeper.Keeper, clientStatuses map[string]exported.Status, logger log.Logger, cdc codec.BinaryCodec) {
	// Iterate over all clients and change their client state to Tendermint client state
	clientKeeper.IterateClients(ctx, func(clientID string, clientState exported.ClientState) bool {
		// Check the status of the client is active after the upgrade
		logger.Info("Checking client status after upgrade", "clientID", clientID)
		status := clientState.Status(ctx, clientKeeper.ClientStore(ctx, clientID), cdc)
		if _, ok := clientStatuses[clientID]; ok {
			if status != clientStatuses[clientID] {
				msg := fmt.Sprintf("client status has changed after upgrade. Expected: %s, got: %s", clientStatuses[clientID], status)
				panic(msg)
			}
		} else {
			fmt.Printf("client status not found for clientID: %s", clientID)
		}
		return false
	})

	logger.Info("Client status verification passed successfully")
}

// updateDelegation updates the delegation for a single node in order to overcome the froopyland chain halt.
// this is a hotfix and should be removed after the chain is upgraded to v2 and enough nodes migrated.
func updateDelegation(ctx sdk.Context, stakingKeeper stakingKeeper.Keeper, logger log.Logger) {

	const (
		// SilkNode validator address
		valBech32Addr = "dymvaloper1kkmputh26f6tyjtdajvwgpjztgcvls7t797jn5"
		// Dymension delegator address
		delBech32Addr = "dym1g3djlajjyqe6lcfz4lphc97csdgnnw249vru73"
	)

	valAddr, valErr := sdk.ValAddressFromBech32(valBech32Addr)
	if valErr != nil {
		logger.Error("error while converting validator address from bech32", "error", valErr)
		panic(valErr)
	}

	validator, found := stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		logger.Error("validator not found")
		panic("validator not found")
	}

	delegatorAddress, err := sdk.AccAddressFromBech32(delBech32Addr)
	if err != nil {
		logger.Error("error while converting delegator address from bech32", "error", err)
		panic(err)
	}

	bondDenom := stakingKeeper.BondDenom(ctx)
	bigIntValue, ok := big.NewInt(0).SetString("27384040000000000000000000", 10)
	if !ok {
		logger.Error("error while converting string to big.Int")
		panic("error while converting string to big.Int")
	}
	amount := sdk.NewCoin(bondDenom, sdk.NewIntFromBigInt(bigIntValue))

	// NOTE: source funds are always unbonded
	_, err = stakingKeeper.Delegate(ctx, delegatorAddress, amount.Amount, stakingtypes.Unbonded, validator, true)
	if err != nil {
		logger.Error("error while delegating", "error", err)
		panic(err)
	}

}

func GetStoreUpgrades() *storetypes.StoreUpgrades {
	storeUpgrades := storetypes.StoreUpgrades{
		// Set migrations for all new modules
		Added: []string{"poolmanager", "delayedack", "denommetadata", "gamm", "incentives", "lockup", "streamer", "epochs", "txfees"},
	}
	return &storeUpgrades
}
