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
	demandOrder := types.NewDemandOrder(*rollappPacket, math.NewIntFromUint64(100), math.NewIntFromUint64(50), sdk.DefaultBondDenom, demandOrderFulfillerAddr, 1, nil)
	suite.Require().Equal(commontypes.Status_PENDING, demandOrder.TrackingPacketStatus)
	err := suite.App.EIBCKeeper.SetDemandOrder(suite.Ctx, demandOrder)
	suite.Require().NoError(err)
	// Update rollapp packet status to finalized
	updatedRollappPacket, err := suite.App.DelayedAckKeeper.UpdateRollappPacketAfterFinalization(suite.Ctx, *rollappPacket)
	suite.Require().NoError(err)
	// Veirfy that the demand order is updated
	updatedDemandOrder, err := suite.App.EIBCKeeper.GetDemandOrder(suite.Ctx, commontypes.Status_FINALIZED, demandOrder.Id)
	suite.Require().NoError(err)
	suite.Require().Equal(commontypes.Status_FINALIZED, updatedDemandOrder.TrackingPacketStatus)
	rollappPacketKey := updatedRollappPacket.RollappPacketKey()
	suite.Require().NoError(err)
	suite.Require().Equal(string(rollappPacketKey), updatedDemandOrder.TrackingPacketKey)
}

func (suite *KeeperTestSuite) TestAfterRollappPacketDeleted() {
	testCases := []struct {
		name          string
		packetStatus  commontypes.Status
		expectedError error
	}{
		{
			name:          "Pending packet",
			packetStatus:  commontypes.Status_PENDING,
			expectedError: types.ErrDemandOrderDoesNotExist,
		},
		{
			name:          "Finalized packet",
			packetStatus:  commontypes.Status_FINALIZED,
			expectedError: types.ErrDemandOrderDoesNotExist,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Set a rollapp packet
			suite.App.DelayedAckKeeper.SetRollappPacket(suite.Ctx, *rollappPacket)

			// Create new demand order
			demandOrderFulfillerAddr := apptesting.AddTestAddrs(suite.App, suite.Ctx, 1, math.NewInt(1000))[0].String()
			demandOrder := types.NewDemandOrder(*rollappPacket, math.NewIntFromUint64(100), math.NewIntFromUint64(50), sdk.DefaultBondDenom, demandOrderFulfillerAddr, 1, nil)
			err := suite.App.EIBCKeeper.SetDemandOrder(suite.Ctx, demandOrder)
			suite.Require().NoError(err)
			_, err = suite.App.EIBCKeeper.GetDemandOrder(suite.Ctx, commontypes.Status_PENDING, demandOrder.Id)
			suite.Require().NoError(err)

			// Update rollapp packet status
			if tc.packetStatus == commontypes.Status_FINALIZED {
				_, err = suite.App.DelayedAckKeeper.UpdateRollappPacketAfterFinalization(suite.Ctx, *rollappPacket)
				suite.Require().NoError(err)
			}

			// delete the rollapp packet
			suite.App.DelayedAckKeeper.DeleteRollappPacket(suite.Ctx, rollappPacket)

			// Verify that the rollapp packet and demand order are deleted
			_, err = suite.App.DelayedAckKeeper.GetRollappPacket(suite.Ctx, string(rollappPacketKey))
			suite.Require().ErrorIs(err, delayedacktypes.ErrRollappPacketDoesNotExist)
			_, err = suite.App.EIBCKeeper.GetDemandOrder(suite.Ctx, tc.packetStatus, demandOrder.Id)
			suite.Require().ErrorIs(err, tc.expectedError)
		})
	}
}
