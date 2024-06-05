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
func (k Keeper) VerifyAndRecordGenesisTransfer(ctx sdk.Context, rollappID string, ix int, n int) error {

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.TransferGenesisKeyPrefix))

	nKey := types.TransferGenesisNumTotalKey(rollappID)
	if store.Has(nKey) {
		nExistingBz := store.Get(nKey)
		nExisting := sdk.BigEndianToUint64(nExistingBz)
		if uint64(n) != nExisting {
			return errorsmod.Wrapf(dymerror.ErrSequencerProtocolViolation,
				"different num total transfers: got: %d: got previously: %d", n, nExisting)
		}
	}
	store.Set(nKey, sdk.Uint64ToBigEndian(uint64(n)))

	if

	return nil
}

func (k Keeper) GetAllGenesisTransfers(ctx sdk.Context) []types.GenesisTransfers {
	var ret []types.GenesisTransfers
	return ret
}
