package types_test

import (
	"testing"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/common/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

type tcase struct {
	orderID string
	tracker string
	bz      []byte
}

func TestFoo(t *testing.T) {

	var testcases = []tcase{
		{
			orderID: "8c2bfdcb05c9f9519ffb08eff4d517e4c6a2c69f1b6c0bd7c508d2f76218d4e1",
			tracker: `/mande_18071918-1/d�/ON_RECV/channel-0/��`,
			bz: commontypes.RollappPacketKey(
				commontypes.Status_PENDING,
				"mande_18071918-1",
				1205408,
				commontypes.RollappPacket_ON_RECV,
				"channel-0",
				53449,
			),
		},
		{
			orderID: "2ab7c88129c2c5a5b401a43e8423b2eab3ed3e2cf67dc71df8a88e97336b97b7",
			tracker: `/mande_18071918-1/k�/ON_RECV/channel-0/��`,
			bz: commontypes.RollappPacketKey(
				commontypes.Status_PENDING,
				"mande_18071918-1",
				1207252,
				commontypes.RollappPacket_ON_RECV,
				"channel-0",
				53450,
			),
		},
		{
			orderID: "c9c834d23707fa53a55944c98c767e319f51ff610fdbb866b8f7d35088ef543f",
			tracker: `/mande_18071918-1/o/ON_RECV/channel-0/��`,
			bz: commontypes.RollappPacketKey(
				commontypes.Status_PENDING,
				"mande_18071918-1",
				1208081,
				commontypes.RollappPacket_ON_RECV,
				"channel-0",
				53451,
			),
		},
	}

	for _, tc := range testcases {
		t.Log(tc.orderID, "Tracker: ", tc.tracker, "Actual reversed:", string(tc.bz))
	}
}

func TestEncodeDecodePacketKey(t *testing.T) {
	packet := commontypes.RollappPacket{
		RollappId:   "rollapp_1234-1",
		Status:      commontypes.Status_PENDING,
		ProofHeight: 8,
		Packet:      getNewTestPacket(t),
	}

	expectedPK := packet.RollappPacketKey()

	encoded := types.EncodePacketKey(expectedPK)
	decoded, err := types.DecodePacketKey(encoded)
	require.NoError(t, err)

	require.Equal(t, expectedPK, decoded)
}

func getNewTestPacket(t *testing.T) *channeltypes.Packet {
	t.Helper()
	data := &transfertypes.FungibleTokenPacketData{
		Receiver: "testReceiver",
	}
	pd, err := transfertypes.ModuleCdc.MarshalJSON(data)
	require.NoError(t, err)
	return &channeltypes.Packet{
		SourcePort:         "testSourcePort",
		SourceChannel:      "testSourceChannel",
		DestinationPort:    "testDestinationPort",
		DestinationChannel: "testDestinationChannel",
		Data:               pd,
		Sequence:           1,
	}
}
