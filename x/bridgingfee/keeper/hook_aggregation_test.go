package keeper_test

import (
	"cosmossdk.io/math"
	hyputil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/keeper"
	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"
)

func (s *KeeperTestSuite) TestAggregationHookPostDispatch() {
	// Create test addresses
	creator := s.CreateRandomAccount()
	sender := s.CreateRandomAccount()

	// Fund accounts
	initialFunds := sdk.NewCoins(
		sdk.NewCoin("stake", math.NewInt(10_000_000)),
	)
	s.FundAcc(creator, initialFunds)
	s.FundAcc(sender, initialFunds)

	// Create a dummy mailbox and ISM
	mailboxId, _ := s.createDummyMailbox(creator.String())

	// Create a dummy token in warp module (using stake for simplicity)
	tokenId := s.createDummyToken(creator.String(), mailboxId, "stake")

	// Create two fee hooks with fees for the same token
	feeMsg1 := &types.MsgCreateBridgingFeeHook{
		Owner: creator.String(),
		Fees: []types.HLAssetFee{
			{
				TokenID:     tokenId.String(),
				InboundFee:  math.LegacyMustNewDecFromStr("0.01"), // 1%
				OutboundFee: math.LegacyMustNewDecFromStr("0.02"), // 2%
			},
		},
	}

	feeMsg2 := &types.MsgCreateBridgingFeeHook{
		Owner: creator.String(),
		Fees: []types.HLAssetFee{
			{
				TokenID:     tokenId.String(),
				InboundFee:  math.LegacyMustNewDecFromStr("0.01"), // 1%
				OutboundFee: math.LegacyMustNewDecFromStr("0.03"), // 3%
			},
		},
	}

	// Create the two fee hooks
	hookId1, err := s.App.BridgingFeeKeeper.CreateFeeHook(s.Ctx, feeMsg1)
	s.Require().NoError(err)
	s.Require().NotEmpty(hookId1)

	hookId2, err := s.App.BridgingFeeKeeper.CreateFeeHook(s.Ctx, feeMsg2)
	s.Require().NoError(err)
	s.Require().NotEmpty(hookId2)

	// Create an aggregation hook that combines both fee hooks
	aggMsg := &types.MsgCreateAggregationHook{
		Owner:   creator.String(),
		HookIds: []hyputil.HexAddress{hookId1, hookId2},
	}

	aggHookId, err := s.App.BridgingFeeKeeper.CreateAggregationHook(s.Ctx, aggMsg)
	s.Require().NoError(err)
	s.Require().NotEmpty(aggHookId)

	// Test PostDispatch flow
	s.Run("PostDispatch with aggregated fee collection", func() {
		transferAmount := math.NewInt(1_000_000)

		// Create message body (warp payload)
		payload, err := warptypes.NewWarpPayload(
			make([]byte, 32), // dummy recipient
			*transferAmount.BigInt(),
			[]byte{}, // no metadata
		)
		s.Require().NoError(err)
		body := payload.Bytes()

		// Create hyperlane message
		recipient, _ := hyputil.DecodeHexAddress("0xd7194459d45619d04a5a0f9e78dc9594a0f37fd6da8382fe12ddda6f2f46d647")
		message := hyputil.HyperlaneMessage{
			Version:     1,
			Nonce:       1,
			Origin:      11,
			Sender:      tokenId,
			Destination: 1,
			Recipient:   recipient,
			Body:        body,
		}

		metadata := hyputil.StandardHookMetadata{
			GasLimit: math.NewInt(50_000),
			Address:  sender,
		}

		// Get the aggregation handler
		aggHandler := keeper.NewAggregationHookHandler(s.App.BridgingFeeKeeper)

		// Test QuoteDispatch first
		maxFee := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(200_000)))
		quotedFee, err := aggHandler.QuoteDispatch(s.Ctx, mailboxId, aggHookId, metadata, message)
		s.Require().NoError(err)
		s.Require().NotEmpty(quotedFee)

		// Expected total fee: 2% + 3% = 5% of 1 stake = 50,000 stake
		expectedFeeAmt := transferAmount.MulRaw(5).QuoRaw(100) // 5% = 50,000 stake
		expectedFee := sdk.NewCoins(sdk.NewCoin("stake", expectedFeeAmt))
		s.Require().Equal(expectedFee, quotedFee)
		s.T().Logf("Transfer amount: %s", transferAmount.String())
		s.T().Logf("Quoted aggregated fee: %s", quotedFee.String())

		// Check initial balance
		initialBalance := s.App.BankKeeper.GetBalance(s.Ctx, sender, "stake")

		// Test PostDispatch
		collectedFee, err := aggHandler.PostDispatch(s.Ctx, mailboxId, aggHookId, metadata, message, maxFee)
		s.Require().NoError(err)
		s.Require().Equal(expectedFee, collectedFee)

		// Check that fee was collected from sender
		finalBalance := s.App.BankKeeper.GetBalance(s.Ctx, sender, "stake")

		expectedFinalBalance := initialBalance.Sub(collectedFee[0])
		s.Require().Equal(expectedFinalBalance, finalBalance)

		// Check that fee was sent to bridgingfee module
		moduleBalance := s.App.BankKeeper.GetBalance(s.Ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), "stake")
		s.Require().True(moduleBalance.Amount.GTE(collectedFee[0].Amount),
			"Module balance should be at least the collected fee: expected >= %s, got %s",
			collectedFee[0], moduleBalance)
	})

	s.Run("PostDispatch with empty aggregation hook", func() {
		// Create an empty aggregation hook (no sub-hooks)
		emptyAggMsg := &types.MsgCreateAggregationHook{
			Owner:   creator.String(),
			HookIds: []hyputil.HexAddress{},
		}

		emptyAggHookId, err := s.App.BridgingFeeKeeper.CreateAggregationHook(s.Ctx, emptyAggMsg)
		s.Require().NoError(err)
		s.Require().NotEmpty(emptyAggHookId)

		transferAmount := math.NewInt(1_000_000)

		// Create message body
		payload, err := warptypes.NewWarpPayload(
			make([]byte, 32), // dummy recipient
			*transferAmount.BigInt(),
			[]byte{}, // no metadata
		)
		s.Require().NoError(err)
		body := payload.Bytes()

		// Create hyperlane message
		recipient, _ := hyputil.DecodeHexAddress("0xd7194459d45619d04a5a0f9e78dc9594a0f37fd6da8382fe12ddda6f2f46d647")
		message := hyputil.HyperlaneMessage{
			Version:     1,
			Nonce:       2,
			Origin:      11,
			Sender:      tokenId,
			Destination: 1,
			Recipient:   recipient,
			Body:        body,
		}

		metadata := hyputil.StandardHookMetadata{
			GasLimit: math.NewInt(50_000),
			Address:  sender,
		}

		aggHandler := keeper.NewAggregationHookHandler(s.App.BridgingFeeKeeper)

		// Should return zero fee for empty aggregation
		quotedFee, err := aggHandler.QuoteDispatch(s.Ctx, mailboxId, emptyAggHookId, metadata, message)
		s.Require().NoError(err)
		s.Require().True(quotedFee.IsZero())

		// PostDispatch should also return zero fee
		maxFee := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100_000)))
		collectedFee, err := aggHandler.PostDispatch(s.Ctx, mailboxId, emptyAggHookId, metadata, message, maxFee)
		s.Require().NoError(err)
		s.Require().True(collectedFee.IsZero())
	})
}
