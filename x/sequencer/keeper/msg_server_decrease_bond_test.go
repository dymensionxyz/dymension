package keeper_test

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func (s *SequencerTestSuite) TestDecreaseBond() {
	s.SetupTest()
	bondDenom := types.DefaultParams().MinBond.Denom
	rollappId, pk := s.createRollapp()
	// setup a default sequencer with has minBond + 20token
	defaultSequencerAddress := s.createSequencerWithBond(s.Ctx, rollappId, pk, bond.AddAmount(sdk.NewInt(20)))
	// setup an unbonded sequencer
	unbondedPk := ed25519.GenPrivKey().PubKey()
	unbondedSequencerAddress := s.createSequencer(s.Ctx, rollappId, unbondedPk)
	unbondedSequencer, _ := s.App.SequencerKeeper.GetRealSequencer(s.Ctx, unbondedSequencerAddress)
	unbondedSequencer.Status = types.Unbonded
	s.App.SequencerKeeper.SetSequencer(s.Ctx, unbondedSequencer)

	testCase := []struct {
		name        string
		msg         types.MsgDecreaseBond
		expectedErr error
	}{
		{
			name: "invalid sequencer",
			msg: types.MsgDecreaseBond{
				Creator:        "invalid_address",
				DecreaseAmount: sdk.NewInt64Coin(bondDenom, 10),
			},
			expectedErr: types.ErrSequencerNotFound,
		},
		{
			name: "decreased bond value to less than minimum bond value",
			msg: types.MsgDecreaseBond{
				Creator:        defaultSequencerAddress,
				DecreaseAmount: sdk.NewInt64Coin(bondDenom, 100),
			},
			expectedErr: types.ErrInsufficientBond,
		},
		{
			name: "trying to decrease more bond than they have tokens bonded",
			msg: types.MsgDecreaseBond{
				Creator:        defaultSequencerAddress,
				DecreaseAmount: bond.AddAmount(sdk.NewInt(30)),
			},
			expectedErr: types.ErrInsufficientBond,
		},
		{
			name: "valid decrease bond",
			msg: types.MsgDecreaseBond{
				Creator:        defaultSequencerAddress,
				DecreaseAmount: sdk.NewInt64Coin(bondDenom, 10),
			},
		},
	}

	for _, tc := range testCase {
		s.Run(tc.name, func() {
			resp, err := s.msgServer.DecreaseBond(s.Ctx, &tc.msg)
			if tc.expectedErr != nil {
				s.Require().ErrorIs(err, tc.expectedErr)
			} else {
				s.Require().NoError(err)
				s.Require().NotNil(resp)
			}
		})
	}
}

func (s *SequencerTestSuite) TestDecreaseBond_BondDecreaseInProgress() {
	s.SetupTest()
	bondDenom := types.DefaultParams().MinBond.Denom
	rollappId, pk := s.createRollapp()
	// setup a default sequencer with has minBond + 20token
	defaultSequencerAddress := s.createSequencerWithBond(s.Ctx, rollappId, pk, bond.AddAmount(sdk.NewInt(20)))
	// decrease the bond of the sequencer
	_, err := s.msgServer.DecreaseBond(s.Ctx, &types.MsgDecreaseBond{
		Creator:        defaultSequencerAddress,
		DecreaseAmount: sdk.NewInt64Coin(bondDenom, 10),
	})
	s.Require().NoError(err)
	// try to decrease the bond again - should be fine as still not below minbond
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1).WithBlockTime(s.Ctx.BlockTime().Add(10))
	_, err = s.msgServer.DecreaseBond(s.Ctx, &types.MsgDecreaseBond{
		Creator:        defaultSequencerAddress,
		DecreaseAmount: sdk.NewInt64Coin(bondDenom, 10),
	})
	s.Require().NoError(err)
	// try to decrease the bond again - should err as below minbond
	s.Ctx = s.Ctx.WithBlockHeight(s.Ctx.BlockHeight() + 1).WithBlockTime(s.Ctx.BlockTime().Add(10))
	_, err = s.msgServer.DecreaseBond(s.Ctx, &types.MsgDecreaseBond{
		Creator:        defaultSequencerAddress,
		DecreaseAmount: sdk.NewInt64Coin(bondDenom, 10),
	})
	s.Require().ErrorIs(err, types.ErrInsufficientBond)
}
