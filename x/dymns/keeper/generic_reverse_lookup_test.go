package keeper_test

import (
	"fmt"
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/stretchr/testify/require"
)

var keyTestReverseLookup = []byte("test-reverse-lookup")

func TestKeeper_GenericAddGetRemoveReverseLookupRecord(t *testing.T) {
	dk, _, _, ctx := testkeeper.DymNSKeeper(t)

	type testEntity struct {
		getter  func(ctx sdk.Context, key []byte) []string
		adder   func(ctx sdk.Context, key []byte, value string)
		remover func(ctx sdk.Context, key []byte, value string)
	}

	pseudoMarshaller := func(list []string) []byte {
		return dk.Codec().MustMarshal(&dymnstypes.ReverseLookupDymNames{
			DymNames: list,
		})
	}
	pseudoUnMarshaller := func(bz []byte) []string {
		var record dymnstypes.ReverseLookupDymNames
		dk.Codec().MustUnmarshal(bz, &record)
		return record.DymNames
	}

	genericTE := testEntity{
		getter: func(ctx sdk.Context, key []byte) []string {
			return dk.GenericGetReverseLookupRecord(ctx, key, pseudoUnMarshaller)
		},
		adder: func(ctx sdk.Context, key []byte, value string) {
			err := dk.GenericAddReverseLookupRecord(
				ctx,
				key, value,
				pseudoMarshaller, pseudoUnMarshaller,
			)
			require.NoError(t, err)
		},
		remover: func(ctx sdk.Context, key []byte, value string) {
			err := dk.GenericRemoveReverseLookupRecord(
				ctx,
				key, value,
				pseudoMarshaller, pseudoUnMarshaller,
			)
			require.NoError(t, err)
		},
	}

	dymNameTE := testEntity{
		getter: func(ctx sdk.Context, key []byte) []string {
			return dk.GenericGetReverseLookupDymNamesRecord(ctx, key).DymNames
		},
		adder: func(ctx sdk.Context, key []byte, value string) {
			err := dk.GenericAddReverseLookupDymNamesRecord(ctx, key, value)
			require.NoError(t, err)
		},
		remover: func(ctx sdk.Context, key []byte, value string) {
			err := dk.GenericRemoveReverseLookupDymNamesRecord(ctx, key, value)
			require.NoError(t, err)
		},
	}

	boIdsTE := testEntity{
		getter: func(ctx sdk.Context, key []byte) []string {
			return dk.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key).OrderIds
		},
		adder: func(ctx sdk.Context, key []byte, value string) {
			err := dk.GenericAddReverseLookupBuyOrderIdsRecord(ctx, key, value)
			require.NoError(t, err)
		},
		remover: func(ctx sdk.Context, key []byte, value string) {
			err := dk.GenericRemoveReverseLookupBuyOrderIdRecord(ctx, key, value)
			require.NoError(t, err)
		},
	}

	tests := []struct {
		name     string
		testFunc func(t *testing.T, te testEntity, ctx sdk.Context, dk dymnskeeper.Keeper)
	}{
		{
			name: "add - should able to add new record",
			testFunc: func(t *testing.T, te testEntity, ctx sdk.Context, dk dymnskeeper.Keeper) {
				records := te.getter(ctx, keyTestReverseLookup)
				require.Empty(t, records)

				te.adder(ctx, keyTestReverseLookup, "test")

				records = te.getter(ctx, keyTestReverseLookup)
				require.Equal(t, []string{"test"}, records)
			},
		},
		{
			name: "add - should able to append new record",
			testFunc: func(t *testing.T, te testEntity, ctx sdk.Context, dk dymnskeeper.Keeper) {
				te.adder(ctx, keyTestReverseLookup, "test")

				records := te.getter(ctx, keyTestReverseLookup)
				require.Equal(t, []string{"test"}, records)

				te.adder(ctx, keyTestReverseLookup, "test2")

				records = te.getter(ctx, keyTestReverseLookup)
				require.Equal(t, []string{"test", "test2"}, records)
			},
		},
		{
			name: "add - should able to add duplicated record but not persist into store",
			testFunc: func(t *testing.T, te testEntity, ctx sdk.Context, dk dymnskeeper.Keeper) {
				te.adder(ctx, keyTestReverseLookup, "test")
				te.adder(ctx, keyTestReverseLookup, "test2")

				records := te.getter(ctx, keyTestReverseLookup)
				require.Equal(t, []string{"test", "test2"}, records)

				te.adder(ctx, keyTestReverseLookup, "test2") // duplicated

				records = te.getter(ctx, keyTestReverseLookup)
				require.Equal(t, []string{"test", "test2"}, records) // still the same
			},
		},
		{
			name: "add - list must be sorted before persist",
			testFunc: func(t *testing.T, te testEntity, ctx sdk.Context, dk dymnskeeper.Keeper) {
				te.adder(ctx, keyTestReverseLookup, "test3")
				te.adder(ctx, keyTestReverseLookup, "test")

				records := te.getter(ctx, keyTestReverseLookup)
				require.Equal(t, []string{"test", "test3"}, records)

				te.adder(ctx, keyTestReverseLookup, "test2")

				records = te.getter(ctx, keyTestReverseLookup)
				require.Equal(t, []string{"test", "test2", "test3"}, records)
			},
		},
		{
			name: "get - returns empty when getting non-exist record",
			testFunc: func(t *testing.T, te testEntity, ctx sdk.Context, dk dymnskeeper.Keeper) {
				records := te.getter(ctx, keyTestReverseLookup)
				require.Empty(t, records)
			},
		},
		{
			name: "get - returns correct list of records",
			testFunc: func(t *testing.T, te testEntity, ctx sdk.Context, dk dymnskeeper.Keeper) {
				te.adder(ctx, keyTestReverseLookup, "test3")
				te.adder(ctx, keyTestReverseLookup, "test2")
				te.adder(ctx, keyTestReverseLookup, "test1")

				records := te.getter(ctx, keyTestReverseLookup)
				require.Equal(t, []string{"test1", "test2", "test3"}, records)
			},
		},
		{
			name: "get - result is ordered",
			testFunc: func(t *testing.T, te testEntity, ctx sdk.Context, dk dymnskeeper.Keeper) {
				te.adder(ctx, keyTestReverseLookup, "test3")
				te.adder(ctx, keyTestReverseLookup, "test2")
				te.adder(ctx, keyTestReverseLookup, "test1")

				records := te.getter(ctx, keyTestReverseLookup)
				require.Equal(t, []string{"test1", "test2", "test3"}, records)
			},
		},
		{
			name: "remove - able to remove non-existing record without error",
			testFunc: func(t *testing.T, te testEntity, ctx sdk.Context, dk dymnskeeper.Keeper) {
				te.remover(ctx, keyTestReverseLookup, "test3")
			},
		},
		{
			name: "remove - able to remove record not-existing in the list without error",
			testFunc: func(t *testing.T, te testEntity, ctx sdk.Context, dk dymnskeeper.Keeper) {
				te.adder(ctx, keyTestReverseLookup, "test1")
				te.adder(ctx, keyTestReverseLookup, "test2")
				records := te.getter(ctx, keyTestReverseLookup)
				require.Equal(t, []string{"test1", "test2"}, records)

				te.remover(ctx, keyTestReverseLookup, "test3")

				records = te.getter(ctx, keyTestReverseLookup)
				require.Equal(t, []string{"test1", "test2"}, records)
			},
		},
		{
			name: "remove - able to remove record from single element list",
			testFunc: func(t *testing.T, te testEntity, ctx sdk.Context, dk dymnskeeper.Keeper) {
				te.adder(ctx, keyTestReverseLookup, "test")
				records := te.getter(ctx, keyTestReverseLookup)
				require.Equal(t, []string{"test"}, records)

				te.remover(ctx, keyTestReverseLookup, "test")

				records = te.getter(ctx, keyTestReverseLookup)
				require.Empty(t, records)
			},
		},
		{
			name: "remove - able to remove record from multiple elements list",
			testFunc: func(t *testing.T, te testEntity, ctx sdk.Context, dk dymnskeeper.Keeper) {
				te.adder(ctx, keyTestReverseLookup, "test1")
				te.adder(ctx, keyTestReverseLookup, "test2")
				te.adder(ctx, keyTestReverseLookup, "test3")
				records := te.getter(ctx, keyTestReverseLookup)
				require.Equal(t, []string{"test1", "test2", "test3"}, records)

				te.remover(ctx, keyTestReverseLookup, "test2")

				records = te.getter(ctx, keyTestReverseLookup)
				require.Equal(t, []string{"test1", "test3"}, records)
			},
		},
		{
			name: "remove - list must be sorted before persist",
			testFunc: func(t *testing.T, te testEntity, ctx sdk.Context, dk dymnskeeper.Keeper) {
				for no := 100; no <= 999; no++ {
					te.adder(ctx, keyTestReverseLookup, fmt.Sprintf("test%d", no))
				}

				records := te.getter(ctx, keyTestReverseLookup)
				require.Len(t, records, 900)

				te.remover(ctx, keyTestReverseLookup, "test500")

				records = te.getter(ctx, keyTestReverseLookup)
				require.Len(t, records, 899)

				require.True(t, sort.StringsAreSorted(records))
			},
		},
		{
			name: "mix - no collision of records between different keys shares the same head/tail",
			testFunc: func(t *testing.T, te testEntity, ctx sdk.Context, dk dymnskeeper.Keeper) {
				key1 := []byte{0x1, 0x2, 0x3, 0x4}
				key2 := append(key1, 0x5, 0x6)            // share head
				key3 := append([]byte{0x5, 0x6}, key1...) // share tail

				te.adder(ctx, key1, "11")
				te.adder(ctx, key2, "21")
				te.adder(ctx, key3, "31")

				require.Equal(t, []string{"11"}, te.getter(ctx, key1))
				require.Equal(t, []string{"21"}, te.getter(ctx, key2))
				require.Equal(t, []string{"31"}, te.getter(ctx, key3))

				te.adder(ctx, key1, "12")
				te.adder(ctx, key2, "22")
				te.adder(ctx, key3, "32")

				require.Equal(t, []string{"11", "12"}, te.getter(ctx, key1))
				require.Equal(t, []string{"21", "22"}, te.getter(ctx, key2))
				require.Equal(t, []string{"31", "32"}, te.getter(ctx, key3))

				te.remover(ctx, key1, "11")
				te.remover(ctx, key2, "21")
				te.remover(ctx, key3, "31")

				require.Equal(t, []string{"12"}, te.getter(ctx, key1))
				require.Equal(t, []string{"22"}, te.getter(ctx, key2))
				require.Equal(t, []string{"32"}, te.getter(ctx, key3))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			branchedCtx1, _ := ctx.CacheContext()
			tt.testFunc(t, genericTE, branchedCtx1, dk)

			branchedCtx2, _ := ctx.CacheContext()
			tt.testFunc(t, dymNameTE, branchedCtx2, dk)

			branchedCtx3, _ := ctx.CacheContext()
			tt.testFunc(t, boIdsTE, branchedCtx3, dk)
		})
	}
}
