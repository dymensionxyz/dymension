package keeper

import (
	"testing"

	tmdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

type ChannelKeeperStub struct{}

func (ChannelKeeperStub) LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capabilitytypes.Capability, error) {
	return "", nil, nil
}

func (ChannelKeeperStub) GetChannel(ctx sdk.Context, portID, channelID string) (channeltypes.Channel, bool) {
	return channeltypes.Channel{}, false
}

func (ChannelKeeperStub) GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, exported.ClientState, error) {
	return "", &ibctypes.ClientState{}, nil
}

type ICS4WrapperStub struct{}

func (ICS4WrapperStub) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, sourcePort string, sourceChannel string, timeoutHeight clienttypes.Height, timeoutTimestamp uint64, data []byte) (sequence uint64, err error) {
	return 0, nil
}

func (ICS4WrapperStub) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	return nil
}

func (ICS4WrapperStub) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return "", true
}

type ClientKeeperStub struct{}

func (ClientKeeperStub) GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool) {
	return nil, false
}

func (ClientKeeperStub) GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool) {
	return connectiontypes.ConnectionEnd{}, false
}

type ConnectionKeeperStub struct{}

func (ConnectionKeeperStub) GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool) {
	return nil, false
}

func (ConnectionKeeperStub) GetClientConsensusState(ctx sdk.Context, clientID string, height exported.Height) (exported.ConsensusState, bool) {
	return nil, false
}

func (ConnectionKeeperStub) GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool) {
	return connectiontypes.ConnectionEnd{}, false
}

type RollappKeeperStub struct{}

func (RollappKeeperStub) GetParams(ctx sdk.Context) rollapptypes.Params {
	return rollapptypes.Params{}
}

func (RollappKeeperStub) GetStateInfo(ctx sdk.Context, rollappId string, index uint64) (val rollapptypes.StateInfo, found bool) {
	return rollapptypes.StateInfo{}, false
}

// MustGetStateInfo implements types.RollappKeeper.
func (r RollappKeeperStub) MustGetStateInfo(ctx sdk.Context, rollappId string, index uint64) rollapptypes.StateInfo {
	return rollapptypes.StateInfo{}
}

func (RollappKeeperStub) GetLatestFinalizedStateIndex(ctx sdk.Context, rollappId string) (val rollapptypes.StateInfoIndex, found bool) {
	return rollapptypes.StateInfoIndex{}, false
}

func (RollappKeeperStub) GetAllRollapps(ctx sdk.Context) (list []rollapptypes.Rollapp) {
	return []rollapptypes.Rollapp{}
}

func (r RollappKeeperStub) GetValidTransfer(ctx sdk.Context, packetData []byte, raPortOnHub, raChanOnHub string) (data rollapptypes.TransferData, err error) {
	return rollapptypes.TransferData{}, nil
}

type SequencerKeeperStub struct{}

func (SequencerKeeperStub) GetSequencer(ctx sdk.Context, sequencerAddress string) (val sequencertypes.Sequencer, found bool) {
	return sequencertypes.Sequencer{}, false
}

func DelayedackKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	paramsSubspace := typesparams.NewSubspace(cdc,
		types.Amino,
		storeKey,
		memStoreKey,
		"DelayedackParams",
	)

	k := keeper.NewKeeper(cdc,
		storeKey,
		paramsSubspace,
		RollappKeeperStub{},
		ICS4WrapperStub{},
		ChannelKeeperStub{},
		nil,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ctx
}
