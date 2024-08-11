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
<<<<<<< HEAD
				SequencerAddress: "rollapp1_addr1",
				RollappId:        "rollapp1",
				Status:           types.Bonded,
				Tokens:           sdk.Coins(nil),
=======
				Address:  "0",
				Status:   types.Bonded,
				Proposer: true,
>>>>>>> main
			},
			// bonded - 100 dym
			{
<<<<<<< HEAD
				SequencerAddress: "rollapp1_addr2",
				RollappId:        "rollapp1",
				Status:           types.Bonded,
				Tokens:           sdk.NewCoins(sdk.NewCoin("dym", sdk.NewInt(100))),
			},
			// unbonding
			{
				SequencerAddress:    "rollapp1_addr3",
				RollappId:           "rollapp1",
				Status:              types.Unbonding,
				Tokens:              sdk.Coins(nil),
				UnbondRequestHeight: 10,
				UnbondTime:          timeToTest,
			},
			// unbonded
			{
				SequencerAddress: "rollapp1_addr4",
				RollappId:        "rollapp1",
				Status:           types.Unbonded,
				Tokens:           sdk.Coins(nil),
			},
			// rollapp 2
			{
				SequencerAddress: "rollapp2_addr1",
				RollappId:        "rollapp2",
				Status:           types.Bonded,
				Tokens:           sdk.Coins(nil),
			},
			// unbonding
			{
				SequencerAddress:    "rollapp2_addr2",
				RollappId:           "rollapp2",
				Status:              types.Unbonding,
				Tokens:              sdk.Coins(nil),
				UnbondRequestHeight: 10,
				UnbondTime:          timeToTest,
			},
			// rollapp 3
			// proposer with notice period
			{
				SequencerAddress:    "rollapp3_addr1",
				RollappId:           "rollapp3",
				Status:              types.Bonded,
				Tokens:              sdk.Coins(nil),
				UnbondRequestHeight: 20,
				NoticePeriodTime:    timeToTest,
=======
				Address: "1",
				Status:  types.Bonded,
			},
		},
		BondReductions: []types.BondReduction{
			{
				SequencerAddress:   "0",
				DecreaseBondAmount: sdk.NewCoin("dym", sdk.NewInt(100)),
				DecreaseBondTime:   time.Now().UTC(),
>>>>>>> main
			},
		},
		GenesisProposers: []types.GenesisProposer{
			{
				Address:   "rollapp1_addr1",
				RollappId: "rollap1",
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

	got := sequencer.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.Equal(t, genesisState.Params, got.Params)
	require.ElementsMatch(t, genesisState.SequencerList, got.SequencerList)
<<<<<<< HEAD
=======
	require.ElementsMatch(t, genesisState.BondReductions, got.BondReductions)
	// this line is used by starport scaffolding # genesis/test/assert
}

func TestExportGenesis(t *testing.T) {
	params := types.Params{
		MinBond:                 sdk.NewCoin("dym", sdk.NewInt(100)),
		UnbondingTime:           100,
		LivenessSlashMultiplier: sdk.ZeroDec(),
	}
	sequencerList := []types.Sequencer{
		{
			Address:  "0",
			Status:   types.Bonded,
			Proposer: true,
		},
		{
			Address: "1",
			Status:  types.Bonded,
		},
	}
	bondReductions := []types.BondReduction{
		{
			SequencerAddress:   "0",
			DecreaseBondAmount: sdk.NewCoin("dym", sdk.NewInt(100)),
			DecreaseBondTime:   time.Now().UTC(),
		},
	}
	k, ctx := keepertest.SequencerKeeper(t)
	k.SetParams(ctx, params)
	for _, sequencer := range sequencerList {
		k.SetSequencer(ctx, sequencer)
	}
	for _, bondReduction := range bondReductions {
		k.SetDecreasingBondQueue(ctx, bondReduction)
	}

	got := sequencer.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.Equal(t, params, got.Params)
	require.ElementsMatch(t, sequencerList, got.SequencerList)
	require.ElementsMatch(t, bondReductions, got.BondReductions)
>>>>>>> main
}
