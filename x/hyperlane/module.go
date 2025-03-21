package hyperlane

import (
	"context"
	"encoding/json"

	"cosmossdk.io/core/appmodule"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/dymensionxyz/dymension/v3/x/hyperlane/keeper"
)

var (
	_ appmodule.AppModule       = AppModule{}
	_ appmodule.HasBeginBlocker = AppModule{}
	_ appmodule.HasEndBlocker   = AppModule{}
)

// AppModule implements an application module for the hyperlane module.
type AppModule struct {
	cdc    codec.Codec
	keeper *keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper *keeper.Keeper) AppModule {
	return AppModule{
		cdc:    cdc,
		keeper: keeper,
	}
}

// IsOnePerModuleType implements the depinject.OnePerModuleType interface.
func (am AppModule) IsOnePerModuleType() {}

// IsAppModule implements the appmodule.AppModule interface.
func (am AppModule) IsAppModule() {}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// TODO: Register services
}

// BeginBlock returns the begin blocker for the hyperlane module.
func (am AppModule) BeginBlock(ctx context.Context) error {
	return nil
}

// EndBlock returns the end blocker for the hyperlane module.
func (am AppModule) EndBlock(ctx context.Context) error {
	return nil
}

// InitGenesis performs genesis initialization for the hyperlane module.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	// TODO: Initialize genesis state
}

// ExportGenesis returns the exported genesis state as raw bytes for the hyperlane module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	// TODO: Export genesis state
	return nil
}

// ConsensusVersion implements ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }
