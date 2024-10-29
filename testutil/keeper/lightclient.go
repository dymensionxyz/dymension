package keeper

import (
	"context"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	commitmenttypes "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"

	cometbftdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/math"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"
)

const (
	Alice = "dym1wg8p6j0pxpnsvhkwfu54ql62cnrumf0v634mft"
)

func LightClientKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.StoreKey + "_mem")

	db := cometbftdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	sequencerPubKey := ed25519.GenPrivKey().PubKey()
	tmPk, err := codectypes.NewAnyWithValue(sequencerPubKey)
	require.NoError(t, err)

	testSequencer := sequencertypes.Sequencer{
		Address:      Alice,
		DymintPubKey: tmPk,
	}
	nextValHash, err := testSequencer.GetDymintPubKeyHash()
	require.NoError(t, err)
	testSequencers := map[string]sequencertypes.Sequencer{
		Alice: testSequencer,
	}
	testConsensusStates := map[string]map[uint64]exported.ConsensusState{
		"canon-client-id": {
			2: &ibctm.ConsensusState{
				Timestamp:          time.Unix(1724392989, 0),
				Root:               commitmenttypes.NewMerkleRoot([]byte("test2")),
				NextValidatorsHash: nextValHash,
			},
		},
	}
	cs := ibctm.NewClientState("rollapp-wants-canon-client",
		ibctm.NewFractionFromTm(math.Fraction{Numerator: 1, Denominator: 1}),
		time.Hour*24*7*2, time.Hour*24*7*3, time.Minute*10,
		ibcclienttypes.MustParseHeight("1-2"), commitmenttypes.GetSDKSpecs(), []string{},
	)
	testGenesisClients := map[string]exported.ClientState{
		"canon-client-id": cs,
	}

	mockIBCKeeper := NewMockIBCClientKeeper(testConsensusStates, testGenesisClients)
	mockSequencerKeeper := NewMockSequencerKeeper(testSequencers)
	mockRollappKeeper := NewMockRollappKeeper()
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		mockIBCKeeper,
		mockSequencerKeeper,
		mockRollappKeeper,
	)

	ctx := sdk.NewContext(stateStore, cometbftproto.Header{}, false, log.NewNopLogger())

	return k, ctx
}

type MockIBCCLientKeeper struct {
	clientConsensusState map[string]map[uint64]exported.ConsensusState
	clientStates         map[string]exported.ClientState
}

// ClientStore implements types.IBCClientKeeperExpected.
func (m *MockIBCCLientKeeper) ClientStore(ctx sdk.Context, clientID string) storetypes.KVStore {
	panic("unimplemented")
}

func NewMockIBCClientKeeper(
	clientCS map[string]map[uint64]exported.ConsensusState,
	genesisClients map[string]exported.ClientState,
) *MockIBCCLientKeeper {
	return &MockIBCCLientKeeper{
		clientConsensusState: clientCS,
		clientStates:         genesisClients,
	}
}

func (m *MockIBCCLientKeeper) GetClientConsensusState(ctx sdk.Context, clientID string, height exported.Height) (exported.ConsensusState, bool) {
	cs, ok := m.clientConsensusState[clientID][height.GetRevisionHeight()]
	return cs, ok
}

func (m *MockIBCCLientKeeper) GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool) {
	cs, ok := m.clientStates[clientID]
	return cs, ok
}

func (m *MockIBCCLientKeeper) IterateClientStates(ctx sdk.Context, prefix []byte, cb func(clientID string, cs exported.ClientState) bool) {
	for clientID, cs := range m.clientStates {
		if cb(clientID, cs) {
			break
		}
	}
}

func (m *MockIBCCLientKeeper) ConsensusStateHeights(c context.Context, req *ibcclienttypes.QueryConsensusStateHeightsRequest) (*ibcclienttypes.QueryConsensusStateHeightsResponse, error) {
	heights := []ibcclienttypes.Height{
		ibcclienttypes.NewHeight(1, 2),
	}
	return &ibcclienttypes.QueryConsensusStateHeightsResponse{
		ConsensusStateHeights: heights,
	}, nil
}

type MockSequencerKeeper struct {
	sequencers map[string]sequencertypes.Sequencer
}

// GetProposer implements types.SequencerKeeperExpected.
func (m *MockSequencerKeeper) GetProposer(ctx sdk.Context, rollappId string) (val sequencertypes.Sequencer, found bool) {
	panic("unimplemented")
}

func (m *MockSequencerKeeper) UnbondingTime(ctx sdk.Context) (res time.Duration) {
	return types.DefaultExpectedCanonicalClientParams().UnbondingPeriod
}

func NewMockSequencerKeeper(sequencers map[string]sequencertypes.Sequencer) *MockSequencerKeeper {
	return &MockSequencerKeeper{
		sequencers: sequencers,
	}
}

func (m *MockSequencerKeeper) GetSequencer(ctx sdk.Context, seqAddr string) (sequencertypes.Sequencer, bool) {
	seq, ok := m.sequencers[seqAddr]
	return seq, ok
}

func (m *MockSequencerKeeper) GetSequencersByRollapp(ctx sdk.Context, rollappId string) (list []sequencertypes.Sequencer) {
	seqs := make([]sequencertypes.Sequencer, 0, len(m.sequencers))
	for _, seq := range m.sequencers {
		seqs = append(seqs, seq)
	}
	return seqs
}

func (m *MockSequencerKeeper) JailSequencerOnFraud(ctx sdk.Context, seqAddr string) error {
	return nil
}

type MockRollappKeeper struct{}

// GetLatestStateInfo implements types.RollappKeeperExpected.
func (m *MockRollappKeeper) GetLatestStateInfo(ctx sdk.Context, rollappId string) (rollapptypes.StateInfo, bool) {
	panic("unimplemented")
}

func NewMockRollappKeeper() *MockRollappKeeper {
	return &MockRollappKeeper{}
}

func (m *MockRollappKeeper) GetRollapp(ctx sdk.Context, rollappId string) (val rollapptypes.Rollapp, found bool) {
	return rollapptypes.Rollapp{}, false
}

func (m *MockRollappKeeper) FindStateInfoByHeight(ctx sdk.Context, rollappId string, height uint64) (*rollapptypes.StateInfo, error) {
	return nil, nil
}

func (m *MockRollappKeeper) GetStateInfo(ctx sdk.Context, rollappId string, index uint64) (val rollapptypes.StateInfo, found bool) {
	return rollapptypes.StateInfo{}, false
}

func (m *MockRollappKeeper) SetRollapp(ctx sdk.Context, rollapp rollapptypes.Rollapp) {
}

func (m *MockRollappKeeper) HardFork(ctx sdk.Context, rollappID string, fraudHeight uint64) error {
	return nil
}
