package keeper

import (
	"testing"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

func TestFoo(t *testing.T) {
	/*
	 ProofHeight: "1205408"
	  acknowledgement: null
	  error: ""
	  original_transfer_target: dym1gurgw3tkfpl8nvt9ypqhtccr3t5elere7aq3x6
	  packet:
	    data: eyJhbW91bnQiOiIyMTQzMDAwMDAwMDAwMDAwMDAwMDAwIiwiZGVub20iOiJhbWFuZCIsIm1lbW8iOiJ7XCJlaWJjXCI6e1wiZmVlXCI6XCIzMjE0NTAwMDAwMDAwMDAwMDAwXCJ9fSIsInJlY2VpdmVyIjoiZHltMWZubXJxdTh0aHdxanNtZ2xtd2pweWY4M2N4eWp0a3Z5MHNkNTk3Iiwic2VuZGVyIjoibWFuZGUxdjdhN216eXh6eTZwd2puaGxnMHZkaGdwa2w0OG14MnZ6M2FzcHMifQ==
	    destination_channel: channel-51
	    destination_port: transfer
	    sequence: "53449"
	    source_channel: channel-0
	    source_port: transfer
	    timeout_height:
	      revision_height: "1"
	      revision_number: "9999"
	    timeout_timestamp: "1730182200000000000"
	  relayer: iO9Wcqyjk5AUJ3oecGw+0bTXAiU=
	  rollapp_id: mande_18071918-1
	  status: PENDING
	  type: ON_RECV
	*/

	x := commontypes.RollappPacketKey(
		commontypes.Status_PENDING,
		"mande_18071918-1",
		1205408,
		commontypes.RollappPacket_ON_RECV,
		"channel-0",
		53449,
	)
	t.Log(commontypes.EncodePacketKey(x))

}
