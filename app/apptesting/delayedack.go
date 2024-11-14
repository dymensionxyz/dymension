package apptesting

import (
	"testing"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

const (
	TestPacketReceiver = "testReceiver"
	TestPacketSender   = "testSender"
)

func GenerateTestPacketData(t *testing.T) []byte {
	t.Helper()
	data := &transfertypes.FungibleTokenPacketData{
		Receiver: TestPacketReceiver,
		Sender:   TestPacketSender,
	}
	pd, err := transfertypes.ModuleCdc.MarshalJSON(data)
	require.NoError(t, err)
	return pd
}

func GenerateTestPacket(t *testing.T, sequence uint64) *channeltypes.Packet {
	t.Helper()
	return &channeltypes.Packet{
		SourcePort:         "testSourcePort",
		SourceChannel:      "testSourceChannel",
		DestinationPort:    "testDestinationPort",
		DestinationChannel: "testDestinationChannel",
		Data:               GenerateTestPacketData(t),
		Sequence:           sequence,
	}
}

func GenerateRollappPackets(t *testing.T, rollappId string, num uint64) []commontypes.RollappPacket {
	t.Helper()
	var packets []commontypes.RollappPacket
	for i := uint64(1); i <= num; i++ {
		packets = append(packets, commontypes.RollappPacket{
			RollappId:   rollappId,
			Packet:      GenerateTestPacket(t, i),
			Status:      commontypes.Status_PENDING,
			ProofHeight: i,
		})
	}
	return packets
}
