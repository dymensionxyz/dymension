package ante_test

import (
	"testing"

	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/ante"
	"github.com/stretchr/testify/require"
)

func TestHandleMsgSubmitMisbehaviour(t *testing.T) {
	keeper, ctx := keepertest.LightClientKeeper(t)
	rollappKeeper := NewMockRollappKeeper()
	ibcclientKeeper := NewMockIBCClientKeeper()
	ibcchannelKeeper := NewMockIBCChannelKeeper()
	keeper.SetCanonicalClient(ctx, "rollapp-has-canon-client", "canon-client-id")
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
				ClientId:     "canon-client-id",
				Misbehaviour: nil,
			},
			err: ibcclienttypes.ErrInvalidClient,
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
