package cmd

import (
	"errors"
	"fmt"
	"path/filepath"

	dbm "github.com/cometbft/cometbft-db"

	cfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/os"
	"github.com/cometbft/cometbft/state"
	"github.com/cometbft/cometbft/store"
)

func loadStateAndBlockStore(config *cfg.Config) (*store.BlockStore, state.Store, error) {
	dbType := dbm.BackendType(config.DBBackend)

	if !os.FileExists(filepath.Join(config.DBDir(), "blockstore.db")) {
		return nil, nil, fmt.Errorf("no blockstore found in %v", config.DBDir())
	}

	// Get BlockStore
	blockStoreDB, err := dbm.NewDB("blockstore", dbType, config.DBDir())
	if err != nil {
		return nil, nil, err
	}
	blockStore := store.NewBlockStore(blockStoreDB)

	if !os.FileExists(filepath.Join(config.DBDir(), "state.db")) {
		return nil, nil, fmt.Errorf("no statestore found in %v", config.DBDir())
	}

	// Get StateStore
	stateDB, err := dbm.NewDB("state", dbType, config.DBDir())
	if err != nil {
		return nil, nil, err
	}
	stateStore := state.NewStore(stateDB, state.StoreOptions{
		DiscardABCIResponses: config.Storage.DiscardABCIResponses,
	})

	return blockStore, stateStore, nil
}

func getTendermintState(config *cfg.Config) error {
	// use the parsed config to load the block and state store
	blockStore, stateStore, err := loadStateAndBlockStore(config)
	if err != nil {
		return err
	}
	defer func() {
		_ = blockStore.Close()
		_ = stateStore.Close()
	}()

	// Read the data from the KVStore
	fmt.Println("LOADING STATE")

	state, err := stateStore.Load()
	if err != nil {
		return err
	}
	if state.IsEmpty() {
		return errors.New("no state found")
	}
	fmt.Printf("%+v\n", state)

	bh := blockStore.Height()
	latestBlock := blockStore.LoadBlock(bh)
	if latestBlock == nil {
		return errors.New("no block found for latest height")
	}
	fmt.Printf("%+v\n", latestBlock)

	if bh != state.LastBlockHeight {
		// printing block for state height
		fmt.Println("LOADING BLOCK FOR STATE HEIGHT")
		block := blockStore.LoadBlock(state.LastBlockHeight)
		if block == nil {
			return errors.New("no block found for state height")
		}
		fmt.Printf("%+v\n", block)
	}

	return nil
}
