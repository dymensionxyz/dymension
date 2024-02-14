package keeper_test

import (
	"testing"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/eibc/keeper"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

const (
	eibcEventType = "eibc"
	// valid constatns used for testing
	portid   = "testportid"
	chanid   = "channel-0"
	cpportid = "testcpport"
	cpchanid = "testcpchannel"
)

var (
	// Supply address
	eibcSenderAddr   = apptesting.CreateRandomAccounts(1)[0]
	eibcReceiverAddr = apptesting.CreateRandomAccounts(1)[0]
	// Rollapp Packet data
	height             = clienttypes.NewHeight(0, 1)
	timeoutHeight      = clienttypes.NewHeight(0, 100)
	timeoutTimestamp   = uint64(100)
	disabledTimeout    = clienttypes.ZeroHeight()
	transferPacketData = transfertypes.NewFungibleTokenPacketData(
		sdk.DefaultBondDenom,
		"100",
		eibcSenderAddr.String(),
		eibcReceiverAddr.String(),
		"",
	)
	packet        = channeltypes.NewPacket(transferPacketData.GetBytes(), 1, portid, chanid, cpportid, cpchanid, timeoutHeight, timeoutTimestamp)
	rollappPacket = &commontypes.RollappPacket{
		RollappId: "testRollappId",
		Status:    commontypes.Status_PENDING,
		Type:      commontypes.RollappPacket_ON_RECV,
		Packet:    &packet,
	}
	rollappPacketKey = commontypes.GetRollappPacketKey(
		rollappPacket.RollappId,
		rollappPacket.Status,
		rollappPacket.ProofHeight,
		*rollappPacket.Packet,
	)
)

type KeeperTestSuite struct {
	apptesting.KeeperTestHelper

	msgServer   types.MsgServer
	queryClient types.QueryClient
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	app := apptesting.Setup(suite.T(), false)
	ctx := app.GetBaseApp().NewContext(false, tmproto.Header{})
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.EIBCKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.App = app
	suite.msgServer = keeper.NewMsgServerImpl(app.EIBCKeeper)
	suite.Ctx = ctx
	suite.queryClient = queryClient
}
