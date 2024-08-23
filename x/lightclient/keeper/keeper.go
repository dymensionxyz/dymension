package keeper

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/log"

	tmprotocrypto "github.com/cometbft/cometbft/proto/tendermint/crypto"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

type Keeper struct {
	cdc             codec.BinaryCodec
	storeKey        storetypes.StoreKey
	ibcClientKeeper types.IBCClientKeeperExpected
	sequencerKeeper types.SequencerKeeperExpected
	accountKeepr    types.AccountKeeperExpected
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ibcKeeper types.IBCClientKeeperExpected,
	sequencerKeeper types.SequencerKeeperExpected,
	accountKeeper types.AccountKeeperExpected,
) *Keeper {
	k := &Keeper{
		cdc:             cdc,
		storeKey:        storeKey,
		ibcClientKeeper: ibcKeeper,
		sequencerKeeper: sequencerKeeper,
		accountKeepr:    accountKeeper,
	}
	return k
}

func (k Keeper) GetTmPubkeyAsBytes(ctx sdk.Context, sequencerAddr string) ([]byte, error) {
	tmPk, err := k.GetTmPubkey(ctx, sequencerAddr)
	if err != nil {
		return nil, err
	}
	tmPkBytes, err := tmPk.Marshal()
	return tmPkBytes, err
}

func (k Keeper) GetTmPubkey(ctx sdk.Context, sequencerAddr string) (tmprotocrypto.PublicKey, error) {
	acc, err := sdk.AccAddressFromBech32(sequencerAddr)
	if err != nil {
		return tmprotocrypto.PublicKey{}, err
	}
	pk, err := k.accountKeepr.GetPubKey(ctx, acc)
	if err != nil {
		return tmprotocrypto.PublicKey{}, err
	}
	tmPk, err := cryptocodec.ToTmProtoPublicKey(pk)
	if err != nil {
		return tmprotocrypto.PublicKey{}, err
	}
	return tmPk, nil
}

func (k Keeper) getAddress(tmPubkeyBz []byte) (string, error) {
	var tmpk tmprotocrypto.PublicKey
	err := tmpk.Unmarshal(tmPubkeyBz)
	if err != nil {
		return "", err
	}
	pubkey, err := cryptocodec.FromTmProtoPublicKey(tmpk)
	if err != nil {
		return "", err
	}
	acc := sdk.AccAddress(pubkey.Address().Bytes())
	return acc.String(), nil
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
