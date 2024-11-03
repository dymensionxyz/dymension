package keeper_test

import (
	"testing"

	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

const (
	delayedAckEventType = "delayedack"
	testSourceChannel   = "testSourceChannel"
)

type DelayedAckTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(DelayedAckTestSuite))
}

func (suite *DelayedAckTestSuite) SetupTest() {
	app := apptesting.Setup(suite.T(), false)
	ctx := app.GetBaseApp().NewContext(false, cometbftproto.Header{})

	suite.App = app
	suite.Ctx = ctx
}

func (suite *DelayedAckTestSuite) FinalizePacket(ctx sdk.Context, p commontypes.RollappPacket) {
	suite.T().Helper()

	handler := suite.App.MsgServiceRouter().Handler(new(types.MsgFinalizePacket))
	resp, err := handler(ctx, &types.MsgFinalizePacket{
		Sender:            apptesting.CreateRandomAccounts(1)[0].String(),
		RollappId:         p.RollappId,
		PacketProofHeight: p.ProofHeight,
		PacketType:        p.Type,
		PacketSrcChannel:  p.Packet.SourceChannel,
		PacketSequence:    p.Packet.Sequence,
	})
	suite.Require().NoError(err)
	suite.Require().NotNil(resp)
}

const testPacketReceiver = "testReceiver"

func (s *DelayedAckTestSuite) GenerateTestPacket(sequence uint64) *channeltypes.Packet {
	return &channeltypes.Packet{
		SourcePort:         "testSourcePort",
		SourceChannel:      testSourceChannel,
		DestinationPort:    "testDestinationPort",
		DestinationChannel: "testDestinationChannel",
		Data:               s.GenerateTestPacketData(),
		Sequence:           sequence,
	}
}

func (s *DelayedAckTestSuite) GenerateTestPacketData() []byte {
	data := &transfertypes.FungibleTokenPacketData{
		Receiver: testPacketReceiver,
	}
	pd, err := transfertypes.ModuleCdc.MarshalJSON(data)
	s.Require().NoError(err)
	return pd
}

func (s *DelayedAckTestSuite) GeneratePackets(rollappId string, num uint64) []commontypes.RollappPacket {
	var packets []commontypes.RollappPacket
	for i := uint64(0); i < num; i++ {
		packets = append(packets, commontypes.RollappPacket{
			RollappId:   rollappId,
			Packet:      s.GenerateTestPacket(i),
			Status:      commontypes.Status_PENDING,
			ProofHeight: i,
		})
	}
	return packets
}
