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
	ibcClientKeeper types.IBCClientKeeper
	sequencerKeeper types.SequencerKeeperExpected
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ibcKeeper types.IBCClientKeeper,
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

func (k Keeper) BeginCanonicalLightClientRegistration(ctx sdk.Context, rollappId string, clientID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.CanonicalLightClientRegistrationKey(rollappId), []byte(clientID))
}

func (k Keeper) GetCanonicalLightClientRegistration(ctx sdk.Context, rollappId string) (string, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CanonicalLightClientRegistrationKey(rollappId))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

func (k Keeper) ClearCanonicalLightClientRegistration(ctx sdk.Context, rollappId string) {
	store := ctx.KVStore(k.storeKey)
	store.Delete(types.CanonicalLightClientRegistrationKey(rollappId))
}

func (k Keeper) SetConsensusStateSigner(ctx sdk.Context, clientID string, height uint64, sequencer string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.ConsensusStateSignerKeyByClientID(clientID, height), []byte(sequencer))
}

func (k Keeper) GetConsensusStateSigner(ctx sdk.Context, clientID string, height uint64) (string, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ConsensusStateSignerKeyByClientID(clientID, height))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

func (k Keeper) GetRollappForClientID(ctx sdk.Context, clientID string) (string, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CanonicalClientKey(clientID))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}
