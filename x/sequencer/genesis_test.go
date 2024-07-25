package sequencer_test

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/stretchr/testify/require"
)

func TestInitGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		SequencerList: []types.Sequencer{
			// rollapp 1
			// bonded - no tokens
			{
				SequencerAddress: "rollapp1_addr1",
				RollappId:        "rollapp1",
				Status:           types.Bonded,
				Tokens:           sdk.Coins(nil),
			},
			// bonded - 100 dym
			{
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
				UnbondTime:          time.Time{}, // todo: set time
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
				SequencerAddress:    "rollapp2_addr1",
				RollappId:           "rollapp2",
				Status:              types.Bonded,
				Tokens:              sdk.Coins(nil),
				UnbondRequestHeight: 0,
				UnbondTime:          time.Time{},
			},
			// rollapp 3
			// proposer with notice period
			{
				SequencerAddress:    "rollapp3_addr1",
				RollappId:           "rollapp3",
				Status:              types.Bonded,
				Tokens:              sdk.Coins(nil),
				UnbondRequestHeight: 20,
				UnbondTime:          time.Time{},
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

	got := sequencer.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.Equal(t, genesisState.Params, got.Params)
	require.ElementsMatch(t, genesisState.SequencerList, got.SequencerList)
}
