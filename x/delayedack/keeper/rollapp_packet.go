package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/delayedack/types"
)

func (k Keeper) SetRollappPacket(ctx sdk.Context, rollappID string, rollappPacket types.RollappPacket) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappPacketKeyPrefix))
	b := k.cdc.MustMarshal(&rollappPacket)
	store.Set(types.RollappPacketKey(
		rollappID,
		ctx.BlockHeight(),
		*rollappPacket.Packet,
	), b)
}

// GetRollappPacketsForRollappAtHeight returns a list of  rollappPackets for a rollapp at packet height
func (k Keeper) ListRollappPacketsForRollappAtHeight(
	ctx sdk.Context,
	rollappId string,
	packetHeight uint64,

) (list []types.RollappPacket) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RollappPacketKeyPrefix))

	// Build the prefix which is composed of the rollappID and the packetHeight
	var prefix []byte
	prefix = append(prefix, []byte(rollappId)...)
	prefix = append(prefix, []byte("/")...)
	prefix = append(prefix, []byte(fmt.Sprint(packetHeight))...)
	prefix = append(prefix, []byte("/")...)

	iterator := sdk.KVStorePrefixIterator(store, prefix)

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.RollappPacket
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return list
}
