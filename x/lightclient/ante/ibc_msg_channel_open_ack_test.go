package ante_test

import (
	"testing"

	ibcconnectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/ante"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/stretchr/testify/require"
)

func TestHandleMsgChannelOpenAck(t *testing.T) {
	keeper, ctx := keepertest.LightClientKeeper(t)
	testRollapps := map[string]rollapptypes.Rollapp{
		"rollapp-has-canon-client": {
			RollappId: "rollapp-has-canon-client",
			ChannelId: "channel-on-canon-client",
		},
		"rollapp-no-canon-channel": {
			RollappId: "rollapp-no-canon-channel",
			ChannelId: "",
		},
	}
	testConnections := map[string]ibcconnectiontypes.ConnectionEnd{
		"new-channel-on-canon-client": {
			ClientId: "canon-client-id",
		},
		"first-channel-on-canon-client": {
			ClientId: "canon-client-id-2",
		},
		"non-canon-channel-id": {
			ClientId: "non-canon-client-id",
		},
	}
	rollappKeeper := NewMockRollappKeeper(testRollapps, nil)
	ibcclientKeeper := NewMockIBCClientKeeper(nil)
	ibcchannelKeeper := NewMockIBCChannelKeeper(testConnections)
	keeper.SetCanonicalClient(ctx, "rollapp-has-canon-client", "canon-client-id")
	keeper.SetCanonicalClient(ctx, "rollapp-no-canon-channel", "canon-client-id-2")
	ibcMsgDecorator := ante.NewIBCMessagesDecorator(*keeper, ibcclientKeeper, ibcchannelKeeper, rollappKeeper)
	testCases := []struct {
		name           string
		inputMsg       ibcchanneltypes.MsgChannelOpenAck
		err            error
		canonClientSet bool
	}{
		{
			name: "port id is not transfer port",
			inputMsg: ibcchanneltypes.MsgChannelOpenAck{
				PortId:    "not-transfer-port",
				ChannelId: "channel-id",
			},
			err:            nil,
			canonClientSet: false,
		},
		{
			name: "channel not on a canonical client",
			inputMsg: ibcchanneltypes.MsgChannelOpenAck{
				PortId:    "transfer",
				ChannelId: "non-canon-channel-id",
			},
			err:            nil,
			canonClientSet: false,
		},
		{
			name: "canonical channel already exists for rollapp",
			inputMsg: ibcchanneltypes.MsgChannelOpenAck{
				PortId:    "transfer",
				ChannelId: "new-channel-on-canon-client",
			},
			err:            gerrc.ErrFailedPrecondition,
			canonClientSet: false,
		},
		{
			name: "canonical channel does not exist - set new channel as canonical",
			inputMsg: ibcchanneltypes.MsgChannelOpenAck{
				PortId:    "transfer",
				ChannelId: "first-channel-on-canon-client",
			},
			err:            nil,
			canonClientSet: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := ibcMsgDecorator.HandleMsgChannelOpenAck(ctx, &tc.inputMsg)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
			}
			if tc.canonClientSet {
				rollapp, found := rollappKeeper.GetRollapp(ctx, "rollapp-no-canon-channel")
				require.True(t, found)
				require.Equal(t, tc.inputMsg.ChannelId, rollapp.ChannelId)
			}
		})
	}
}
