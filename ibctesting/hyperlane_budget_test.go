package ibctesting_test

import (
	"cosmossdk.io/math"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	ismtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"
	pdtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/types"
	coretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	warpkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/warp/keeper"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	ibctesting "github.com/cosmos/ibc-go/v8/testing"
	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	forwardtypes "github.com/dymensionxyz/dymension/v3/x/forward/types"
)

func (s *eibcForwardSuite) TestIBCFullBudgetHyperlane() {
	for _, rollapp := range []bool{true, false} {
		for _, memoType := range []string{"direct", "eibc"} {
			source := "non_rollapp"
			if rollapp {
				source = "finalized_rollapp"
			}
			s.Run(source+"/"+memoType, func() {
				s.SetupTest()
				params := s.hubApp().DelayedAckKeeper.GetParams(s.hubCtx())
				params.BridgingFee = math.LegacyMustNewDecFromStr("0.001")
				s.hubApp().DelayedAckKeeper.SetParams(s.hubCtx(), params)
				path := s.path
				chain := s.rollappChain()
				if !rollapp {
					chain = s.cosmosChain()
					path = s.newTransferPath(s.hubChain(), chain)
					s.coordinator.Setup(path)
				}
				owner := s.hubChain().SenderAccount.GetAddress().String()
				denom := transfertypes.ParseDenomTrace(transfertypes.GetPrefixedDenom(path.EndpointA.ChannelConfig.PortID, path.EndpointA.ChannelID, sdk.DefaultBondDenom)).IBCDenom()

				// Configure real Hyperlane keepers. Noop dispatch hooks charge no gas, so the
				// only newly received funds left after forwarding should be the fee reserve.
				res, err := s.hubChain().SendMsgs(&ismtypes.MsgCreateNoopIsm{Creator: owner})
				s.Require().NoError(err)
				var txData sdk.TxMsgData
				s.Require().NoError(proto.Unmarshal(res.Data, &txData))
				var ism ismtypes.MsgCreateNoopIsmResponse
				s.Require().NoError(proto.Unmarshal(txData.MsgResponses[0].Value, &ism))
				res, err = s.hubChain().SendMsgs(&pdtypes.MsgCreateNoopHook{Owner: owner})
				s.Require().NoError(err)
				txData.Reset()
				s.Require().NoError(proto.Unmarshal(res.Data, &txData))
				var dispatch pdtypes.MsgCreateNoopHookResponse
				s.Require().NoError(proto.Unmarshal(txData.MsgResponses[0].Value, &dispatch))
				mailbox, err := s.hubApp().HyperCoreKeeper.CreateMailbox(s.hubCtx(), &coretypes.MsgCreateMailbox{Owner: owner, DefaultIsm: ism.Id, DefaultHook: &dispatch.Id, RequiredHook: &dispatch.Id})
				s.Require().NoError(err)
				tokenID, err := s.hubApp().HyperWarpKeeper.CreateCollateralToken(s.hubCtx(), &warptypes.MsgCreateCollateralToken{Owner: owner, OriginMailbox: mailbox, OriginDenom: denom})
				s.Require().NoError(err)
				_, err = warpkeeper.NewMsgServerImpl(s.hubApp().HyperWarpKeeper).EnrollRemoteRouter(s.hubCtx(), &warptypes.MsgEnrollRemoteRouter{Owner: owner, TokenId: tokenID, RemoteRouter: &warptypes.RemoteRouter{ReceiverDomain: 1, ReceiverContract: hyperutil.HexAddress{}.String(), Gas: math.ZeroInt()}})
				s.Require().NoError(err)

				gross := math.NewInt(1_000_000)
				fee := math.NewInt(10) // Much smaller than the 1,000-unit bridging fee.
				net := gross
				if rollapp {
					net = net.Sub(s.hubApp().DelayedAckKeeper.BridgingFeeFromAmt(s.hubCtx(), gross))
				}
				hook := forwardtypes.NewHookForwardToHL(tokenID, 1, hyperutil.HexAddress{}, gross, sdk.NewCoin(denom, fee), math.ZeroInt(), nil, "", true, net.Sub(fee))
				var memo string
				if memoType == "direct" {
					memo, err = forwardtypes.MakeIBCForwardToHLMemoString(hook)
				} else {
					memo, err = forwardtypes.MakeRolForwardToHLMemoString("100", hook)
				}
				s.Require().NoError(err)
				recipient := s.hubChain().SenderAccount.GetAddress()
				s.Require().True(s.hubApp().BankKeeper.GetBalance(s.hubCtx(), recipient, denom).IsZero())
				previousBalance := math.NewInt(77)
				apptesting.FundAccount(s.hubApp(), s.hubCtx(), recipient, sdk.NewCoins(sdk.NewCoin(denom, previousBalance)))
				res, err = chain.SendMsgs(transfertypes.NewMsgTransfer(path.EndpointB.ChannelConfig.PortID, path.EndpointB.ChannelID, sdk.NewCoin(sdk.DefaultBondDenom, gross), chain.SenderAccount.GetAddress().String(), recipient.String(), clienttypes.NewHeight(100, 110), 0, memo))
				s.Require().NoError(err)
				packet, err := ibctesting.ParsePacketFromEvents(res.GetEvents())
				s.Require().NoError(err)
				if rollapp {
					s.rollappChain().NextBlock()
					height := uint64(s.rollappCtx().BlockHeight())
					s.updateRollappState(height)
					_, err = s.finalizeRollappState(1, height)
					s.Require().NoError(err)
				}
				s.Require().NoError(path.RelayPacket(packet))

				// Dispatch and actual collateral movement must commit, rather than just an
				// IBC success acknowledgement with a failed forward left at the recipient.
				storedMailbox, err := s.hubApp().HyperCoreKeeper.Mailboxes.Get(s.hubCtx(), mailbox.GetInternalId())
				s.Require().NoError(err)
				s.Require().Equal(uint32(1), storedMailbox.MessageSent)
				token, err := s.hubApp().HyperWarpKeeper.HypTokens.Get(s.hubCtx(), tokenID.GetInternalId())
				s.Require().NoError(err)
				s.Require().Equal(net.Sub(fee), token.CollateralBalance)
				s.Require().Equal(previousBalance.Add(fee), s.hubApp().BankKeeper.GetBalance(s.hubCtx(), recipient, denom).Amount)
			})
		}
	}
}
