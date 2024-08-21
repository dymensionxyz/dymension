package sequencer_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func TestInitGenesis(t *testing.T) {
	timeToTest := time.Now().Round(0).UTC()

	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		SequencerList: []types.Sequencer{
			// rollapp 1
			// bonded - no tokens
			{
				Address:   "rollapp1_addr1",
				RollappId: "rollapp1",
				Status:    types.Bonded,
				Tokens:    sdk.Coins(nil),
			},
			// bonded - 100 dym
			{
				Address:   "rollapp1_addr2",
				RollappId: "rollapp1",
				Status:    types.Bonded,
				Tokens:    sdk.NewCoins(sdk.NewCoin("dym", sdk.NewInt(100))),
			},
			// unbonding
			{
				Address:             "rollapp1_addr3",
				RollappId:           "rollapp1",
				Status:              types.Unbonding,
				Tokens:              sdk.Coins(nil),
				UnbondRequestHeight: 10,
				UnbondTime:          timeToTest,
			},
			// unbonded
			{
				Address:   "rollapp1_addr4",
				RollappId: "rollapp1",
				Status:    types.Unbonded,
				Tokens:    sdk.Coins(nil),
			},
			// rollapp 2
			{
				Address:   "rollapp2_addr1",
				RollappId: "rollapp2",
				Status:    types.Bonded,
				Tokens:    sdk.Coins(nil),
			},
			// unbonding
			{
				Address:             "rollapp2_addr2",
				RollappId:           "rollapp2",
				Status:              types.Unbonding,
				Tokens:              sdk.Coins(nil),
				UnbondRequestHeight: 10,
				UnbondTime:          timeToTest,
			},
			// rollapp 3
			// proposer with notice period
			{
				Address:             "rollapp3_addr1",
				RollappId:           "rollapp3",
				Status:              types.Bonded,
				Tokens:              sdk.Coins(nil),
				UnbondRequestHeight: 20,
				NoticePeriodTime:    timeToTest,
			},
		},
		BondReductions: []types.BondReduction{
			{
				SequencerAddress:   "rollapp1_addr1",
				DecreaseBondAmount: sdk.NewCoin("dym", sdk.NewInt(100)),
				DecreaseBondTime:   timeToTest,
			},
		},
		GenesisProposers: []types.GenesisProposer{
			{
				Address:   "rollapp1_addr1",
				RollappId: "rollapp1",
			},
			{
				Address:   "rollapp3_addr1",
				RollappId: "rollapp3",
			},
			// rollapp2 has no proposer
		},
	}

	// change the params for assertion
	genesisState.Params.NoticePeriod = 100

	k, ctx := keepertest.SequencerKeeper(t)
	sequencer.InitGenesis(ctx, *k, genesisState)

	require.Len(t, k.GetMatureNoticePeriodSequencers(ctx, timeToTest), 1)
	require.Len(t, k.GetMatureUnbondingSequencers(ctx, timeToTest), 2)
	require.Len(t, k.GetAllProposers(ctx), 2)
	require.Len(t, k.GetAllBondReductions(ctx), 1)

	got := sequencer.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.Equal(t, genesisState.Params, got.Params)
	require.ElementsMatch(t, genesisState.SequencerList, got.SequencerList)
	require.ElementsMatch(t, genesisState.GenesisProposers, got.GenesisProposers)
	require.ElementsMatch(t, genesisState.BondReductions, got.BondReductions)
}
