package simulation

import (
	"fmt"
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	dymsimtypes "github.com/dymensionxyz/dymension/v3/simulation/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

const (
	WeightCreateRollapp            = 100
	WeightUpdateRollappInformation = 50
	WeightTransferOwnership        = 50
	WeightUpdateState              = 100
	WeightAddApp                   = 50
	WeightUpdateApp                = 50
	WeightRemoveApp                = 50
	WeightFraudProposal            = 20
	WeightMarkObsoleteRollapps     = 20
	WeightForceGenesisInfoChange   = 20
)

type ChannelKeeper interface {
	keeper.ChannelKeeper
}

type SequencerKeeper interface {
	keeper.SequencerKeeper
}

type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	keeper.BankKeeper
}

type CanonicalLightClientKeeper interface {
	keeper.CanonicalLightClientKeeper
}

type TransferKeeper interface {
	keeper.TransferKeeper
}

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
}

type Keepers struct {
	Chan     ChannelKeeper
	Seq      SequencerKeeper
	Bank     BankKeeper
	Light    CanonicalLightClientKeeper
	Transfer TransferKeeper
	Acc      AccountKeeper
}

type OpFactory struct {
	keeper.Keeper
	k Keepers
	module.SimulationState
}

func NewOpFactory(k keeper.Keeper, ks Keepers, simState module.SimulationState) OpFactory {
	return OpFactory{
		Keeper:          k,
		k:               ks,
		SimulationState: simState,
	}
}

// WeightedOperations returns the simulation operations (messages) with weights.
func (f OpFactory) Messages() simulation.WeightedOperations {
	var wCreateRollapp, wUpdateInfo, wTransferOwner, wUpdateState, wAddApp, wUpdateApp, wRemoveApp, wFraud, wMarkObs, wForceGen int

	f.AppParams.GetOrGenerate(f.Cdc, "rollapp_create", &wCreateRollapp, nil, func(r *rand.Rand) { wCreateRollapp = WeightCreateRollapp })
	f.AppParams.GetOrGenerate(f.Cdc, "rollapp_update_info", &wUpdateInfo, nil, func(r *rand.Rand) { wUpdateInfo = WeightUpdateRollappInformation })
	f.AppParams.GetOrGenerate(f.Cdc, "rollapp_transfer_owner", &wTransferOwner, nil, func(r *rand.Rand) { wTransferOwner = WeightTransferOwnership })
	f.AppParams.GetOrGenerate(f.Cdc, "rollapp_update_state", &wUpdateState, nil, func(r *rand.Rand) { wUpdateState = WeightUpdateState })
	f.AppParams.GetOrGenerate(f.Cdc, "rollapp_add_app", &wAddApp, nil, func(r *rand.Rand) { wAddApp = WeightAddApp })
	f.AppParams.GetOrGenerate(f.Cdc, "rollapp_update_app", &wUpdateApp, nil, func(r *rand.Rand) { wUpdateApp = WeightUpdateApp })
	f.AppParams.GetOrGenerate(f.Cdc, "rollapp_remove_app", &wRemoveApp, nil, func(r *rand.Rand) { wRemoveApp = WeightRemoveApp })
	f.AppParams.GetOrGenerate(f.Cdc, "rollapp_fraud_proposal", &wFraud, nil, func(r *rand.Rand) { wFraud = WeightFraudProposal })
	f.AppParams.GetOrGenerate(f.Cdc, "rollapp_mark_obsolete", &wMarkObs, nil, func(r *rand.Rand) { wMarkObs = WeightMarkObsoleteRollapps })
	f.AppParams.GetOrGenerate(f.Cdc, "rollapp_force_genesis_change", &wForceGen, nil, func(r *rand.Rand) { wForceGen = WeightForceGenesisInfoChange })

	protoCdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(wCreateRollapp, f.simulateMsgCreateRollapp(protoCdc)),
		simulation.NewWeightedOperation(wUpdateInfo, f.simulateMsgUpdateRollappInformation(protoCdc)),
		simulation.NewWeightedOperation(wTransferOwner, f.simulateMsgTransferOwnership(protoCdc)),
		simulation.NewWeightedOperation(wUpdateState, f.simulateMsgUpdateState(protoCdc)),
		simulation.NewWeightedOperation(wAddApp, f.simulateMsgAddApp(protoCdc)),
		simulation.NewWeightedOperation(wUpdateApp, f.simulateMsgUpdateApp(protoCdc)),
		simulation.NewWeightedOperation(wRemoveApp, f.simulateMsgRemoveApp(protoCdc)),
		simulation.NewWeightedOperation(wFraud, f.simulateMsgFraudProposal(protoCdc)),
		simulation.NewWeightedOperation(wMarkObs, f.simulateMsgMarkObsoleteRollapps(protoCdc)),
		simulation.NewWeightedOperation(wForceGen, f.simulateMsgForceGenesisInfoChange(protoCdc)),
	}
}

func (f OpFactory) simulateMsgCreateRollapp(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		creator, _ := simtypes.RandomAcc(r, accs)
		rollappId := fmt.Sprintf("rollapp_%s_1-1", simtypes.RandStringOfLength(r, 5)) // chain-id like pattern name_epoc-rev
		minBond := sdk.NewInt64Coin(sdk.DefaultBondDenom, 1e6)
		msg := types.NewMsgCreateRollapp(creator.Address.String(), rollappId, creator.Address.String(), minBond, "alias", types.Rollapp_VMType(1), nil, nil)

		return f.deliverTx(r, app, ctx, accs, cdc, msg, creator)
	}
}

func (f OpFactory) simulateMsgUpdateRollappInformation(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// fetch rollapps
		rollapps := f.GetAllRollapps(ctx)
		if len(rollapps) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, "update_rollapp", "no rollapps"), nil, nil
		}
		ra := dymsimtypes.RandChoice(r, rollapps)
		ownerAddr, err := sdk.AccAddressFromBech32(ra.Owner)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, "update_rollapp", "invalid owner"), nil, nil
		}
		// find sim acc for owner
		ownerAcc, found := simtypes.FindAccount(accs, ownerAddr)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, "update_rollapp", "owner not in accs"), nil, nil
		}

		msg := types.NewMsgUpdateRollappInformation(ownerAcc.Address.String(), ra.RollappId, "", sdk.Coin{}, nil, nil)

		return f.deliverTx(r, app, ctx, accs, cdc, msg, ownerAcc)
	}
}

func (f OpFactory) simulateMsgTransferOwnership(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		rollapps := f.GetAllRollapps(ctx)
		if len(rollapps) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, "transfer_ownership", "no rollapps"), nil, nil
		}
		ra := dymsimtypes.RandChoice(r, rollapps)
		ownerAddr, err := sdk.AccAddressFromBech32(ra.Owner)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, "transfer_ownership", "invalid owner"), nil, nil
		}
		ownerAcc, found := simtypes.FindAccount(accs, ownerAddr)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, "transfer_ownership", "owner not in accs"), nil, nil
		}

		newOwner, _ := simtypes.RandomAcc(r, accs)
		if newOwner.Address.Equals(ownerAcc.Address) {
			return simtypes.NoOpMsg(types.ModuleName, "transfer_ownership", "same owner"), nil, nil
		}

		msg := types.NewMsgTransferOwnership(ownerAcc.Address.String(), newOwner.Address.String(), ra.RollappId)
		return f.deliverTx(r, app, ctx, accs, cdc, msg, ownerAcc)
	}
}

func (f OpFactory) simulateMsgUpdateState(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		rollapps := f.GetAllRollapps(ctx)
		if len(rollapps) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, "update_state", "no rollapps"), nil, nil
		}

		ra := dymsimtypes.RandChoice(r, rollapps)
		ownerAddr, err := sdk.AccAddressFromBech32(ra.Owner)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, "update_state", "invalid owner"), nil, nil
		}
		ownerAcc, found := simtypes.FindAccount(accs, ownerAddr)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, "update_state", "owner not in accs"), nil, nil
		}

		startHeight := uint64(r.Intn(100) + 1)
		numBlocks := uint64(r.Intn(10) + 1)
		BDs := &types.BlockDescriptors{}
		for i := uint64(0); i < numBlocks; i++ {
			BDs.BD = append(BDs.BD, types.BlockDescriptor{
				Height:     startHeight + i,
				StateRoot:  make([]byte, 32),
				Timestamp:  ctx.BlockTime(),
				DrsVersion: 0,
			})
		}

		msg := types.NewMsgUpdateState(ownerAcc.Address.String(), ra.RollappId, "daPath", startHeight, numBlocks, BDs)
		return f.deliverTx(r, app, ctx, accs, cdc, msg, ownerAcc)
	}
}

func (f OpFactory) simulateMsgAddApp(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		rollapps := f.GetAllRollapps(ctx)
		if len(rollapps) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, "add_app", "no rollapps"), nil, nil
		}
		ra := dymsimtypes.RandChoice(r, rollapps)
		ownerAddr, err := sdk.AccAddressFromBech32(ra.Owner)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, "add_app", "invalid owner"), nil, nil
		}
		ownerAcc, found := simtypes.FindAccount(accs, ownerAddr)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, "add_app", "owner not in accs"), nil, nil
		}

		name := fmt.Sprintf("app_%s", simtypes.RandStringOfLength(r, 5))
		msg := types.NewMsgAddApp(ownerAcc.Address.String(), name, ra.RollappId, "desc", "image", "url", 1)
		return f.deliverTx(r, app, ctx, accs, cdc, msg, ownerAcc)
	}
}

func (f OpFactory) simulateMsgUpdateApp(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// For simplicity, no specific app retrieval. If no apps, no-op.
		// If we had a real way to fetch apps, we would. For now, just no-op commonly.
		return simtypes.NoOpMsg(types.ModuleName, "update_app", "no apps known"), nil, nil
	}
}

func (f OpFactory) simulateMsgRemoveApp(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Similarly, no-op due to complexity of fetching an existing app.
		return simtypes.NoOpMsg(types.ModuleName, "remove_app", "no apps known"), nil, nil
	}
}

func (f OpFactory) simulateMsgFraudProposal(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Requires authority (gov)
		return simtypes.NoOpMsg(types.ModuleName, "fraud_proposal", "not simulating gov authority"), nil, nil
	}
}

func (f OpFactory) simulateMsgMarkObsoleteRollapps(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Requires authority (gov)
		return simtypes.NoOpMsg(types.ModuleName, "mark_obsolete_rollapps", "not simulating gov authority"), nil, nil
	}
}

func (f OpFactory) simulateMsgForceGenesisInfoChange(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, _ string) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Requires authority (gov)
		return simtypes.NoOpMsg(types.ModuleName, "force_genesis_info_change", "not simulating gov authority"), nil, nil
	}
}

// Helper function to deliver a tx with random fees
func (f OpFactory) deliverTx(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, cdc *codec.ProtoCodec, msg sdk.Msg, simAccount simtypes.Account) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	txCtx := simulation.OperationInput{
		R:             r,
		App:           app,
		TxGen:         moduletestutil.MakeTestEncodingConfig().TxConfig,
		Cdc:           cdc,
		Msg:           msg,
		MsgType:       sdk.MsgTypeURL(msg),
		SimAccount:    simAccount,
		Context:       ctx,
		AccountKeeper: f.k.Acc,
		Bankkeeper:    f.k.Bank,
		ModuleName:    types.ModuleName,
	}

	opMsg, fOps, err := simulation.GenAndDeliverTxWithRandFees(txCtx)
	if err != nil {
		return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), err.Error()), nil, err
	}
	return opMsg, fOps, err
}
