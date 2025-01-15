package keeper

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
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
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	cometbftdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cometbft/cometbft/libs/math"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"
)

const (
	CanonClientID  = "canon"
	DefaultRollapp = "default"
)

var Alice = func() sequencertypes.Sequencer {
	ret := sequencertypes.NewTestSequencer(ed25519.GenPrivKey().PubKey())
	ret.Status = sequencertypes.Bonded
	ret.RollappId = DefaultRollapp
	return ret
}()

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

	seqs := map[string]*sequencertypes.Sequencer{
		Alice.Address: &Alice,
	}
	consStates := map[string]map[uint64]exported.ConsensusState{
		CanonClientID: {
			2: &ibctm.ConsensusState{
				Timestamp:          time.Unix(1724392989, 0),
				Root:               commitmenttypes.NewMerkleRoot([]byte("test2")),
				NextValidatorsHash: Alice.MustValsetHash(),
			},
		},
	}
	cs := ibctm.NewClientState("rollapp-wants-canon-client",
		ibctm.NewFractionFromTm(math.Fraction{Numerator: 1, Denominator: 1}),
		time.Hour*24*7*2, time.Hour*24*7*3, time.Minute*10,
		ibcclienttypes.MustParseHeight("1-2"), commitmenttypes.GetSDKSpecs(), []string{},
	)
	genesisClients := map[string]exported.ClientState{
		CanonClientID: cs,
	}

	mockIBCKeeper := NewMockIBCClientKeeper(consStates, genesisClients)
	mockSequencerKeeper := NewMockSequencerKeeper(seqs)
	mockRollappKeeper := NewMockRollappKeeper()
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		mockIBCKeeper,
		nil,
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
	sequencers map[string]*sequencertypes.Sequencer
}

func (m *MockSequencerKeeper) SequencerByDymintAddr(ctx sdk.Context, addr cryptotypes.Address) (sequencertypes.Sequencer, error) {
	for _, s := range m.sequencers {
		if bytes.Equal(s.MustProposerAddr(), addr) {
			return *s, nil
		}
	}
	return sequencertypes.Sequencer{}, gerrc.ErrNotFound
}

func (m *MockSequencerKeeper) RealSequencer(ctx sdk.Context, addr string) (sequencertypes.Sequencer, error) {
	seq, ok := m.sequencers[addr]
	var err error
	if !ok {
		err = gerrc.ErrNotFound
	}
	return *seq, err
}

func (m *MockSequencerKeeper) RollappSequencers(ctx sdk.Context, rollappId string) (list []sequencertypes.Sequencer) {
	seqs := make([]sequencertypes.Sequencer, 0, len(m.sequencers))
	for _, seq := range m.sequencers {
		seqs = append(seqs, *seq)
	}
	return seqs
}

// GetProposer implements types.SequencerKeeperExpected.
func (m *MockSequencerKeeper) GetProposer(ctx sdk.Context, rollappId string) (val sequencertypes.Sequencer) {
	panic("unimplemented")
}

func NewMockSequencerKeeper(sequencers map[string]*sequencertypes.Sequencer) *MockSequencerKeeper {
	return &MockSequencerKeeper{
		sequencers: sequencers,
	}
}

type MockRollappKeeper struct{}

func (m *MockRollappKeeper) IsFirstHeightOfLatestFork(ctx sdk.Context, rollappId string, revision, height uint64) bool {
	return false
}

func (m *MockRollappKeeper) GetLatestHeight(ctx sdk.Context, rollappId string) (uint64, bool) {
	panic("implement me")
}

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
