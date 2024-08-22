package keeper

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	commitmenttypes "github.com/cosmos/ibc-go/v7/modules/core/23-commitment/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"

	cometbftdb "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/libs/log"
	cometbftproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"
)

const (
	Alice = "dym1wg8p6j0pxpnsvhkwfu54ql62cnrumf0v634mft"
)

func LightClientKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	storeKey := sdk.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.StoreKey + "_mem")
	transientStoreKey := storetypes.NewTransientStoreKey(types.TransientKey)

	db := cometbftdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db)
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	stateStore.MountStoreWithDB(transientStoreKey, storetypes.StoreTypeTransient, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	testConsensusStates := map[string]map[uint64]exported.ConsensusState{
		"canon-client-id": {
			2: &ibctm.ConsensusState{
				Timestamp:          time.Now().UTC(),
				Root:               commitmenttypes.NewMerkleRoot([]byte("test2")),
				NextValidatorsHash: []byte{156, 132, 96, 43, 190, 214, 140, 148, 216, 119, 98, 162, 97, 120, 115, 32, 39, 223, 114, 56, 224, 180, 80, 228, 190, 243, 9, 248, 190, 33, 188, 23},
			},
		},
	}
	sequencerPubKey := ed25519.GenPrivKey().PubKey()
	testPubKeys := map[string]cryptotypes.PubKey{
		Alice: sequencerPubKey,
	}

	mockIBCKeeper := NewMockIBCClientKeeper(testConsensusStates)
	mockSequencerKeeper := NewMockSequencerKeeper()
	mockAccountKeeper := NewMockAccountKeeper(testPubKeys)
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		mockIBCKeeper,
		mockSequencerKeeper,
		mockAccountKeeper,
	)

	ctx := sdk.NewContext(stateStore, cometbftproto.Header{}, false, log.NewNopLogger())

	return k, ctx
}

type MockIBCCLientKeeper struct {
	clientConsensusState map[string]map[uint64]exported.ConsensusState
}

func NewMockIBCClientKeeper(clientCS map[string]map[uint64]exported.ConsensusState) *MockIBCCLientKeeper {
	return &MockIBCCLientKeeper{
		clientConsensusState: clientCS,
	}
}

func (m *MockIBCCLientKeeper) GetClientConsensusState(ctx sdk.Context, clientID string, height exported.Height) (exported.ConsensusState, bool) {
	cs, ok := m.clientConsensusState[clientID][height.GetRevisionHeight()]
	return cs, ok
}

func (m *MockIBCCLientKeeper) GenerateClientIdentifier(ctx sdk.Context, clientType string) string {
	return ""
}

func (m *MockIBCCLientKeeper) GetClientState(ctx sdk.Context, clientID string) (exported.ClientState, bool) {
	return nil, false
}

type MockSequencerKeeper struct {
}

func NewMockSequencerKeeper() *MockSequencerKeeper {
	return &MockSequencerKeeper{}
}

func (m *MockSequencerKeeper) JailSequencerOnFraud(ctx sdk.Context, seqAddr string) error {
	return nil
}

type MockAccountKeeper struct {
	pubkeys map[string]cryptotypes.PubKey
}

func NewMockAccountKeeper(pubkeys map[string]cryptotypes.PubKey) *MockAccountKeeper {
	return &MockAccountKeeper{
		pubkeys: pubkeys,
	}
}

func (m *MockAccountKeeper) GetPubKey(ctx sdk.Context, addr sdk.AccAddress) (cryptotypes.PubKey, error) {
	pubkey, ok := m.pubkeys[addr.String()]
	if !ok {
		return nil, sdkerrors.ErrUnknownAddress
	}
	return pubkey, nil
}
