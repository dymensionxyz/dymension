package eibc

import (
	"encoding/base64"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/keeper"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	for _, demandOrder := range genState.DemandOrders {
		// Create a copy of demandOrder to avoid reusing the same memory address
		demandOrderCopy := demandOrder

		// Decode base64 tracking_packet_key if it exists
		if demandOrderCopy.TrackingPacketKey != "" {
			decodedKey, err := base64.StdEncoding.DecodeString(demandOrderCopy.TrackingPacketKey)
			if err != nil {
				panic(fmt.Errorf("failed to decode tracking_packet_key: %w", err))
			}
			demandOrderCopy.TrackingPacketKey = string(decodedKey)
		}

		err := k.SetDemandOrder(ctx, &demandOrderCopy)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	// Add the demand orders
	allDemandOrders, err := k.ListAllDemandOrders(ctx)
	if err != nil {
		panic(err)
	}

	genesis.DemandOrders = make([]types.DemandOrder, len(allDemandOrders))
	for i, order := range allDemandOrders {
		// Create a copy to avoid modifying the original
		orderCopy := *order

		// Base64 encode tracking_packet_key if it exists
		if orderCopy.TrackingPacketKey != "" {
			encodedKey := base64.StdEncoding.EncodeToString([]byte(orderCopy.TrackingPacketKey))
			orderCopy.TrackingPacketKey = encodedKey
		}

		genesis.DemandOrders[i] = orderCopy
	}

	return genesis
}
