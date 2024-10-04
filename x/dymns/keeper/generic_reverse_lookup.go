package keeper

import (
	"slices"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenericAddReverseLookupRecord is a utility method that help to add a reverse lookup record.
func (k Keeper) GenericAddReverseLookupRecord(
	ctx sdk.Context,
	key []byte, newElement string,
	marshaller func([]string) []byte,
	unMarshaller func([]byte) []string,
) error {
	var modifiedRecord []string

	store := ctx.KVStore(k.storeKey)
	bz := store.Get(key)
	if bz != nil {
		existingRecord := unMarshaller(bz)

		if slices.Contains(existingRecord, newElement) {
			// already exist
			return nil
		}

		modifiedRecord = append(existingRecord, newElement)
	} else {
		modifiedRecord = []string{newElement}
	}

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
	key []byte, elementToRemove string,
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
	modifiedRecord := slices.DeleteFunc(existingRecord, func(r string) bool {
		return r == elementToRemove
	})

	if len(existingRecord) == len(modifiedRecord) {
		// not found
		return nil
	}

	if len(modifiedRecord) == 0 {
		// no more, remove record
		store.Delete(key)
		return nil
	}

	// just for safety, sort the records
	slices.SortFunc(modifiedRecord, func(a, b string) int {
		return strings.Compare(a, b)
	})

	bz = marshaller(modifiedRecord)
	store.Set(key, bz)

	return nil
}
