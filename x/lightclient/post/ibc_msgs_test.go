package post_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	ibcsolomachine "github.com/cosmos/ibc-go/v7/modules/light-clients/06-solomachine"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
)

type MockIBCCLientKeeper struct{}

func NewMockIBCClientKeeper() *MockIBCCLientKeeper {
	return &MockIBCCLientKeeper{}
}

func (m *MockIBCCLientKeeper) GetClientConsensusState(ctx sdk.Context, clientID string, height exported.Height) (exported.ConsensusState, bool) {
	return nil, false
}

func (m *MockIBCCLientKeeper) GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool) {
	switch clientID {
	case "non-tm-client-id":
		clientState := ibcsolomachine.ClientState{}
		return &clientState, true
	case "canon-client-id":
		clientState := ibctm.ClientState{
			ChainId: "rollapp-has-canon-client",
		}
		return &clientState, true
	}
	return nil, false
}

func (m *MockIBCCLientKeeper) GenerateClientIdentifier(ctx sdk.Context, clientType string) string {
	return ""
}
