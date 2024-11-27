package types_test

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/stretchr/testify/require"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

type tcase struct {
	orderID string
	tracker string
	bz      []byte
}

func TestFooKey(t *testing.T) {

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
		idGotRebuild := eibctypes.BuildDemandIDFromPacketKey(string(tc.bz))
		idGotTracker := eibctypes.BuildDemandIDFromPacketKey(string(tc.tracker))
		k := string(tc.bz)
		fmt.Println(tc.orderID, "key", len(k), k)

		t.Log(fmt.Sprintf(`
order id: %s,
tracker: %s,
actual : %s,
order id rebuild : %s,
order id tracker : %s,
`, tc.orderID, tc.tracker, string(tc.bz), idGotRebuild, idGotTracker))
		//t.Log(tc.orderID, "Tracker: ", tc.tracker, "Actual reversed:", string(tc.bz))
	}
	t.FailNow()
}

func decodeTrackingPacketKey(encodedKey string) (string, error) {
	// Replace Unicode escape sequences with their actual characters
	decodedKey := strings.ReplaceAll(encodedKey, "\\u0000", "\x00")
	decodedKey = strings.ReplaceAll(decodedKey, "\\u0001", "\x01")
	decodedKey = strings.ReplaceAll(decodedKey, "\\u0012", "\x12")
	// Add more replacements as needed

	// Decode any remaining hex-encoded characters
	decodedBytes, err := hex.DecodeString(decodedKey)
	if err != nil {
		return "", err
	}

	return string(decodedBytes), nil
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
