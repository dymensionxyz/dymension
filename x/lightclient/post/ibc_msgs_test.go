package post_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
)

type MockIBCCLientKeeper struct {
	clientStates map[string]exported.ClientState
}

func NewMockIBCClientKeeper(cs map[string]exported.ClientState) *MockIBCCLientKeeper {
	return &MockIBCCLientKeeper{
		clientStates: cs,
	}
}

func (m *MockIBCCLientKeeper) GetClientConsensusState(ctx sdk.Context, clientID string, height exported.Height) (exported.ConsensusState, bool) {
	return nil, false
}

func (m *MockIBCCLientKeeper) GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool) {
	cs, ok := m.clientStates[clientID]
	return cs, ok
}

func (m *MockIBCCLientKeeper) GenerateClientIdentifier(ctx sdk.Context, clientType string) string {
	return ""
}
