package keeper_test

import (
	"cosmossdk.io/math"
	hyputil "github.com/bcp-innovations/hyperlane-cosmos/util"
	ismTypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"
	coreTypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/keeper"
	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"
)

func (s *KeeperTestSuite) TestFeeHookPostDispatch() {
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

	// Create a fee hook with fees for this token
	feeMsg := &types.MsgCreateBridgingFeeHook{
		Owner: creator.String(),
		Fees: []types.HLAssetFee{
			{
				TokenID:     tokenId.String(),
				InboundFee:  math.LegacyMustNewDecFromStr("0.01"), // 1%
				OutboundFee: math.LegacyMustNewDecFromStr("0.02"), // 2%
			},
		},
	}

	// Create the fee hook
	hookId, err := s.App.BridgingFeeKeeper.CreateFeeHook(s.Ctx, feeMsg)
	s.Require().NoError(err)
	s.Require().NotEmpty(hookId)

	// Test PostDispatch flow
	s.Run("PostDispatch with fee collection", func() {
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

		// Create metadata
		metadata := hyputil.StandardHookMetadata{
			GasLimit: math.NewInt(50_000),
			Address:  sender,
		}

		// Get the fee handler
		feeHandler := keeper.NewFeeHookHandler(s.App.BridgingFeeKeeper)

		// Test QuoteDispatch first
		maxFee := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100_000)))
		quotedFee, err := feeHandler.QuoteDispatch(s.Ctx, mailboxId, hookId, metadata, message)
		s.Require().NoError(err)
		s.Require().NotEmpty(quotedFee)

		// Verify quoted fee is reasonable (2% of 1 stake)
		expectedFeeAmt := transferAmount.MulRaw(2).QuoRaw(100) // 2% = 20,000 stake
		expectedFee := sdk.NewCoins(sdk.NewCoin("stake", expectedFeeAmt))
		s.Require().Equal(expectedFee, quotedFee)

		// Check initial balance
		initialBalance := s.App.BankKeeper.GetBalance(s.Ctx, sender, "stake")

		// Test PostDispatch
		collectedFee, err := feeHandler.PostDispatch(s.Ctx, mailboxId, hookId, metadata, message, maxFee)
		s.Require().NoError(err)
		s.Require().Equal(expectedFee, collectedFee)

		// Check that fee was collected from sender
		finalBalance := s.App.BankKeeper.GetBalance(s.Ctx, sender, "stake")

		expectedFinalBalance := initialBalance.Sub(collectedFee[0])
		s.Require().Equal(expectedFinalBalance, finalBalance)

		// Check that fee was sent to bridgingfee module
		moduleBalance := s.App.BankKeeper.GetBalance(s.Ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), "stake")
		s.Require().Equal(collectedFee[0], moduleBalance)
	})

	s.Run("PostDispatch with no fee configuration", func() {
		// Create a different token ID that has no fee configuration
		unknownTokenId, _ := hyputil.DecodeHexAddress("0x1234567890123456789012345678901234567890123456789012345678901234")

		transferAmount := math.NewInt(1000_000)

		// Create message body
		payload, err := warptypes.NewWarpPayload(
			make([]byte, 32), // dummy recipient
			*transferAmount.BigInt(),
			[]byte{}, // no metadata
		)
		s.Require().NoError(err)
		body := payload.Bytes()

		// Create hyperlane message with unknown token
		recipient, _ := hyputil.DecodeHexAddress("0xd7194459d45619d04a5a0f9e78dc9594a0f37fd6da8382fe12ddda6f2f46d647")
		message := hyputil.HyperlaneMessage{
			Version:     1,
			Nonce:       2,
			Origin:      11,
			Sender:      unknownTokenId,
			Destination: 1,
			Recipient:   recipient,
			Body:        body,
		}

		metadata := hyputil.StandardHookMetadata{
			GasLimit: math.NewInt(50_000),
			Address:  sender,
		}

		feeHandler := keeper.NewFeeHookHandler(s.App.BridgingFeeKeeper)

		// Should return zero fee for unknown token
		quotedFee, err := feeHandler.QuoteDispatch(s.Ctx, mailboxId, hookId, metadata, message)
		s.Require().NoError(err)
		s.Require().True(quotedFee.IsZero())

		// PostDispatch should also return zero fee
		maxFee := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(100_000)))
		collectedFee, err := feeHandler.PostDispatch(s.Ctx, mailboxId, hookId, metadata, message, maxFee)
		s.Require().NoError(err)
		s.Require().True(collectedFee.IsZero())
	})

	s.Run("PostDispatch with non-base denom and AMM pool conversion", func() {
		// Create a new creator and sender for this test
		creator2 := s.CreateRandomAccount()
		sender2 := s.CreateRandomAccount()

		// Fund accounts with both adym and stake
		initialFunds := sdk.NewCoins(
			sdk.NewCoin("adym", math.NewInt(10_000_000)),
			sdk.NewCoin("stake", math.NewInt(10_000_000)),
		)
		s.FundAcc(creator2, initialFunds)
		s.FundAcc(sender2, initialFunds)

		// Create AMM pool between adym and stake (1:2 ratio)
		poolCoins := sdk.NewCoins(
			sdk.NewCoin("adym", math.NewInt(1_000_000)),  // 1M adym
			sdk.NewCoin("stake", math.NewInt(2_000_000)), // 2M stake (base denom)
		)
		s.PreparePoolWithCoins(poolCoins)

		// Create a dummy mailbox and ISM
		mailboxId2, _ := s.createDummyMailbox(creator2.String())

		// Create a dummy token in warp module using adym (non-base denom)
		tokenId2 := s.createDummyToken(creator2.String(), mailboxId2, "adym")

		// Create a fee hook with fees for the adym token
		feeMsg2 := &types.MsgCreateBridgingFeeHook{
			Owner: creator2.String(),
			Fees: []types.HLAssetFee{
				{
					TokenID:     tokenId2.String(),
					InboundFee:  math.LegacyMustNewDecFromStr("0.01"), // 1%
					OutboundFee: math.LegacyMustNewDecFromStr("0.03"), // 3%
				},
			},
		}

		// Create the fee hook
		hookId2, err := s.App.BridgingFeeKeeper.CreateFeeHook(s.Ctx, feeMsg2)
		s.Require().NoError(err)
		s.Require().NotEmpty(hookId2)

		transferAmount := math.NewInt(1_000_000) // 1M adym transfer

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
			Nonce:       3,
			Origin:      11,
			Sender:      tokenId2,
			Destination: 1,
			Recipient:   recipient,
			Body:        body,
		}

		metadata := hyputil.StandardHookMetadata{
			GasLimit: math.NewInt(50_000),
			Address:  sender2,
		}

		// Get the fee handler
		feeHandler := keeper.NewFeeHookHandler(s.App.BridgingFeeKeeper)

		// Test QuoteDispatch first
		maxFee := sdk.NewCoins(sdk.NewCoin("stake", math.NewInt(200_000)))
		quotedFee, err := feeHandler.QuoteDispatch(s.Ctx, mailboxId2, hookId2, metadata, message)
		s.Require().NoError(err)
		s.Require().NotEmpty(quotedFee)

		// The fee should be in stake (base denom) even though the token is adym
		s.Require().Equal("stake", quotedFee[0].Denom)
		s.T().Logf("Transfer amount: %s adym", transferAmount.String())
		s.T().Logf("Quoted fee: %s", quotedFee.String())

		// Expected fee calculation:
		// 3% of 1M adym = 30,000 adym
		// With 1:2 pool ratio (1 adym = 2 stake), 30,000 adym should convert to ~60,000 stake
		expectedFeeAdym := transferAmount.MulRaw(3).QuoRaw(100) // 3% = 30,000 adym
		s.T().Logf("Expected fee in adym: %s", expectedFeeAdym.String())

		// Check initial balances
		initialBalanceStake := s.App.BankKeeper.GetBalance(s.Ctx, sender2, "stake")
		initialBalanceAdym := s.App.BankKeeper.GetBalance(s.Ctx, sender2, "adym")
		s.T().Logf("Initial sender stake balance: %s", initialBalanceStake.String())
		s.T().Logf("Initial sender adym balance: %s", initialBalanceAdym.String())

		// Test PostDispatch
		collectedFee, err := feeHandler.PostDispatch(s.Ctx, mailboxId2, hookId2, metadata, message, maxFee)
		s.Require().NoError(err)
		s.Require().Equal(quotedFee, collectedFee)

		// Check that fee was collected from sender in stake (base denom)
		finalBalanceStake := s.App.BankKeeper.GetBalance(s.Ctx, sender2, "stake")
		finalBalanceAdym := s.App.BankKeeper.GetBalance(s.Ctx, sender2, "adym")
		s.T().Logf("Final sender stake balance: %s", finalBalanceStake.String())
		s.T().Logf("Final sender adym balance: %s", finalBalanceAdym.String())

		// Stake balance should have decreased by the collected fee
		expectedFinalBalanceStake := initialBalanceStake.Sub(collectedFee[0])
		s.Require().Equal(expectedFinalBalanceStake, finalBalanceStake)

		// Adym balance should remain unchanged (fee was charged in stake)
		s.Require().Equal(initialBalanceAdym, finalBalanceAdym)

		// Check that fee was sent to bridgingfee module
		moduleBalance := s.App.BankKeeper.GetBalance(s.Ctx, s.App.AccountKeeper.GetModuleAddress(types.ModuleName), "stake")
		s.T().Logf("Expected module balance: %s", collectedFee[0].String())
		s.T().Logf("Actual module balance: %s", moduleBalance.String())

		// Note: There might be residual balance from previous tests, so we check that the module
		// balance increased by at least the collected fee amount
		s.Require().True(moduleBalance.Amount.GTE(collectedFee[0].Amount),
			"Module balance should be at least the collected fee: expected >= %s, got %s",
			collectedFee[0], moduleBalance)
	})
}

// Helper function to create a real mailbox and ISM
func (s *KeeperTestSuite) createDummyMailbox(creator string) (hyputil.HexAddress, hyputil.HexAddress) {
	s.T().Helper()

	// Create a noop ISM first
	ismMsg := &ismTypes.MsgCreateNoopIsm{
		Creator: creator,
	}

	ismId, err := s.App.HyperCoreKeeper.IsmKeeper.CreateNoopIsm(s.Ctx, ismMsg)
	s.Require().NoError(err)

	// Create a mailbox
	mailboxMsg := &coreTypes.MsgCreateMailbox{
		Owner:        creator,
		LocalDomain:  11,
		DefaultIsm:   ismId,
		DefaultHook:  nil,
		RequiredHook: nil,
	}

	mailboxId, err := s.App.HyperCoreKeeper.CreateMailbox(s.Ctx, mailboxMsg)
	s.Require().NoError(err)

	return mailboxId, ismId
}

// Helper function to create a real token in the warp module
func (s *KeeperTestSuite) createDummyToken(creator string, mailboxId hyputil.HexAddress, originDenom string) hyputil.HexAddress {
	s.T().Helper()

	// Create a collateral token
	tokenMsg := &warptypes.MsgCreateCollateralToken{
		Owner:         creator,
		OriginMailbox: mailboxId,
		OriginDenom:   originDenom,
	}

	tokenId, err := s.App.HyperWarpKeeper.CreateCollateralToken(s.Ctx, tokenMsg)
	s.Require().NoError(err)

	return tokenId
}
