package keeper_test

import (
	ibctransfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	damodule "github.com/dymensionxyz/dymension/v3/x/delayedack"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

func (suite *DelayedAckTestSuite) TestHandleFraud() {
	keeper, ctx := suite.App.DelayedAckKeeper, suite.Ctx
	transferStack := damodule.NewIBCMiddleware(
		damodule.WithIBCModule(ibctransfer.NewIBCModule(suite.App.TransferKeeper)),
		damodule.WithKeeper(keeper),
		damodule.WithRollappKeeper(suite.App.RollappKeeper),
	)

	rollappId := "testRollappId"
	pkts := apptesting.GenerateRollappPackets(suite.T(), rollappId, 5)
	rollappId2 := "testRollappId2"
	pkts2 := apptesting.GenerateRollappPackets(suite.T(), rollappId2, 5)
	prefixPending1 := types.ByRollappIDByStatus(rollappId, commontypes.Status_PENDING)
	prefixPending2 := types.ByRollappIDByStatus(rollappId2, commontypes.Status_PENDING)
	prefixReverted := types.ByRollappIDByStatus(rollappId, commontypes.Status_REVERTED)
	prefixFinalized := types.ByRollappIDByStatus(rollappId, commontypes.Status_FINALIZED)
	prefixFinalized2 := types.ByRollappIDByStatus(rollappId, commontypes.Status_FINALIZED)

	for _, pkt := range append(pkts, pkts2...) {
		keeper.SetRollappPacket(ctx, pkt)
	}

	suite.Require().Equal(5, len(keeper.ListRollappPackets(ctx, prefixPending1)))
	suite.Require().Equal(5, len(keeper.ListRollappPackets(ctx, prefixPending2)))

	// finalize some packets
	_, err := keeper.UpdateRollappPacketWithStatus(ctx, pkts[0], commontypes.Status_FINALIZED)
	suite.Require().Nil(err)
	_, err = keeper.UpdateRollappPacketWithStatus(ctx, pkts2[0], commontypes.Status_FINALIZED)
	suite.Require().Nil(err)

	err = keeper.HandleFraud(ctx, rollappId, transferStack)
	suite.Require().Nil(err)

	suite.Require().Equal(0, len(keeper.ListRollappPackets(ctx, prefixPending1)))
	suite.Require().Equal(4, len(keeper.ListRollappPackets(ctx, prefixPending2)))
	suite.Require().Equal(4, len(keeper.ListRollappPackets(ctx, prefixReverted)))
	suite.Require().Equal(1, len(keeper.ListRollappPackets(ctx, prefixFinalized)))
	suite.Require().Equal(1, len(keeper.ListRollappPackets(ctx, prefixFinalized2)))
}

func (suite *DelayedAckTestSuite) TestDeletionOfRevertedPackets() {
	keeper, ctx := suite.App.DelayedAckKeeper, suite.Ctx
	transferStack := damodule.NewIBCMiddleware(
		damodule.WithIBCModule(ibctransfer.NewIBCModule(suite.App.TransferKeeper)),
		damodule.WithKeeper(keeper),
		damodule.WithRollappKeeper(suite.App.RollappKeeper),
	)

	rollappId := "testRollappId"
	pkts := apptesting.GenerateRollappPackets(suite.T(), rollappId, 5)
	rollappId2 := "testRollappId2"
	pkts2 := apptesting.GenerateRollappPackets(suite.T(), rollappId2, 5)

	for _, pkt := range append(pkts, pkts2...) {
		keeper.SetRollappPacket(ctx, pkt)
	}

	err := keeper.HandleFraud(ctx, rollappId, transferStack)
	suite.Require().Nil(err)

	suite.Require().Equal(10, len(keeper.GetAllRollappPackets(ctx)))

	keeper.SetParams(ctx, types.Params{EpochIdentifier: "minute", BridgingFee: keeper.BridgingFee(ctx)})
	epochHooks := keeper.GetEpochHooks()
	err = epochHooks.AfterEpochEnd(ctx, "minute", 1)
	suite.Require().NoError(err)

	suite.Require().Equal(5, len(keeper.GetAllRollappPackets(ctx)))
}

// TODO: test refunds of pending packets
