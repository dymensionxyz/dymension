package ibctesting_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v6/testing"
)

// Transfer from cosmos chain to the hub. No delay expected
func (suite *KeeperTestSuite) Test_Bridge_n_Lock() {
	testCases := []struct {
		name           string
		memo           string
		expectedLocked bool
		DstIsModule    bool
	}{
		{
			name:           "missing memo",
			memo:           "",
			expectedLocked: false,
		},
		{
			name:           "missing bridge_and_lock memo",
			memo:           "{test: {}}",
			expectedLocked: false,
		},
		{
			name:           "happy flow",
			memo:           "{bridge_and_lock: {to_lock:true}}}",
			expectedLocked: true,
		},
		// {
		// 	name:           "happy flow - token originated from hub",
		// 	memo:           "bridge_and_lock",
		// 	expectedLocked: true,
		// },
		// {
		// 	name:           "not locking native token (DYM)",
		// 	memo:           "bridge_and_lock",
		// 	expectedLocked: false,
		// },
		{
			name:           "bridge_and_lock - false",
			memo:           "{bridge_and_lock: {to_lock:false}}}",
			expectedLocked: false,
		},
		{
			name:           "bridge_and_lock - dest is module addr",
			memo:           "{bridge_and_lock: {to_lock:true}}}",
			expectedLocked: false,
			DstIsModule:    true,
		},
	}

	for _, tc := range testCases {
		// setup between cosmosChain and hubChain
		path := suite.NewTransferPath(suite.hubChain, suite.cosmosChain)
		suite.coordinator.Setup(path)

		hubEndpoint := path.EndpointA
		cosmosEndpoint := path.EndpointB

		timeoutHeight := clienttypes.NewHeight(100, 110)
		amount, ok := sdk.NewIntFromString("10000000000000000000") //10DYM
		suite.Require().True(ok)
		coinToSend := sdk.NewCoin(sdk.DefaultBondDenom, amount)

		// send from cosmosChain to hubChain
		dstAddr := hubEndpoint.Chain.SenderAccount.GetAddress().String()
		if tc.DstIsModule {
			dstAddr = ConvertToApp(suite.hubChain).AccountKeeper.GetModuleAddress("streamer").String()
		}
		msg := types.NewMsgTransfer(cosmosEndpoint.ChannelConfig.PortID, cosmosEndpoint.ChannelID, coinToSend, cosmosEndpoint.Chain.SenderAccount.GetAddress().String(), dstAddr, timeoutHeight, 0, tc.memo)
		res, err := cosmosEndpoint.Chain.SendMsgs(msg)
		suite.Require().NoError(err, tc.name) // message committed

		packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
		suite.Require().NoError(err)
		stakeVoucherDenom := types.ParseDenomTrace(types.GetPrefixedDenom(packet.GetDestPort(), packet.GetDestChannel(), sdk.DefaultBondDenom))
		balanceBefore := ConvertToApp(suite.hubChain).BankKeeper.GetBalance(suite.hubChain.GetContext(), hubEndpoint.Chain.SenderAccount.GetAddress(), stakeVoucherDenom.IBCDenom())

		// relay send
		err = path.RelayPacket(packet)
		suite.Require().NoError(err, tc.name) // relay committed

		balanceAfter := ConvertToApp(suite.hubChain).BankKeeper.GetBalance(suite.hubChain.GetContext(), hubEndpoint.Chain.SenderAccount.GetAddress(), stakeVoucherDenom.IBCDenom())
		if !tc.expectedLocked {
			suite.Require().True(balanceBefore.IsLT(balanceAfter), tc.name)
		} else {
			suite.Require().Equal(balanceBefore, balanceAfter, tc.name)
			locks := ConvertToApp(suite.hubChain).LockupKeeper.GetAccountPeriodLocks(suite.hubChain.GetContext(), hubEndpoint.Chain.SenderAccount.GetAddress())
			suite.Require().Len(locks, 1, tc.name)
			suite.Require().Equal(locks[0].Coins[0].Denom, stakeVoucherDenom.IBCDenom(), tc.name)
		}
	}
}
