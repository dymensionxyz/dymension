package keeper_test

import (
	"testing"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

const (
	delayedAckEventType = "delayedack"
	// valid constatns used for testing
	portid   = "testportid"
	chanid   = "channel-0"
	cpportid = "testcpport"
	cpchanid = "testcpchannel"
)

var (
	// Ibc senders and recievers
	ibcSenderAddr   = apptesting.CreateRandomAccounts(1)[0]
	ibcReceiverAddr = apptesting.CreateRandomAccounts(1)[0]
	// Rollapp Packet data
	height             = clienttypes.NewHeight(0, 1)
	timeoutHeight      = clienttypes.NewHeight(0, 100)
	timeoutTimestamp   = uint64(100)
	disabledTimeout    = clienttypes.ZeroHeight()
	transferPacketData = transfertypes.NewFungibleTokenPacketData(
		sdk.DefaultBondDenom,
		"100",
		ibcSenderAddr.String(),
		ibcReceiverAddr.String(),
		"",
	)
	packet        = channeltypes.NewPacket(transferPacketData.GetBytes(), 1, portid, chanid, cpportid, cpchanid, timeoutHeight, timeoutTimestamp)
	rollappPacket = &commontypes.RollappPacket{
		RollappId: "testRollappId",
		Status:    commontypes.Status_PENDING,
		Type:      commontypes.RollappPacket_ON_RECV,
		Packet:    &packet,
	}
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	app := apptesting.Setup(suite.T(), false)
	ctx := app.GetBaseApp().NewContext(false, tmproto.Header{})

	suite.App = app
	suite.Ctx = ctx
}
