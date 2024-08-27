package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

type Keeper struct {
	cdc             codec.BinaryCodec
	storeKey        storetypes.StoreKey
	ibcClientKeeper types.IBCClientKeeperExpected
	sequencerKeeper types.SequencerKeeperExpected
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ibcKeeper types.IBCClientKeeperExpected,
	sequencerKeeper types.SequencerKeeperExpected,
) *Keeper {
	k := &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		ibcClientKeeper: ibcKeeper,
		sequencerKeeper: sequencerKeeper,
	}
	return k
}

// GetSeqeuncerHash returns the seqeuncer's tendermint public key hash
func (k Keeper) GetSeqeuncerHash(ctx sdk.Context, sequencerAddr string) ([]byte, error) {
	seq, found := k.sequencerKeeper.GetSequencer(ctx, sequencerAddr)
	if !found {
		return nil, fmt.Errorf("sequencer not found")
	}
	return seq.GetDymintPubKeyHash()
}

func (k Keeper) GetSequencerPubKey(ctx sdk.Context, sequencerAddr string) ([]byte, error) {
	seq, found := k.sequencerKeeper.GetSequencer(ctx, sequencerAddr)
	if !found {
		return nil, fmt.Errorf("sequencer not found")
	}
	return seq.GetDymintPubKeyBytes()
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetCanonicalClient(ctx sdk.Context, rollappId string) (string, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.RollappClientKey(rollappId))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

func (k Keeper) SetCanonicalClient(ctx sdk.Context, rollappId string, clientID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.RollappClientKey(rollappId), []byte(clientID))
	store.Set(types.CanonicalClientKey(clientID), []byte(rollappId))
}

func (k Keeper) SetConsensusStateSigner(ctx sdk.Context, clientID string, height uint64, sequencer []byte) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ConsensusStateSignerKeyByClientID(clientID, height), sequencer)
}

func (k Keeper) GetConsensusStateSigner(ctx sdk.Context, clientID string, height uint64) ([]byte, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ConsensusStateSignerKeyByClientID(clientID, height))
	if bz == nil {
		return []byte{}, false
	}
	return bz, true
}

func (k Keeper) GetRollappForClientID(ctx sdk.Context, clientID string) (string, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CanonicalClientKey(clientID))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}
