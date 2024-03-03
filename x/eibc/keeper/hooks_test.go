package keeper_test

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func (suite *KeeperTestSuite) TestAfterRollappPacketUpdated() {
	// Set a rollapp packet
	suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)
	// Create new demand order
	demandOrderFulfillerAddr := apptesting.AddTestAddrs(suite.App, suite.Ctx, 1, math.NewInt(1000))[0].String()
	demandOrder, err := types.NewDemandOrder(*rollappPacket, "100", "50", sdk.DefaultBondDenom, demandOrderFulfillerAddr)
	suite.Require().Equal(demandOrder.TrackingPacketStatus, commontypes.Status_PENDING)
	suite.Require().NoError(err)
	err = suite.App.EIBCKeeper.SetDemandOrder(suite.Ctx, demandOrder)
	suite.Require().NoError(err)
	// Update rollapp packet status to finalized
	updatedRollappPacket, err := suite.App.DelayedAckKeeper.UpdateRollappPacketWithStatus(suite.Ctx, *rollappPacket, commontypes.Status_FINALIZED)
	suite.Require().NoError(err)
	// Veirfy that the demand order is updated
	updatedDemandOrder, err := suite.App.EIBCKeeper.GetDemandOrder(suite.Ctx, commontypes.Status_FINALIZED, demandOrder.Id)
	suite.Require().NoError(err)
	suite.Require().Equal(updatedDemandOrder.TrackingPacketStatus, commontypes.Status_FINALIZED)
	rollappPacketKey, err := commontypes.RollappPacketKey(&updatedRollappPacket)
	suite.Require().NoError(err)
	suite.Require().Equal(updatedDemandOrder.TrackingPacketKey, string(rollappPacketKey))
}

func (suite *KeeperTestSuite) TestAfterRollappPacketDeleted() {
	// Set a rollapp packet
	suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)
	// Create new demand order
	demandOrderFulfillerAddr := apptesting.AddTestAddrs(suite.App, suite.Ctx, 1, math.NewInt(1000))[0].String()
	demandOrder, err := types.NewDemandOrder(*rollappPacket, "100", "50", sdk.DefaultBondDenom, demandOrderFulfillerAddr)
	suite.Require().NoError(err)
	err = suite.App.EIBCKeeper.SetDemandOrder(suite.Ctx, demandOrder)
	suite.Require().NoError(err)
	// Update rollapp packet status to finalized
	suite.App.DelayedAckKeeper.UpdateRollappPacketWithStatus(suite.Ctx, *rollappPacket, commontypes.Status_FINALIZED)
	// Trigger the delayed ack hook which should delete the rollapp packet and the demand order
	epochIdentifier := "minute"
	suite.App.DelayedAckKeeper.SetParams(suite.Ctx, delayedacktypes.Params{EpochIdentifier: epochIdentifier})
	hooks := suite.App.DelayedAckKeeper.GetEpochHooks()
	hooks.AfterEpochEnd(suite.Ctx, epochIdentifier, 1)
	// Verify that the rollapp packet and demand order are deleted
	_, err = suite.App.DelayedAckKeeper.GetRollappPacket(suite.Ctx, string(rollappPacketKey))
	suite.Require().ErrorIs(err, delayedacktypes.ErrRollappPacketDoesNotExist)
	_, err = suite.App.EIBCKeeper.GetDemandOrder(suite.Ctx, commontypes.Status_FINALIZED, demandOrder.Id)
	suite.Require().ErrorIs(err, types.ErrDemandOrderDoesNotExist)
}
