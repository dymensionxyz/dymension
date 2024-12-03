package simulation

import (
	"math/rand"
	"time"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	dymsimtypes "github.com/dymensionxyz/dymension/v3/simulation/types"
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// Simulation operation weights constants
const (
	OpWeightMsgCreateStream          = "op_weight_msg_create_stream"
	OpWeightMsgCreateSponsoredStream = "op_weight_msg_create_sponsored_stream"
	OpWeightMsgUpdateStream          = "op_weight_msg_update_stream"
	OpWeightMsgTerminateStream       = "op_weight_msg_terminate_stream"

	DefaultWeightMsgCreateStream          = 50
	DefaultWeightMsgCreateSponsoredStream = 50
	DefaultWeightMsgUpdateStream          = 30
	DefaultWeightMsgTerminateStream       = 20
)

// WeightedOperations returns all the operations from the module with their respective weights
func WeightedOperations(
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	ik dymsimtypes.IncentivesKeeper,
	k Keeper,
) simulation.WeightedOperations {
	var (
		weightMsgCreateStream          int
		weightMsgCreateSponsoredStream int
		weightMsgUpdateStream          int
		weightMsgTerminateStream       int
	)

	appParams.GetOrGenerate(cdc, OpWeightMsgCreateStream, &weightMsgCreateStream, nil,
		func(*rand.Rand) { weightMsgCreateStream = DefaultWeightMsgCreateStream })
	appParams.GetOrGenerate(cdc, OpWeightMsgCreateSponsoredStream, &weightMsgCreateSponsoredStream, nil,
		func(*rand.Rand) { weightMsgCreateSponsoredStream = DefaultWeightMsgCreateSponsoredStream })
	appParams.GetOrGenerate(cdc, OpWeightMsgUpdateStream, &weightMsgUpdateStream, nil,
		func(*rand.Rand) { weightMsgUpdateStream = DefaultWeightMsgUpdateStream })
	appParams.GetOrGenerate(cdc, OpWeightMsgTerminateStream, &weightMsgTerminateStream, nil,
		func(*rand.Rand) { weightMsgTerminateStream = DefaultWeightMsgTerminateStream })

	protoCdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgCreateStream,
			SimulateMsgCreateStream(protoCdc, ak, bk, ik, k),
		),
		simulation.NewWeightedOperation(
			weightMsgCreateSponsoredStream,
			SimulateMsgCreateSponsoredStream(protoCdc, ak, bk, ik, k),
		),
		simulation.NewWeightedOperation(
			weightMsgUpdateStream,
			SimulateMsgUpdateStream(protoCdc, ak, bk, ik, k),
		),
		simulation.NewWeightedOperation(
			weightMsgTerminateStream,
			SimulateMsgTerminateStream(protoCdc, ak, bk, ik, k),
		),
	}
}

// SimulateMsgCreateStream generates a random stream creation
func SimulateMsgCreateStream(
	cdc *codec.ProtoCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	ik dymsimtypes.IncentivesKeeper,
	k Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount := accs[r.Intn(len(accs))]
		
		// Get random gauges
		gauges := dymsimtypes.RandomGaugeSubset(ctx, r, ik)
		if len(gauges) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateStream, "no gauges"), nil, nil
		}

		// Generate random distribution records
		records := make([]types.DistrRecord, len(gauges))
		totalWeight := math.ZeroInt()
		for i, gauge := range gauges {
			// Generate random weight between 1 and MaxInt64/len(gauges) to avoid overflow
			weight, err := dymsimtypes.RandIntBetween(r, math.OneInt(), math.NewIntFromUint64(^uint64(0)/uint64(len(gauges))))
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateStream, "failed to generate weight"), nil, nil
			}
			records[i] = types.DistrRecord{
				GaugeId: gauge.Id,
				Weight:  weight,
			}
			totalWeight = totalWeight.Add(weight)
		}

		// Generate random coins
		spendable := bk.SpendableCoins(ctx, simAccount.Address)
		if spendable.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateStream, "no spendable coins"), nil, nil
		}
		
		// Pick a random coin and amount
		coin := spendable[r.Intn(len(spendable))]
		amt, err := dymsimtypes.RandIntBetween(r, math.OneInt(), coin.Amount)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateStream, "failed to generate amount"), nil, nil
		}
		coins := sdk.NewCoins(sdk.NewCoin(coin.Denom, amt))

		// Random start time between now and 1 week in the future
		startTime := ctx.BlockTime().Add(time.Duration(r.Int63n(7*24*60*60)) * time.Second)
		
		// Random number of epochs between 1 and 100
		numEpochs := uint64(r.Int63n(100) + 1)

		// Random epoch identifier
		epochIdentifiers := []string{"day", "week", "month"}
		epochIdentifier := epochIdentifiers[r.Intn(len(epochIdentifiers))]

		streamID, err := k.CreateStream(ctx, coins, records, startTime, epochIdentifier, numEpochs, false)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateStream, err.Error()), nil, nil
		}

		return simtypes.NewOperationMsg(&types.MsgCreateStream{}, true, ""), nil, nil
	}
}

// SimulateMsgCreateSponsoredStream generates a random sponsored stream creation
func SimulateMsgCreateSponsoredStream(
	cdc *codec.ProtoCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	ik dymsimtypes.IncentivesKeeper,
	k Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount := accs[r.Intn(len(accs))]

		// Get random gauges
		gauges := dymsimtypes.RandomGaugeSubset(ctx, r, ik)
		if len(gauges) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateStream, "no gauges"), nil, nil
		}

		// Generate random distribution records - for sponsored streams these don't matter
		// but we need valid ones for the API
		records := make([]types.DistrRecord, len(gauges))
		for i, gauge := range gauges {
			records[i] = types.DistrRecord{
				GaugeId: gauge.Id,
				Weight:  math.OneInt(),
			}
		}

		// Generate random coins similar to non-sponsored case
		spendable := bk.SpendableCoins(ctx, simAccount.Address)
		if spendable.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateStream, "no spendable coins"), nil, nil
		}
		
		coin := spendable[r.Intn(len(spendable))]
		amt, err := dymsimtypes.RandIntBetween(r, math.OneInt(), coin.Amount)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateStream, "failed to generate amount"), nil, nil
		}
		coins := sdk.NewCoins(sdk.NewCoin(coin.Denom, amt))

		startTime := ctx.BlockTime().Add(time.Duration(r.Int63n(7*24*60*60)) * time.Second)
		numEpochs := uint64(r.Int63n(100) + 1)
		epochIdentifiers := []string{"day", "week", "month"}
		epochIdentifier := epochIdentifiers[r.Intn(len(epochIdentifiers))]

		streamID, err := k.CreateStream(ctx, coins, records, startTime, epochIdentifier, numEpochs, true)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateStream, err.Error()), nil, nil
		}

		return simtypes.NewOperationMsg(&types.MsgCreateStream{}, true, ""), nil, nil
	}
}

// SimulateMsgUpdateStream generates a random stream update
func SimulateMsgUpdateStream(
	cdc *codec.ProtoCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	ik dymsimtypes.IncentivesKeeper,
	k Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Get all streams
		streams := k.GetStreams(ctx)
		if len(streams) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateStream, "no streams"), nil, nil
		}

		// Pick a random stream
		stream := streams[r.Intn(len(streams))]

		// Get random gauges
		gauges := dymsimtypes.RandomGaugeSubset(ctx, r, ik)
		if len(gauges) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateStream, "no gauges"), nil, nil
		}

		// Generate random distribution records
		records := make([]types.DistrRecord, len(gauges))
		totalWeight := math.ZeroInt()
		for i, gauge := range gauges {
			weight, err := dymsimtypes.RandIntBetween(r, math.OneInt(), math.NewIntFromUint64(^uint64(0)/uint64(len(gauges))))
			if err != nil {
				return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateStream, "failed to generate weight"), nil, nil
			}
			records[i] = types.DistrRecord{
				GaugeId: gauge.Id,
				Weight:  weight,
			}
			totalWeight = totalWeight.Add(weight)
		}

		// Randomly choose between ReplaceDistrRecords and UpdateDistrRecords
		var err error
		if r.Int()%2 == 0 {
			err = k.ReplaceDistrRecords(ctx, stream.Id, records)
		} else {
			err = k.UpdateDistrRecords(ctx, stream.Id, records)
		}
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateStream, err.Error()), nil, nil
		}

		return simtypes.NewOperationMsg(&types.MsgUpdateStream{}, true, ""), nil, nil
	}
}

// SimulateMsgTerminateStream generates a random stream termination
func SimulateMsgTerminateStream(
	cdc *codec.ProtoCodec,
	ak dymsimtypes.AccountKeeper,
	bk dymsimtypes.BankKeeper,
	ik dymsimtypes.IncentivesKeeper,
	k Keeper,
) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Get all streams
		streams := k.GetStreams(ctx)
		if len(streams) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgTerminateStream, "no streams"), nil, nil
		}

		// Pick a random stream
		stream := streams[r.Intn(len(streams))]

		err := k.TerminateStream(ctx, stream.Id)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgTerminateStream, err.Error()), nil, nil
		}

		return simtypes.NewOperationMsg(&types.MsgTerminateStream{}, true, ""), nil, nil
	}
}
