package keepers

import (
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	feemarkettypes "github.com/evmos/ethermint/x/feemarket/types"
)

// GenerateKeys generates new keys (KV Store, Transient store, and memory store).
func (a *AppKeepers) GenerateKeys() {
	// Define what keys will be used in the cosmos-sdk key/value store.
	// Cosmos-SDK modules each have a "key" that allows the application to reference what they've stored on the chain.
	a.keys = KVStoreKeys

	// Define transient store keys
	a.tkeys = sdk.NewTransientStoreKeys(paramstypes.TStoreKey, evmtypes.TransientKey, feemarkettypes.TransientKey)

	// MemKeys are for information that is stored only in RAM.
	a.memKeys = sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)
}

// GetSubspace gets existing substore from keeper.
func (a *AppKeepers) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := a.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// GetKVStoreKeys gets KV Store keys.
func (a *AppKeepers) GetKVStoreKeys() map[string]*storetypes.KVStoreKey {
	return a.keys
}

// GetTransientStoreKey gets Transient Store keys.
func (a *AppKeepers) GetTransientStoreKey() map[string]*storetypes.TransientStoreKey {
	return a.tkeys
}

// GetMemoryStoreKey get memory Store keys.
func (a *AppKeepers) GetMemoryStoreKey() map[string]*storetypes.MemoryStoreKey {
	return a.memKeys
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (a *AppKeepers) GetKey(storeKey string) *storetypes.KVStoreKey {
	return a.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (a *AppKeepers) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return a.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (a *AppKeepers) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return a.memKeys[storeKey]
}
