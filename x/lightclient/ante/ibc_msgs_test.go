package ante_test

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type MockRollappKeeper struct {
	rollapps   map[string]rollapptypes.Rollapp
	stateInfos map[string]map[uint64]rollapptypes.StateInfo
}

func NewMockRollappKeeper(rollapps map[string]rollapptypes.Rollapp, stateInfos map[string]map[uint64]rollapptypes.StateInfo) *MockRollappKeeper {
	return &MockRollappKeeper{
		rollapps:   rollapps,
		stateInfos: stateInfos,
	}
}

func (m *MockRollappKeeper) GetRollapp(ctx sdk.Context, rollappId string) (val rollapptypes.Rollapp, found bool) {
	val, found = m.rollapps[rollappId]
	return val, found
}

func (m *MockRollappKeeper) SetRollapp(ctx sdk.Context, rollapp rollapptypes.Rollapp) {
	m.rollapps[rollapp.RollappId] = rollapp
}

func (m *MockRollappKeeper) FindStateInfoByHeight(ctx sdk.Context, rollappId string, height uint64) (*rollapptypes.StateInfo, error) {
	stateInfos, found := m.stateInfos[rollappId]
	if !found {
		return nil, rollapptypes.ErrUnknownRollappID
	}
	stateInfo, found := stateInfos[height]
	if !found {
		return nil, rollapptypes.ErrNotFound
	}
	return &stateInfo, nil
}

func (m *MockRollappKeeper) GetStateInfo(ctx sdk.Context, rollappId string, index uint64) (val rollapptypes.StateInfo, found bool) {
	stateInfos, found := m.stateInfos[rollappId]
	if !found {
		return val, false
	}
	val, found = stateInfos[index]
	return val, found
}

func (m *MockRollappKeeper) HandleFraud(ctx sdk.Context, rollappID, clientId string, fraudHeight uint64, seqAddr string) error {
	return nil
}

type MockIBCClientKeeper struct {
	clientStates map[string]exported.ClientState
}

func NewMockIBCClientKeeper(cs map[string]exported.ClientState) *MockIBCClientKeeper {
	return &MockIBCClientKeeper{
		clientStates: cs,
	}
}

func (m *MockIBCClientKeeper) GetClientConsensusState(ctx sdk.Context, clientID string, height exported.Height) (exported.ConsensusState, bool) {
	return nil, false
}

func (m *MockIBCClientKeeper) GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool) {
	val, found := m.clientStates[clientID]
	return val, found
}

func (m *MockIBCClientKeeper) GetAllGenesisClients(ctx sdk.Context) ibcclienttypes.IdentifiedClientStates {
	return nil
}

func (m *MockIBCClientKeeper) ConsensusStateHeights(c context.Context, req *ibcclienttypes.QueryConsensusStateHeightsRequest) (*ibcclienttypes.QueryConsensusStateHeightsResponse, error) {
	return nil, nil
}

type MockIBCChannelKeeper struct {
	channelConnections map[string]ibcconnectiontypes.ConnectionEnd
}

func NewMockIBCChannelKeeper(connections map[string]ibcconnectiontypes.ConnectionEnd) *MockIBCChannelKeeper {
	return &MockIBCChannelKeeper{
		channelConnections: connections,
	}
}

func (m *MockIBCChannelKeeper) GetChannelConnection(ctx sdk.Context, portID, channelID string) (string, exported.ConnectionI, error) {
	if portID == "transfer" {
		return "", m.channelConnections[channelID], nil
	}
	return "", nil, nil
}
