package keeper

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	connectiontypes "github.com/cosmos/ibc-go/v6/modules/core/03-connection/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	ibctypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	"github.com/dymensionxyz/dymension/x/delayedack/keeper"
	"github.com/dymensionxyz/dymension/x/delayedack/types"
	rollapptypes "github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
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

func (ConnectionKeeperStub) GetConnection(ctx sdk.Context, connectionID string) (connectiontypes.ConnectionEnd, bool) {
	return connectiontypes.ConnectionEnd{}, false
}

type RollappKeeperStub struct{}

func (RollappKeeperStub) GetParams(ctx sdk.Context) rollapptypes.Params {
	return rollapptypes.Params{}
}

func (RollappKeeperStub) GetRollapp(ctx sdk.Context, chainID string) (rollapptypes.Rollapp, bool) {
	return rollapptypes.Rollapp{}, false
}

func (RollappKeeperStub) StateInfo(c context.Context, req *rollapptypes.QueryGetStateInfoRequest) (*rollapptypes.QueryGetStateInfoResponse, error) {
	return nil, nil
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

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,

		RollappKeeperStub{},
		ICS4WrapperStub{},
		ChannelKeeperStub{},
		ClientKeeperStub{},
		ConnectionKeeperStub{},
		nil,
		nil,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, log.NewNopLogger())

	return k, ctx
}
