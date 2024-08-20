package keeper

import (
	"testing"
	"time"

	cmttypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
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
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/stretchr/testify/require"
)

const (
	Alice = "cosmos1c4k24jzduc365kywrsvf5ujz4ya6mwymy8vq4q"
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

	mockIBCKeeper := NewMockIBCClientKeeper()
	mockSequencerKeeper := NewMockSequencerKeeper()
	mockAccountKeeper := NewMockAccountKeeper()
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
}

func NewMockIBCClientKeeper() *MockIBCCLientKeeper {
	return &MockIBCCLientKeeper{}
}

func (m *MockIBCCLientKeeper) GetClientConsensusState(ctx sdk.Context, clientID string, height exported.Height) (exported.ConsensusState, bool) {
	if clientID == "canon-client-id-no-state" {
		return nil, false
	}
	if clientID == "canon-client-id" && height.GetRevisionHeight() == 2 {
		val := cmttypes.NewMockPV()
		pk, _ := val.GetPubKey()
		cs := ibctm.NewConsensusState(
			time.Now().UTC(),
			commitmenttypes.NewMerkleRoot([]byte("test2")),
			cmttypes.NewValidatorSet([]*cmttypes.Validator{cmttypes.NewValidator(pk, 1)}).Hash(),
		)
		return cs, true
	}
	return nil, false
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

func (m *MockSequencerKeeper) SlashAndJailFraud(ctx sdk.Context, seqAddr string) error {
	return nil
}

type MockAccountKeeper struct {
	pubkey cryptotypes.PubKey
}

func NewMockAccountKeeper() *MockAccountKeeper {
	pubkey := ed25519.GenPrivKey().PubKey()
	return &MockAccountKeeper{
		pubkey: pubkey,
	}
}

func (m *MockAccountKeeper) GetPubKey(ctx sdk.Context, addr sdk.AccAddress) (cryptotypes.PubKey, error) {
	if addr.String() == Alice {
		return m.pubkey, nil
	}
	return nil, sdkerrors.ErrUnknownAddress
}
