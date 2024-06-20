package keeper

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// VerifyAndRecordGenesisTransfer takes a transfer 'index' from the rollapp sequencer and book keeps it
// If we have previously seen a different n, we reject it, the sequencer is not following protocol.
// If we have previously seen the same IX already, we reject it, as IBC guarantees exactly once delivery, then the sequencer must not be following protocol
// Once we have recorded n indexes, this rollapp can proceed to the next step of the genesis transfer protocol
// Returns the number of transfers recorded so far (including this one)
func (k Keeper) VerifyAndRecordGenesisTransfer(ctx sdk.Context, rollappID string, nTotal uint64) (uint64, error) {
	ra := k.MustGetRollapp(ctx, rollappID)
	if ra.GenesisState.TransfersEnabled {
		// Could plausibly occur if a chain sends too many genesis transfers (not matching their memo)
		// or if a chain which registered with the bridge enabled tries to send some genesis transfers
		return 0, errorsmod.Wrap(gerrc.ErrFault, "received genesis transfer but all bridge transfers are already enabled")
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TransferGenesisMapKeyPrefix))

	nKey := types.TransferGenesisNumKey(rollappID)
	nTotalKey := types.TransferGenesisNumTotalKey(rollappID)

	n := uint64(0)

	if store.Has(nTotalKey) {
		nTotalExistingBz := store.Get(nTotalKey)
		nTotalExisting := sdk.BigEndianToUint64(nTotalExistingBz)
		if nTotal != nTotalExisting {
			return 0, errorsmod.Wrapf(gerrc.ErrFault, "different num total transfers: got: %d: got previously: %d", nTotal, nTotalExisting)
		}
		nBz := store.Get(nKey)
		n = sdk.BigEndianToUint64(nBz)
	}

	n++
	store.Set(nTotalKey, sdk.Uint64ToBigEndian(nTotal))
	store.Set(nKey, sdk.Uint64ToBigEndian(n))
	return n, nil
}

func (k Keeper) EnableTransfers(ctx sdk.Context, rollappID string) {
	ra := k.MustGetRollapp(ctx, rollappID)
	ra.GenesisState.TransfersEnabled = true
	k.SetRollapp(ctx, ra)
	ctx.EventManager().EmitEvent(sdk.NewEvent(types.EventTypeTransferGenesisTransfersEnabled,
		sdk.NewAttribute(types.AttributeKeyRollappId, rollappID),
	))
}

func (k Keeper) SetGenesisTransfers(ctx sdk.Context, transfers []types.GenesisTransfers) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TransferGenesisMapKeyPrefix))

	for _, transfer := range transfers {

		nKey := types.TransferGenesisNumKey(transfer.RollappID)
		nTotalKey := types.TransferGenesisNumTotalKey(transfer.RollappID)
		store.Set(nTotalKey, sdk.Uint64ToBigEndian(transfer.NumTotal))
		store.Set(nKey, sdk.Uint64ToBigEndian(transfer.NumReceived))
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
		ret = append(ret, x)
	}

	return ret
}
