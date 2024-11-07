package ante_test

import (
	"testing"

	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibcsolomachine "github.com/cosmos/ibc-go/v7/modules/light-clients/06-solomachine"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/ante"
	"github.com/stretchr/testify/require"
)

func TestHandleMsgSubmitMisbehaviour(t *testing.T) {
	keeper, ctx := keepertest.LightClientKeeper(t)
	rollappKeeper := NewMockRollappKeeper(nil, nil)
	testClientStates := map[string]exported.ClientState{
		"non-tm-client-id": &ibcsolomachine.ClientState{},
		keepertest.CanonClientID: &ibctm.ClientState{
			ChainId: keepertest.DefaultRollapp,
		},
	}
	ibcclientKeeper := NewMockIBCClientKeeper(testClientStates)
	ibcchannelKeeper := NewMockIBCChannelKeeper(nil)
	keeper.SetCanonicalClient(ctx, keepertest.DefaultRollapp, keepertest.CanonClientID)
	ibcMsgDecorator := ante.NewIBCMessagesDecorator(*keeper, ibcclientKeeper, ibcchannelKeeper, rollappKeeper)
	testCases := []struct {
		name     string
		inputMsg ibcclienttypes.MsgSubmitMisbehaviour
		err      error
	}{
		{
			name: "Could not unpack light client state as tendermint client state",
			inputMsg: ibcclienttypes.MsgSubmitMisbehaviour{
				ClientId:     "non-tm-client-id",
				Misbehaviour: nil,
			},
			err: nil,
		},
		{
			name: "Client is a known canonical client for a rollapp",
			inputMsg: ibcclienttypes.MsgSubmitMisbehaviour{
				ClientId:     keepertest.CanonClientID,
				Misbehaviour: nil,
			},
			err: gerrc.ErrInvalidArgument,
		},
		{
			name: "Client is not a known canonical client",
			inputMsg: ibcclienttypes.MsgSubmitMisbehaviour{
				ClientId:     "client-id",
				Misbehaviour: nil,
			},
			err: nil,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ibcMsgDecorator.HandleMsgSubmitMisbehaviour(ctx, &tc.inputMsg)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
