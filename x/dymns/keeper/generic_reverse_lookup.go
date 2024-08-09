package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// GenericAddReverseLookupRecord is a utility method that help to add a reverse lookup record.
func (k Keeper) GenericAddReverseLookupRecord(
	ctx sdk.Context,
	key []byte, offerId string,
	marshaller func([]string) []byte,
	unMarshaller func([]byte) []string,
) error {
	modifiedRecord := dymnstypes.StringList{
		offerId,
	}

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(key)
	if bz != nil {
		existingRecord := unMarshaller(bz)

		modifiedRecord = dymnstypes.StringList(existingRecord).Combine(
			modifiedRecord,
		)

		if len(modifiedRecord) == len(existingRecord) {
			// no new mapping to add
			return nil
		}
	}

	modifiedRecord = modifiedRecord.Sort()

	bz = marshaller(modifiedRecord)
	store.Set(key, bz)

	return nil
}

// GenericGetReverseLookupRecord is a utility method that help to get a reverse lookup record.
func (k Keeper) GenericGetReverseLookupRecord(
	ctx sdk.Context, key []byte,
	unMarshaller func([]byte) []string,
) (result []string) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(key)
	if bz != nil {
		result = unMarshaller(bz)
	}

	return
}

// GenericRemoveReverseLookupRecord is a utility method that help to remove a reverse lookup record.
func (k Keeper) GenericRemoveReverseLookupRecord(
	ctx sdk.Context,
	key []byte, offerId string,
	marshaller func([]string) []byte,
	unMarshaller func([]byte) []string,
) error {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(key)
	if bz == nil {
		// no mapping to remove
		return nil
	}

	existingRecord := unMarshaller(bz)

	modifiedRecord := dymnstypes.StringList(existingRecord).Exclude([]string{offerId})

	if len(modifiedRecord) == 0 {
		// no more, remove record
		store.Delete(key)
		return nil
	}

	modifiedRecord = modifiedRecord.Sort()

	bz = marshaller(modifiedRecord)
	store.Set(key, bz)

	return nil
}
