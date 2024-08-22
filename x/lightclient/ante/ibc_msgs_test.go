package ante_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	ibcconnectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
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

func (m *MockIBCClientKeeper) GenerateClientIdentifier(ctx sdk.Context, clientType string) string {
	return "new-canon-client-1"
}

type MockIBCChannelKeeper struct {
	channelConnections map[string]ibcconnectiontypes.ConnectionEnd
}

func NewMockIBCChannelKeeper(connections map[string]ibcconnectiontypes.ConnectionEnd) *MockIBCChannelKeeper {
	return &MockIBCChannelKeeper{
		channelConnections: connections,
	}
}

func (m *MockIBCChannelKeeper) GetChannel(ctx sdk.Context, portID, channelID string) (channel ibcchanneltypes.Channel, found bool) {
	return ibcchanneltypes.Channel{}, false
}

func (m *MockIBCChannelKeeper) GetChannelConnection(ctx sdk.Context, portID, channelID string) (string, exported.ConnectionI, error) {
	if portID == "transfer" {
		return "", m.channelConnections[channelID], nil
	}
	return "", nil, nil
}
