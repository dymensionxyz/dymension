package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/derr"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// VerifyAndRecordGenesisTransfer takes a transfer 'index' from the rollapp sequencer and book keeps it
// If we have previously seen a different n, we reject it, the sequencer is not following protocol.
// If we have previously seen the same IX already, we reject it, as IBC guarantees exactly once delivery, then the sequencer must not be following protocol
// Once we have recorded n indexes, this rollapp can proceed to the next step of the genesis transfer protocol
// Returns the number of transfers recorded so far (including this one)
func (k Keeper) VerifyAndRecordGenesisTransfer(ctx sdk.Context, rollappID string, ix, nTotal uint64) (uint64, error) {
	ra := k.MustGetRollapp(ctx, rollappID)
	if ra.GenesisState.TransfersEnabled {
		// Could plausibly occur if a chain sends too many genesis transfers (not matching their memo)
		// or if a chain which registered with the bridge enabled tries to send some genesis transfers
		return 0, errorsmod.Wrap(derr.ErrViolatesDymensionRollappStandard, "received genesis transfer but all bridge transfers are already enabled")
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TransferGenesisMapKeyPrefix))

	nKey := types.TransferGenesisNumKey(rollappID)
	nTotalKey := types.TransferGenesisNumTotalKey(rollappID)
	ixKey := types.TransferGenesisSetMembershipKey(rollappID, ix)

	n := uint64(0)
	/*
		We do all the verification first and only write at the end, to make it easier to reason about partial failures
	*/

	if store.Has(nTotalKey) {
		nTotalExistingBz := store.Get(nTotalKey)
		nTotalExisting := sdk.BigEndianToUint64(nTotalExistingBz)
		if nTotal != nTotalExisting {
			return 0, errorsmod.Wrapf(derr.ErrViolatesDymensionRollappStandard, "different num total transfers: got: %d: got previously: %d", nTotal, nTotalExisting)
		}
		nBz := store.Get(nKey)
		n = sdk.BigEndianToUint64(nBz)
	}
	if nTotal <= ix {
		return 0, errorsmod.Wrapf(derr.ErrViolatesDymensionRollappStandard, "ix must be less than nTotal: ix: %d: nTotal: %d", ix, nTotal)
	}
	if store.Has(ixKey) {
		return 0, errorsmod.Wrapf(derr.ErrViolatesDymensionRollappStandard, "already received genesis transfer: ix: %d", ix)
	}

	n++
	store.Set(nTotalKey, sdk.Uint64ToBigEndian(nTotal))
	store.Set(nKey, sdk.Uint64ToBigEndian(n))
	store.Set(ixKey, []byte{})
	return n, nil
}

func (k Keeper) EnableTransfers(ctx sdk.Context, rollappID string) {
	ra := k.MustGetRollapp(ctx, rollappID)
	ra.GenesisState.TransfersEnabled = true
	k.SetRollapp(ctx, ra)
	ctx.EventManager().EmitEvent(transfersEnabledEvent(rollappID))
}

func transfersEnabledEvent(raID string) sdk.Event {
	return sdk.NewEvent(types.EventTypeTransferGenesisTransfersEnabled,
		sdk.NewAttribute(types.AttributeKeyRollappId, raID),
	)
}

func (k Keeper) SetGenesisTransfers(ctx sdk.Context, transfers []types.GenesisTransfers) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TransferGenesisMapKeyPrefix))

	for _, transfer := range transfers {

		nKey := types.TransferGenesisNumKey(transfer.RollappID)
		nTotalKey := types.TransferGenesisNumTotalKey(transfer.RollappID)
		for _, i := range transfer.GetReceived() {
			ixKey := types.TransferGenesisSetMembershipKey(transfer.RollappID, i)
			store.Set(nTotalKey, sdk.Uint64ToBigEndian(transfer.NumTotal))
			store.Set(nKey, sdk.Uint64ToBigEndian(transfer.NumReceived))
			store.Set(ixKey, []byte{})
		}

	}
}

func (k Keeper) GetAllGenesisTransfers(ctx sdk.Context) []types.GenesisTransfers {
	var ret []types.GenesisTransfers

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TransferGenesisMapKeyPrefix))

	rollapps := k.GetAllRollapps(ctx)

	for _, ra := range rollapps {

		raID := ra.RollappId
		nTotalKey := types.TransferGenesisNumTotalKey(raID)
		nTotalBz := store.Get(nTotalKey)
		nTotal := sdk.BigEndianToUint64(nTotalBz)
		nKey := types.TransferGenesisNumKey(raID)
		nBz := store.Get(nKey)
		n := sdk.BigEndianToUint64(nBz)
		x := types.GenesisTransfers{
			RollappID:   raID,
			NumTotal:    nTotal,
			NumReceived: n,
		}
		for ix := range nTotal {
			ixKey := types.TransferGenesisSetMembershipKey(raID, ix)
			if store.Has(ixKey) {
				x.Received = append(x.Received, ix)
			}
		}

		ret = append(ret, x)
	}

	return ret
}

// DelGenesisTransfers deletes bookkeeping for one rollapp, it's not needed for correctness, but it's good to prune state
// TODO: need to justify correct pruning property for this and the other state, in all possibilities (e.g. fraud recovery)
func (k Keeper) DelGenesisTransfers(ctx sdk.Context, raID string) { // TODO: needs to be public?

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TransferGenesisMapKeyPrefix))

	nTotalKey := types.TransferGenesisNumTotalKey(raID)
	nTotalBz := store.Get(nTotalKey)
	nTotal := sdk.BigEndianToUint64(nTotalBz)
	nKey := types.TransferGenesisNumKey(raID)
	store.Delete(nTotalBz)
	store.Delete(nKey)

	for ix := range nTotal {
		store.Delete(types.TransferGenesisSetMembershipKey(raID, ix))
	}
}
