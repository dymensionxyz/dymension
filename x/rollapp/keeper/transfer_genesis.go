package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymerror "github.com/dymensionxyz/dymension/v3/x/common/errors"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// VerifyAndRecordGenesisTransfer takes a transfer 'index' from the rollapp sequencer and book keeps it
// If we have previously seen a different n, we reject it, the sequencer is not following protocol.
// If we have previously seen the same IX already, we reject it, as IBC guarantees exactly once delivery, then the sequencer must not be following protocol
// Once we have recorded n indexes, this rollapp can proceed to the next step of the genesis transfer protocol
// Returns the number of transfers recorded so far (including this one)
func (k Keeper) VerifyAndRecordGenesisTransfer(ctx sdk.Context, rollappID string, ix, nTotal uint64) (uint64, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TransferGenesisKeyPrefix))

	nTotalKey := types.TransferGenesisNumTotalKey(rollappID)
	nKey := types.TransferGenesisNumKey(rollappID)
	if store.Has(nTotalKey) {
		nTotalExistingBz := store.Get(nTotalKey)
		nTotalExisting := sdk.BigEndianToUint64(nTotalExistingBz)
		if nTotal != nTotalExisting {
			return 0, errorsmod.Wrapf(dymerror.ErrProtocolViolation,
				"different num total transfers: got: %d: got previously: %d", nTotal, nTotalExisting)
		}
	} else {
		store.Set(nTotalKey, sdk.Uint64ToBigEndian(nTotal))
		store.Set(nKey, sdk.Uint64ToBigEndian(0))
	}

	ixKey := types.TransferGenesisSetMembershipKey(rollappID, ix)
	if store.Has(ixKey) {
		return 0, errorsmod.Wrapf(dymerror.ErrProtocolViolation,
			"already received genesis transfer: ix: %d", ix)
	}
	if !(0 <= ix && ix < nTotal) {
		return 0, errorsmod.Wrapf(dymerror.ErrProtocolViolation,
			"ix must be less than nTotal: ix: %d: nTotal: %d", ix, nTotal)
	}

	nBz := store.Get(nKey)
	n := sdk.BigEndianToUint64(nBz)
	n++
	store.Set(nKey, sdk.Uint64ToBigEndian(n))
	return n, nil
}

func (k Keeper) EnableTransfers(ctx sdk.Context, rollappID string) {
	ra := k.MustGetRollapp(ctx, rollappID)
	ra.GenesisState.TransfersEnabled = true
	k.SetRollapp(ctx, ra)
}

func (k Keeper) GetAllGenesisTransfers(ctx sdk.Context) []types.GenesisTransfers {
	var ret []types.GenesisTransfers
	// TODO: impl
	return ret
}
