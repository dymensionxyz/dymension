package keeper_test

import (
	"fmt"
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

var keyTestReverseLookup = []byte("test-reverse-lookup")

func (s *KeeperTestSuite) TestKeeper_GenericAddGetRemoveReverseLookupRecord() {
	codec := s.cdc

	type testEntity struct {
		getter  func(ctx sdk.Context, key []byte, s *KeeperTestSuite) []string
		adder   func(ctx sdk.Context, key []byte, value string, s *KeeperTestSuite)
		remover func(ctx sdk.Context, key []byte, value string, s *KeeperTestSuite)
	}

	pseudoMarshaller := func(list []string) []byte {
		return codec.MustMarshal(&dymnstypes.ReverseLookupDymNames{
			DymNames: list,
		})
	}
	pseudoUnMarshaller := func(bz []byte) []string {
		var record dymnstypes.ReverseLookupDymNames
		codec.MustUnmarshal(bz, &record)
		return record.DymNames
	}

	genericTE := testEntity{
		getter: func(ctx sdk.Context, key []byte, s *KeeperTestSuite) []string {
			return s.dymNsKeeper.GenericGetReverseLookupRecord(ctx, key, pseudoUnMarshaller)
		},
		adder: func(ctx sdk.Context, key []byte, value string, s *KeeperTestSuite) {
			err := s.dymNsKeeper.GenericAddReverseLookupRecord(
				ctx,
				key, value,
				pseudoMarshaller, pseudoUnMarshaller,
			)
			s.NoError(err)
		},
		remover: func(ctx sdk.Context, key []byte, value string, s *KeeperTestSuite) {
			err := s.dymNsKeeper.GenericRemoveReverseLookupRecord(
				ctx,
				key, value,
				pseudoMarshaller, pseudoUnMarshaller,
			)
			s.NoError(err)
		},
	}

	dymNameTE := testEntity{
		getter: func(ctx sdk.Context, key []byte, s *KeeperTestSuite) []string {
			return s.dymNsKeeper.GenericGetReverseLookupDymNamesRecord(ctx, key).DymNames
		},
		adder: func(ctx sdk.Context, key []byte, value string, s *KeeperTestSuite) {
			err := s.dymNsKeeper.GenericAddReverseLookupDymNamesRecord(ctx, key, value)
			s.NoError(err)
		},
		remover: func(ctx sdk.Context, key []byte, value string, s *KeeperTestSuite) {
			err := s.dymNsKeeper.GenericRemoveReverseLookupDymNamesRecord(ctx, key, value)
			s.NoError(err)
		},
	}

	boIdsTE := testEntity{
		getter: func(ctx sdk.Context, key []byte, s *KeeperTestSuite) []string {
			return s.dymNsKeeper.GenericGetReverseLookupBuyOrderIdsRecord(ctx, key).OrderIds
		},
		adder: func(ctx sdk.Context, key []byte, value string, s *KeeperTestSuite) {
			err := s.dymNsKeeper.GenericAddReverseLookupBuyOrderIdsRecord(ctx, key, value)
			s.NoError(err)
		},
		remover: func(ctx sdk.Context, key []byte, value string, s *KeeperTestSuite) {
			err := s.dymNsKeeper.GenericRemoveReverseLookupBuyOrderIdRecord(ctx, key, value)
			s.NoError(err)
		},
	}

	tests := []struct {
		name     string
		testFunc func(te testEntity, ctx sdk.Context, s *KeeperTestSuite)
	}{
		{
			name: "add - should able to add new record",
			testFunc: func(te testEntity, ctx sdk.Context, s *KeeperTestSuite) {
				records := te.getter(ctx, keyTestReverseLookup, s)
				s.Empty(records)

				te.adder(ctx, keyTestReverseLookup, "test", s)

				records = te.getter(ctx, keyTestReverseLookup, s)
				s.Equal([]string{"test"}, records)
			},
		},
		{
			name: "add - should able to append new record",
			testFunc: func(te testEntity, ctx sdk.Context, s *KeeperTestSuite) {
				te.adder(ctx, keyTestReverseLookup, "test", s)

				records := te.getter(ctx, keyTestReverseLookup, s)
				s.Equal([]string{"test"}, records)

				te.adder(ctx, keyTestReverseLookup, "test2", s)

				records = te.getter(ctx, keyTestReverseLookup, s)
				s.Equal([]string{"test", "test2"}, records)
			},
		},
		{
			name: "add - should able to add duplicated record but not persist into store",
			testFunc: func(te testEntity, ctx sdk.Context, s *KeeperTestSuite) {
				te.adder(ctx, keyTestReverseLookup, "test", s)
				te.adder(ctx, keyTestReverseLookup, "test2", s)

				records := te.getter(ctx, keyTestReverseLookup, s)
				s.Equal([]string{"test", "test2"}, records)

				te.adder(ctx, keyTestReverseLookup, "test2", s) // duplicated

				records = te.getter(ctx, keyTestReverseLookup, s)
				s.Equal([]string{"test", "test2"}, records) // still the same
			},
		},
		{
			name: "add - list must be sorted before persist",
			testFunc: func(te testEntity, ctx sdk.Context, s *KeeperTestSuite) {
				te.adder(ctx, keyTestReverseLookup, "test3", s)
				te.adder(ctx, keyTestReverseLookup, "test", s)

				records := te.getter(ctx, keyTestReverseLookup, s)
				s.Equal([]string{"test", "test3"}, records)

				te.adder(ctx, keyTestReverseLookup, "test2", s)

				records = te.getter(ctx, keyTestReverseLookup, s)
				s.Equal([]string{"test", "test2", "test3"}, records)
			},
		},
		{
			name: "get - returns empty when getting non-exist record",
			testFunc: func(te testEntity, ctx sdk.Context, s *KeeperTestSuite) {
				records := te.getter(ctx, keyTestReverseLookup, s)
				s.Empty(records)
			},
		},
		{
			name: "get - returns correct list of records",
			testFunc: func(te testEntity, ctx sdk.Context, s *KeeperTestSuite) {
				te.adder(ctx, keyTestReverseLookup, "test3", s)
				te.adder(ctx, keyTestReverseLookup, "test2", s)
				te.adder(ctx, keyTestReverseLookup, "test1", s)

				records := te.getter(ctx, keyTestReverseLookup, s)
				s.Equal([]string{"test1", "test2", "test3"}, records)
			},
		},
		{
			name: "get - result is ordered",
			testFunc: func(te testEntity, ctx sdk.Context, s *KeeperTestSuite) {
				te.adder(ctx, keyTestReverseLookup, "test3", s)
				te.adder(ctx, keyTestReverseLookup, "test2", s)
				te.adder(ctx, keyTestReverseLookup, "test1", s)

				records := te.getter(ctx, keyTestReverseLookup, s)
				s.Equal([]string{"test1", "test2", "test3"}, records)
			},
		},
		{
			name: "remove - able to remove non-existing record without error",
			testFunc: func(te testEntity, ctx sdk.Context, s *KeeperTestSuite) {
				te.remover(ctx, keyTestReverseLookup, "test3", s)
			},
		},
		{
			name: "remove - able to remove record not-existing in the list without error",
			testFunc: func(te testEntity, ctx sdk.Context, s *KeeperTestSuite) {
				te.adder(ctx, keyTestReverseLookup, "test1", s)
				te.adder(ctx, keyTestReverseLookup, "test2", s)
				records := te.getter(ctx, keyTestReverseLookup, s)
				s.Equal([]string{"test1", "test2"}, records)

				te.remover(ctx, keyTestReverseLookup, "test3", s)

				records = te.getter(ctx, keyTestReverseLookup, s)
				s.Equal([]string{"test1", "test2"}, records)
			},
		},
		{
			name: "remove - able to remove record from single element list",
			testFunc: func(te testEntity, ctx sdk.Context, s *KeeperTestSuite) {
				te.adder(ctx, keyTestReverseLookup, "test", s)
				records := te.getter(ctx, keyTestReverseLookup, s)
				s.Equal([]string{"test"}, records)

				te.remover(ctx, keyTestReverseLookup, "test", s)

				records = te.getter(ctx, keyTestReverseLookup, s)
				s.Empty(records)
			},
		},
		{
			name: "remove - able to remove record from multiple elements list",
			testFunc: func(te testEntity, ctx sdk.Context, s *KeeperTestSuite) {
				te.adder(ctx, keyTestReverseLookup, "test1", s)
				te.adder(ctx, keyTestReverseLookup, "test2", s)
				te.adder(ctx, keyTestReverseLookup, "test3", s)
				records := te.getter(ctx, keyTestReverseLookup, s)
				s.Equal([]string{"test1", "test2", "test3"}, records)

				te.remover(ctx, keyTestReverseLookup, "test2", s)

				records = te.getter(ctx, keyTestReverseLookup, s)
				s.Equal([]string{"test1", "test3"}, records)
			},
		},
		{
			name: "remove - list must be sorted before persist",
			testFunc: func(te testEntity, ctx sdk.Context, s *KeeperTestSuite) {
				for no := 100; no <= 999; no++ {
					te.adder(ctx, keyTestReverseLookup, fmt.Sprintf("test%d", no), s)
				}

				records := te.getter(ctx, keyTestReverseLookup, s)
				s.Len(records, 900)

				te.remover(ctx, keyTestReverseLookup, "test500", s)

				records = te.getter(ctx, keyTestReverseLookup, s)
				s.Len(records, 899)

				s.True(sort.StringsAreSorted(records))
			},
		},
		{
			name: "mix - no collision of records between different keys shares the same head/tail",
			testFunc: func(te testEntity, ctx sdk.Context, s *KeeperTestSuite) {
				key1 := []byte{0x1, 0x2, 0x3, 0x4}
				key2 := append(key1, 0x5, 0x6)            // share head
				key3 := append([]byte{0x5, 0x6}, key1...) // share tail

				te.adder(ctx, key1, "11", s)
				te.adder(ctx, key2, "21", s)
				te.adder(ctx, key3, "31", s)

				s.Equal([]string{"11"}, te.getter(ctx, key1, s))
				s.Equal([]string{"21"}, te.getter(ctx, key2, s))
				s.Equal([]string{"31"}, te.getter(ctx, key3, s))

				te.adder(ctx, key1, "12", s)
				te.adder(ctx, key2, "22", s)
				te.adder(ctx, key3, "32", s)

				s.Equal([]string{"11", "12"}, te.getter(ctx, key1, s))
				s.Equal([]string{"21", "22"}, te.getter(ctx, key2, s))
				s.Equal([]string{"31", "32"}, te.getter(ctx, key3, s))

				te.remover(ctx, key1, "11", s)
				te.remover(ctx, key2, "21", s)
				te.remover(ctx, key3, "31", s)

				s.Equal([]string{"12"}, te.getter(ctx, key1, s))
				s.Equal([]string{"22"}, te.getter(ctx, key2, s))
				s.Equal([]string{"32"}, te.getter(ctx, key3, s))
			},
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			branchedCtx1, _ := s.ctx.CacheContext()
			tt.testFunc(genericTE, branchedCtx1, s)

			branchedCtx2, _ := s.ctx.CacheContext()
			tt.testFunc(dymNameTE, branchedCtx2, s)

			branchedCtx3, _ := s.ctx.CacheContext()
			tt.testFunc(boIdsTE, branchedCtx3, s)
		})
	}
}
