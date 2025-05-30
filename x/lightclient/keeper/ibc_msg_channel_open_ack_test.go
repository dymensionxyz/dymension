package keeper_test

import (
	"testing"

	ibcconnectiontypes "github.com/cosmos/ibc-go/v8/modules/core/03-connection/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	lightclientkeeper "github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/stretchr/testify/require"
)

func TestHandleMsgChannelOpenAck(t *testing.T) {
	keeper, ctx := keepertest.LightClientKeeper(t)
	testRollapps := map[string]rollapptypes.Rollapp{
		keepertest.DefaultRollapp: {
			RollappId: keepertest.DefaultRollapp,
			ChannelId: "channel-on-canon-client",
		},
		"rollapp-no-canon-channel": {
			RollappId: "rollapp-no-canon-channel",
			ChannelId: "",
		},
	}
	testConnections := map[string]ibcconnectiontypes.ConnectionEnd{
		"new-channel-on-canon-client": {
			ClientId: keepertest.CanonClientID,
		},
		"first-channel-on-canon-client": {
			ClientId: "keepertest.CanonClientID-2",
		},
		"non-canon-channel-id": {
			ClientId: "non-keepertest.CanonClientID",
		},
	}
	rollappKeeper := NewMockRollappKeeper(testRollapps, nil)
	ibcclientKeeper := NewMockIBCClientKeeper(nil)
	ibcchannelKeeper := NewMockIBCChannelKeeper(testConnections)
	keeper.SetCanonicalClient(ctx, keepertest.DefaultRollapp, keepertest.CanonClientID)
	keeper.SetCanonicalClient(ctx, "rollapp-no-canon-channel", "keepertest.CanonClientID-2")
	ibcMsgDecorator := lightclientkeeper.NewIBCMessagesDecorator(*keeper, ibcclientKeeper, ibcchannelKeeper, rollappKeeper)
	testCases := []struct {
		name     string
		inputMsg ibcchanneltypes.MsgChannelOpenAck
		err      error
	}{
		{
			name: "port id is not transfer port",
			inputMsg: ibcchanneltypes.MsgChannelOpenAck{
				PortId:    "not-transfer-port",
				ChannelId: "channel-id",
			},
			err: nil,
		},
		{
			name: "channel not on a canonical client",
			inputMsg: ibcchanneltypes.MsgChannelOpenAck{
				PortId:    "transfer",
				ChannelId: "non-canon-channel-id",
			},
			err: nil,
		},
		{
			name: "canonical channel already exists for rollapp",
			inputMsg: ibcchanneltypes.MsgChannelOpenAck{
				PortId:    "transfer",
				ChannelId: "new-channel-on-canon-client",
			},
			err: gerrc.ErrFailedPrecondition,
		},
		{
			name: "canonical channel does not exist",
			inputMsg: ibcchanneltypes.MsgChannelOpenAck{
				PortId:    "transfer",
				ChannelId: "first-channel-on-canon-client",
			},
			err: nil,
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
		})
	}
}
