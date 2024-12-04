package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	keeper "github.com/dymensionxyz/dymension/v3/x/streamer/keeper"
)

// Operation weights for simulating the module
const (
	OpWeightBeginBlock = "op_weight_begin_block"
	OpWeightEndBlock   = "op_weight_end_block"

	DefaultWeightBeginBlock = 100
	DefaultWeightEndBlock   = 100
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	k keeper.Keeper,
) simulation.WeightedOperations {
	var (
		weightBeginBlock int
		weightEndBlock   int
	)

	appParams.GetOrGenerate(cdc, OpWeightBeginBlock, &weightBeginBlock, nil,
		func(*rand.Rand) { weightBeginBlock = DefaultWeightBeginBlock })
	appParams.GetOrGenerate(cdc, OpWeightEndBlock, &weightEndBlock, nil,
		func(*rand.Rand) { weightEndBlock = DefaultWeightEndBlock })

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightBeginBlock,
			SimulateMsgBeginBlocker(k),
		),
		simulation.NewWeightedOperation(
			weightEndBlock,
			SimulateMsgEndBlocker(k),
		),
	}
}

