package ibctesting_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
	app "github.com/dymensionxyz/dymension/app"
)

// Transfer from cosmos chain to the hub. No delay expected
func (suite *KeeperTestSuite) Test_Bridge_n_Lock() {
	testCases := []struct {
		name           string
		memo           string
		expectedLocked bool
		DstIsModule    bool
		srcToken       bool
		dymToken       bool
	}{
		{
			name:           "missing memo",
			memo:           "",
			expectedLocked: false,
		},
		{
			name:           "missing bridge_and_lock memo",
			memo:           "{\"test\": {}}",
			expectedLocked: false,
		},
		{
			name:           "happy flow",
			memo:           "{\"bridge_and_lock\": {\"to_lock\":true}}",
			expectedLocked: true,
		},
		{
			name:           "happy flow - token originated from hub",
			memo:           "{\"bridge_and_lock\": {\"to_lock\":true}}",
			expectedLocked: true,
			srcToken:       true,
		},
		{
			name:           "not locking native token (DYM)",
			memo:           "{\"bridge_and_lock\": {\"to_lock\":true}}",
			expectedLocked: false,
			dymToken:       true,
			srcToken:       true,
		},
		{
			name:           "bridge_and_lock - false",
			memo:           "{\"bridge_and_lock\": {\"to_lock\":false}}",
			expectedLocked: false,
		},
		{
			name:           "bridge_and_lock - dest is module addr",
			memo:           "{\"bridge_and_lock\": {\"to_lock\":true}}",
			expectedLocked: false,
			DstIsModule:    true,
		},
	}

	for _, tc := range testCases {
		suite.SetupTest()

		// setup between cosmosChain and hubChain
		path := suite.NewTransferPath(suite.hubChain, suite.cosmosChain)
		suite.coordinator.Setup(path)

		hubEndpoint := path.EndpointA
		cosmosEndpoint := path.EndpointB

		timeoutHeight := clienttypes.NewHeight(100, 110)
		amount, ok := sdk.NewIntFromString("10000000000000000000") //10DYM
		suite.Require().True(ok)
		coinToSend := sdk.NewCoin(sdk.DefaultBondDenom, amount)
		expectedLockDenom := ""

		//send from hubChain to cosmosChain, than send it back and lock it
		if tc.srcToken {
			expectedLockDenom = sdk.DefaultBondDenom
			if !tc.dymToken {
				coinToSend = sdk.NewCoin("atom", amount)
				app.FundAccount(ConvertToApp(suite.hubChain), suite.hubChain.GetContext(), suite.hubChain.SenderAccount.GetAddress(), sdk.Coins{coinToSend})
				expectedLockDenom = "atom"
			}

			// send from cosmosChain to hubChain
			dstAcc := cosmosEndpoint.Chain.SenderAccount.GetAddress()
			msg := types.NewMsgTransfer(hubEndpoint.ChannelConfig.PortID, hubEndpoint.ChannelID, coinToSend, hubEndpoint.Chain.SenderAccount.GetAddress().String(), dstAcc.String(), timeoutHeight, 0, "")
			res, err := hubEndpoint.Chain.SendMsgs(msg)
			suite.Require().NoError(err, tc.name) // message committed

			packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
			suite.Require().NoError(err)

			// relay send
			err = path.RelayPacket(packet)
			suite.Require().NoError(err, tc.name) // relay committed

			coinToSend = sdk.NewCoin(types.ParseDenomTrace(types.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), coinToSend.Denom)).IBCDenom(), amount)
		}

		// send from cosmosChain to hubChain
		dstAcc := hubEndpoint.Chain.SenderAccount.GetAddress()
		if tc.DstIsModule {
			dstAcc = ConvertToApp(suite.hubChain).AccountKeeper.GetModuleAddress("streamer")
		}
		msg := types.NewMsgTransfer(cosmosEndpoint.ChannelConfig.PortID, cosmosEndpoint.ChannelID, coinToSend, cosmosEndpoint.Chain.SenderAccount.GetAddress().String(), dstAcc.String(), timeoutHeight, 0, tc.memo)
		res, err := cosmosEndpoint.Chain.SendMsgs(msg)
		suite.Require().NoError(err, tc.name) // message committed

		packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
		suite.Require().NoError(err)

		if expectedLockDenom == "" {
			stakeVoucherDenom := types.ParseDenomTrace(types.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), sdk.DefaultBondDenom))
			expectedLockDenom = stakeVoucherDenom.IBCDenom()
		}
		balanceBefore := ConvertToApp(suite.hubChain).BankKeeper.GetBalance(suite.hubChain.GetContext(), dstAcc, expectedLockDenom)

		// relay send
		err = path.RelayPacket(packet)
		suite.Require().NoError(err, tc.name) // relay committed

		balanceAfter := ConvertToApp(suite.hubChain).BankKeeper.GetBalance(suite.hubChain.GetContext(), dstAcc, expectedLockDenom)
		if !tc.expectedLocked {
			suite.Require().True(balanceBefore.IsLT(balanceAfter), tc.name)
		} else {
			suite.Require().Equal(balanceBefore.String(), balanceAfter.String(), tc.name)
			locks := ConvertToApp(suite.hubChain).LockupKeeper.GetAccountPeriodLocks(suite.hubChain.GetContext(), dstAcc)
			suite.Require().Len(locks, 1, tc.name)
			suite.Require().Equal(locks[0].Coins[0].Denom, expectedLockDenom, tc.name)
		}
	}
}
