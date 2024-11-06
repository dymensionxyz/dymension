package sequencer_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func anyPk(pk cryptotypes.PubKey) *codectypes.Any {
	pkAny, err := codectypes.NewAnyWithValue(pk)
	if err != nil {
		panic(err)
	}
	return pkAny
}

func TestInitGenesis(t *testing.T) {
	timeToTest := time.Now().Round(0).UTC()

	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		SequencerList: []types.Sequencer{
			// rollapp 1
			// bonded - no tokens
			{
				Address:      "rollapp1_addr1",
				RollappId:    "rollapp1",
				Status:       types.Bonded,
				Tokens:       sdk.Coins(nil),
				DymintPubKey: anyPk(ed25519.GenPrivKey().PubKey()),
			},
			// bonded - 100 dym
			{
				Address:      "rollapp1_addr2",
				RollappId:    "rollapp1",
				Status:       types.Bonded,
				Tokens:       sdk.NewCoins(sdk.NewCoin("dym", sdk.NewInt(100))),
				DymintPubKey: anyPk(ed25519.GenPrivKey().PubKey()),
			},

			// unbonded
			{
				Address:      "rollapp1_addr4",
				RollappId:    "rollapp1",
				Status:       types.Unbonded,
				Tokens:       sdk.Coins(nil),
				DymintPubKey: anyPk(ed25519.GenPrivKey().PubKey()),
			},
			// rollapp 2
			{
				Address:      "rollapp2_addr1",
				RollappId:    "rollapp2",
				Status:       types.Bonded,
				Tokens:       sdk.Coins(nil),
				DymintPubKey: anyPk(ed25519.GenPrivKey().PubKey()),
			},

			// rollapp 3
			// proposer with notice period
			{
				Address:          "rollapp3_addr1",
				RollappId:        "rollapp3",
				Status:           types.Bonded,
				Tokens:           sdk.Coins(nil),
				NoticePeriodTime: timeToTest,
				DymintPubKey:     anyPk(ed25519.GenPrivKey().PubKey()),
			},
			// rollapp 4
			// proposer with notice period
			{
				Address:          "rollapp4_addr1",
				RollappId:        "rollapp4",
				Status:           types.Bonded,
				Tokens:           sdk.Coins(nil),
				NoticePeriodTime: timeToTest,
				DymintPubKey:     anyPk(ed25519.GenPrivKey().PubKey()),
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

		GenesisSuccessors: []types.GenesisProposer{
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

		NoticeQueue: []string{"rollapp3_addr1", "rollapp4_addr1"},
	}

	// change the params for assertion
	genesisState.Params.NoticePeriod = 100

	k, ctx := keepertest.SequencerKeeper(t)
	sequencer.InitGenesis(ctx, k, genesisState)

	got := sequencer.ExportGenesis(ctx, k)
	require.NotNil(t, got)

	require.Equal(t, genesisState.Params, got.Params)
	require.ElementsMatch(t, genesisState.SequencerList, got.SequencerList)
	require.ElementsMatch(t, genesisState.GenesisProposers, got.GenesisProposers)
	require.ElementsMatch(t, genesisState.GenesisSuccessors, got.GenesisSuccessors)
	require.ElementsMatch(t, genesisState.NoticeQueue, got.NoticeQueue)
}
