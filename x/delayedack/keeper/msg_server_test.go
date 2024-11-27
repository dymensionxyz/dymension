package keeper

import (
	"fmt"
	"testing"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

func TestFoo(t *testing.T) {
	key := `\0\x01/mande_18071918-1/\0\0\0\0\0\x12k�/ON_RECV/channel-0/\0\0\0\0\0\0��`
	packetKey, err := commontypes.DecodePacketKey(key)
	if err != nil {
		panic(fmt.Errorf("failed to decode base64 packet key: %w", err))
	}
	t.Log(packetKey)

	/*
		status Status,
		rollappID string,
		proofHeight uint64,
		packetType RollappPacket_Type,
		packetSrcChannel string,
		packetSequence uint64,
	*/

}
