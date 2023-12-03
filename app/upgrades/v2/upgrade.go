package rc

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
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
	cdc codec.BinaryCodec,

) upgradetypes.UpgradeHandler {

	// Create a map from client id to status where the statuses will be saved
	clientStatuses := make(map[string]exported.Status)
	return func(ctx sdk.Context, _ upgradetypes.Plan, vm module.VersionMap) (module.VersionMap, error) {

		logger := ctx.Logger().With("upgrade", UpgradeName)

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

		logger.Info("Upgraded all clients Successfully. Length", "length", len(clientStatuses))

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
				return false
			} else {
				panic(fmt.Sprintf("client status not found for clientID: %s", clientID))
			}
		})

		logger.Info("Upgraded all consensus states successfully")

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

func GetStoreUpgrades() *storetypes.StoreUpgrades {
	storeUpgrades := storetypes.StoreUpgrades{
		// Set migrations for all new modules
		Added: []string{"poolmanager", "delayedack", "denommetadata", "gamm", "incentives", "lockup", "streamer", "epochs"},
	}
	return &storeUpgrades
}
