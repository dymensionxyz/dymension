package types

import (
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v8/modules/core/exported"

	storetypes "cosmossdk.io/store/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

type SequencerKeeperExpected interface {
	SequencerByDymintAddr(ctx sdk.Context, addr cryptotypes.Address) (sequencertypes.Sequencer, error)
	RealSequencer(ctx sdk.Context, addr string) (sequencertypes.Sequencer, error)
}

type RollappKeeperExpected interface {
	GetRollapp(ctx sdk.Context, rollappId string) (val rollapptypes.Rollapp, found bool)
	GetLatestHeight(ctx sdk.Context, rollappId string) (uint64, bool)
	FindStateInfoByHeight(ctx sdk.Context, rollappId string, height uint64) (*rollapptypes.StateInfo, error)
	GetLatestStateInfo(ctx sdk.Context, rollappId string) (rollapptypes.StateInfo, bool)
	SetRollapp(ctx sdk.Context, rollapp rollapptypes.Rollapp)
	IsFirstHeightOfLatestFork(ctx sdk.Context, rollappId string, revision, height uint64) bool

	GetLatestStateInfoIndex(ctx sdk.Context, rollappId string) (rollapptypes.StateInfoIndex, bool)
	GetStateInfo(ctx sdk.Context, rollappId string, index uint64) (sInfo rollapptypes.StateInfo, found bool)
}

type IBCClientKeeperExpected interface {
	GetClientConsensusState(ctx sdk.Context, clientID string, height exported.Height) (exported.ConsensusState, bool)
	GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool)
	ClientStore(ctx sdk.Context, clientID string) storetypes.KVStore
	IterateConsensusStates(ctx sdk.Context, cb func(clientID string, cs ibcclienttypes.ConsensusStateWithHeight) bool)
}

type IBCChannelKeeperExpected interface {
	GetChannel(ctx sdk.Context, portID, channelID string) (types.Channel, bool)
	GetChannelConnection(ctx sdk.Context, portID, channelID string) (string, exported.ConnectionI, error)
}
