package keeper

import (
	"bytes"
	"fmt"

	"github.com/cometbft/cometbft/libs/log"
	tmprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"

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
	rollappKeeper   types.RollappKeeperExpected
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ibcKeeper types.IBCClientKeeperExpected,
	sequencerKeeper types.SequencerKeeperExpected,
	rollappKeeper types.RollappKeeperExpected,
) *Keeper {
	k := &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		ibcClientKeeper: ibcKeeper,
		sequencerKeeper: sequencerKeeper,
		rollappKeeper:   rollappKeeper,
	}
	return k
}

// GetSequencerHash returns the seqeuncer's tendermint public key hash
func (k Keeper) GetSequencerHash(ctx sdk.Context, sequencerAddr string) ([]byte, error) {
	seq, found := k.sequencerKeeper.GetSequencer(ctx, sequencerAddr)
	if !found {
		return nil, fmt.Errorf("sequencer not found")
	}
	return seq.GetDymintPubKeyHash()
}

func (k Keeper) GetSequencerPubKey(ctx sdk.Context, sequencerAddr string) (tmprotocrypto.PublicKey, error) {
	seq, found := k.sequencerKeeper.GetSequencer(ctx, sequencerAddr)
	if !found {
		return tmprotocrypto.PublicKey{}, fmt.Errorf("sequencer not found")
	}
	return seq.GetCometPubKey()
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetSequencerFromValHash(ctx sdk.Context, rollappID string, blockValHash []byte) (string, error) {
	sequencerList := k.sequencerKeeper.GetSequencersByRollapp(ctx, rollappID)
	for _, seq := range sequencerList {
		seqHash, err := seq.GetDymintPubKeyHash()
		if err != nil {
			return "", err
		}
		if bytes.Equal(seqHash, blockValHash) {
			return seq.Address, nil
		}
	}
	return "", types.ErrSequencerNotFound
}

// SetConsenusStateSigner sets block valHash for the given height of the client
func (k Keeper) SetConsensusStateValHash(ctx sdk.Context, clientID string, height uint64, blockValHash []byte) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ConsensusStateValhashKeyByClientID(clientID, height), blockValHash)
}

func (k Keeper) RemoveConsensusStateValHash(ctx sdk.Context, clientID string, height uint64) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.ConsensusStateValhashKeyByClientID(clientID, height))
}

// GetConsensusStateValHash returns the block valHash for the given height of the client
func (k Keeper) GetConsensusStateValHash(ctx sdk.Context, clientID string, height uint64) ([]byte, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ConsensusStateValhashKeyByClientID(clientID, height))
	if bz == nil {
		return nil, false
	}
	return bz, true
}

func (k Keeper) GetAllConsensusStateSigners(ctx sdk.Context) (signers []types.ConsensusStateSigner) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.ConsensusStateValhashKey)
	defer iterator.Close() // nolint: errcheck
	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()
		clientID, height := types.ParseConsensusStateValhashKey(key)
		signers = append(signers, types.ConsensusStateSigner{
			IbcClientId:  clientID,
			Height:       height,
			BlockValHash: string(iterator.Value()),
		})
	}
	return
}

func (k Keeper) GetRollappForClientID(ctx sdk.Context, clientID string) (string, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CanonicalClientKey(clientID))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}
