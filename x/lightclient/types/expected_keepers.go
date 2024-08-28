package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

type SequencerKeeperExpected interface {
	GetSequencer(ctx sdk.Context, sequencerAddress string) (val sequencertypes.Sequencer, found bool)
}

type RollappKeeperExpected interface {
	GetRollapp(ctx sdk.Context, rollappId string) (val rollapptypes.Rollapp, found bool)
	FindStateInfoByHeight(ctx sdk.Context, rollappId string, height uint64) (*rollapptypes.StateInfo, error)
	GetStateInfo(ctx sdk.Context, rollappId string, index uint64) (val rollapptypes.StateInfo, found bool)
	SetRollapp(ctx sdk.Context, rollapp rollapptypes.Rollapp)
	HandleFraud(ctx sdk.Context, rollappID, clientId string, fraudHeight uint64, seqAddr string) error
}

type IBCClientKeeperExpected interface {
	GetClientConsensusState(ctx sdk.Context, clientID string, height exported.Height) (exported.ConsensusState, bool)
	GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool)
	GetAllGenesisClients(ctx sdk.Context) ibcclienttypes.IdentifiedClientStates
}

type IBCChannelKeeperExpected interface {
	GetChannelConnection(ctx sdk.Context, portID, channelID string) (string, exported.ConnectionI, error)
}
