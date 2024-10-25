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

			// rollapp 3
			// proposer with notice period
			{
				Address:          "rollapp3_addr1",
				RollappId:        "rollapp3",
				Status:           types.Bonded,
				Tokens:           sdk.Coins(nil),
				NoticePeriodTime: timeToTest,
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

	noticeElapsed, err := k.NoticeElapsedSequencers(ctx, timeToTest)
	require.NoError(t, err)
	require.Len(t, noticeElapsed, 1)
	require.Len(t, k.GetAllProposers(ctx), 2)

	got := sequencer.ExportGenesis(ctx, *k)
	require.NotNil(t, got)
	require.Equal(t, genesisState.Params, got.Params)
	require.ElementsMatch(t, genesisState.SequencerList, got.SequencerList)
	require.ElementsMatch(t, genesisState.GenesisProposers, got.GenesisProposers)
}
