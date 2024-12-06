package simulation

import (
	"math/rand"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	dymsimtypes "github.com/dymensionxyz/dymension/v3/simulation/types"
	"github.com/dymensionxyz/dymension/v3/utils/ukey"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
)

const (
	WeightMsgCreateSequencer           = 50
	WeightMsgIncreaseBond              = 50
	WeightMsgDecreaseBond              = 50
	WeightMsgUnbond                    = 50
	WeightMsgKickProposer              = 30
	WeightMsgUpdateOptInStatus         = 20
	WeightMsgUpdateRewardAddress       = 10
	WeightMsgUpdateWhitelistedRelayers = 10
)

type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	types.AccountKeeper
}

type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	types.BankKeeper
}

type RollappKeeper interface {
	types.RollappKeeper
}

type Keepers struct {
	Acc  AccountKeeper
	Bank BankKeeper
	Roll RollappKeeper
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
	var wCreate, wIncBond, wDecBond, wUnbond, wKick, wOptIn, wRew, wRelayers int

	f.AppParams.GetOrGenerate(
		f.Cdc, "sequencer_create",
		&wCreate, nil, func(r *rand.Rand) { wCreate = WeightMsgCreateSequencer },
	)
	f.AppParams.GetOrGenerate(
		f.Cdc, "sequencer_increase_bond",
		&wIncBond, nil, func(r *rand.Rand) { wIncBond = WeightMsgIncreaseBond },
	)
	f.AppParams.GetOrGenerate(
		f.Cdc, "sequencer_decrease_bond",
		&wDecBond, nil, func(r *rand.Rand) { wDecBond = WeightMsgDecreaseBond },
	)
	f.AppParams.GetOrGenerate(
		f.Cdc, "sequencer_unbond",
		&wUnbond, nil, func(r *rand.Rand) { wUnbond = WeightMsgUnbond },
	)
	f.AppParams.GetOrGenerate(
		f.Cdc, "sequencer_kick_proposer",
		&wKick, nil, func(r *rand.Rand) { wKick = WeightMsgKickProposer },
	)
	f.AppParams.GetOrGenerate(
		f.Cdc, "sequencer_update_opt_in",
		&wOptIn, nil, func(r *rand.Rand) { wOptIn = WeightMsgUpdateOptInStatus },
	)
	f.AppParams.GetOrGenerate(
		f.Cdc, "sequencer_update_reward_addr",
		&wRew, nil, func(r *rand.Rand) { wRew = WeightMsgUpdateRewardAddress },
	)
	f.AppParams.GetOrGenerate(
		f.Cdc, "sequencer_update_relayers",
		&wRelayers, nil, func(r *rand.Rand) { wRelayers = WeightMsgUpdateWhitelistedRelayers },
	)

	protoCdc := codec.NewProtoCodec(codectypes.NewInterfaceRegistry())

	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(wCreate, f.simulateMsgCreateSequencer(protoCdc)),
		//simulation.NewWeightedOperation(wIncBond, f.simulateMsgIncreaseBond(protoCdc)),
		//simulation.NewWeightedOperation(wDecBond, f.simulateMsgDecreaseBond(protoCdc)),
		//simulation.NewWeightedOperation(wUnbond, f.simulateMsgUnbond(protoCdc)),
		//simulation.NewWeightedOperation(wKick, f.simulateMsgKickProposer(protoCdc)),
		simulation.NewWeightedOperation(wOptIn, f.simulateMsgUpdateOptInStatus(protoCdc)),
		simulation.NewWeightedOperation(wRew, f.simulateMsgUpdateRewardAddress(protoCdc)),
		simulation.NewWeightedOperation(wRelayers, f.simulateMsgUpdateWhitelistedRelayers(protoCdc)),
	}
}

func keyAny(pk cryptotypes.PubKey) *codectypes.Any {
	pkAny, err := codectypes.NewAnyWithValue(pk)
	if err != nil {
		panic(err)
	}
	return pkAny
}

func (f OpFactory) simulateMsgCreateSequencer(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp,
		ctx sdk.Context, accs []simtypes.Account, _ string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Fetch an existing rollapp from state
		rollapps := f.k.Roll.GetAllRollapps(ctx)
		if len(rollapps) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, "MsgCreateSequencer", "no rollapps"), nil, nil
		}
		rollapp := dymsimtypes.RandChoice(r, rollapps)

		creator, _ := simtypes.RandomAcc(r, accs)
		bondAmt := math.NewInt(1_000_000)
		if f.k.Bank.SpendableCoins(ctx, creator.Address).AmountOf(sdk.DefaultBondDenom).LT(bondAmt) {
			return simtypes.NoOpMsg(types.ModuleName, "MsgCreateSequencer", "not enough funds"), nil, nil
		}
		if _, err := f.RealSequencer(ctx, creator.Address.String()); err != nil {
			return simtypes.NoOpMsg(types.ModuleName, "MsgCreateSequencer",
				"sequencer already exists"), nil, nil
		}

		msg := &types.MsgCreateSequencer{
			Creator:   creator.Address.String(),
			RollappId: rollapp.RollappId,
			Metadata: types.SequencerMetadata{
				Moniker:        "myseq",
				Details:        "",
				P2PSeeds:       nil,
				Rpcs:           []string{"https://rpc.example.com"},  // at least one URL
				EvmRpcs:        []string{"https://evm.example.com"},  // at least one URL
				RestApiUrls:    []string{"https://rest.example.com"}, // at least one URL
				ExplorerUrl:    "",
				GenesisUrls:    nil, // empty allowed, no validateURLs call
				ContactDetails: nil,
				ExtraData:      nil,
				Snapshots:      nil,
				GasPrice:       nil,
			},
			Bond:         rollapptypes.DefaultParams().MinSequencerBondGlobal,
			DymintPubKey: keyAny(ukey.RandomTMPubKey()),
		}

		return f.deliverTx(r, app, ctx, cdc, msg, creator, accs)
	}
}

func (f OpFactory) simulateMsgIncreaseBond(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp,
		ctx sdk.Context, accs []simtypes.Account, _ string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		seqs := f.AllSequencers(ctx)
		if len(seqs) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, "MsgIncreaseBond", "no sequencers"), nil, nil
		}

		seq := dymsimtypes.RandChoice(r, seqs)
		creatorAddr, err := sdk.AccAddressFromBech32(seq.Address)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, "MsgIncreaseBond", "invalid addr"), nil, nil
		}
		acc, found := simtypes.FindAccount(accs, creatorAddr)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, "MsgIncreaseBond", "seq account not found"), nil, nil
		}

		spendable := f.k.Bank.SpendableCoins(ctx, acc.Address).AmountOf(sdk.DefaultBondDenom)
		if spendable.IsZero() {
			return simtypes.NoOpMsg(types.ModuleName, "MsgIncreaseBond", "no funds"), nil, nil
		}
		amt := simtypes.RandIntBetween(r, 1, int(spendable.Int64())+1)

		msg := &types.MsgIncreaseBond{
			Creator:   acc.Address.String(),
			AddAmount: sdk.NewCoin(sdk.DefaultBondDenom, math.NewInt(int64(amt))),
		}

		return f.deliverTx(r, app, ctx, cdc, msg, acc, accs)
	}
}

func (f OpFactory) simulateMsgDecreaseBond(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp,
		ctx sdk.Context, accs []simtypes.Account, _ string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		seqs := f.AllSequencers(ctx)
		if len(seqs) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, "MsgDecreaseBond", "no sequencers"), nil, nil
		}
		seq := dymsimtypes.RandChoice(r, seqs)
		if seq.Sentinel() {
			return simtypes.NoOpMsg(types.ModuleName, "MsgDecreaseBond", "sentinel"), nil, nil
		}
		creatorAddr, err := sdk.AccAddressFromBech32(seq.Address)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, "MsgDecreaseBond", "invalid addr"), nil, nil
		}
		acc, found := simtypes.FindAccount(accs, creatorAddr)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, "MsgDecreaseBond", "acc not found"), nil, nil
		}
		// Check if seq is proposer or successor
		proposer := f.Keeper.GetProposer(ctx, seq.RollappId)
		successor := f.Keeper.GetSuccessor(ctx, seq.RollappId)
		if seq.Address == proposer.Address || seq.Address == successor.Address {
			return simtypes.NoOpMsg(types.ModuleName, "MsgUnbond", "sequencer is proposer or successor"), nil, nil
		}

		tokens := seq.TokensCoin().Amount
		if tokens.LTE(math.OneInt()) {
			return simtypes.NoOpMsg(types.ModuleName, "MsgDecreaseBond", "no room to decrease"), nil, nil
		}
		decAmt := tokens.QuoRaw(2)
		msg := &types.MsgDecreaseBond{
			Creator:        seq.Address,
			DecreaseAmount: sdk.NewCoin(seq.TokensCoin().Denom, decAmt),
		}

		return f.deliverTx(r, app, ctx, cdc, msg, acc, accs)
	}
}

func (f OpFactory) simulateMsgUnbond(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp,
		ctx sdk.Context, accs []simtypes.Account, _ string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		seqs := f.AllSequencers(ctx)
		if len(seqs) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, "MsgUnbond", "no sequencers"), nil, nil
		}
		seq := dymsimtypes.RandChoice(r, seqs)
		if seq.Sentinel() {
			return simtypes.NoOpMsg(types.ModuleName, "MsgUnbond", "sentinel"), nil, nil
		}

		// Check if seq is proposer or successor
		proposer := f.Keeper.GetProposer(ctx, seq.RollappId)
		successor := f.Keeper.GetSuccessor(ctx, seq.RollappId)
		if seq.Address == proposer.Address || seq.Address == successor.Address {
			return simtypes.NoOpMsg(types.ModuleName, "MsgUnbond", "sequencer is proposer or successor"), nil, nil
		}

		creatorAddr, err := sdk.AccAddressFromBech32(seq.Address)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, "MsgUnbond", "invalid addr"), nil, nil
		}
		acc, found := simtypes.FindAccount(accs, creatorAddr)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, "MsgUnbond", "acc not found"), nil, nil
		}

		msg := &types.MsgUnbond{Creator: seq.Address}
		return f.deliverTx(r, app, ctx, cdc, msg, acc, accs)
	}
}

func (f OpFactory) simulateMsgKickProposer(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp,
		ctx sdk.Context, accs []simtypes.Account, _ string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Need a bonded sequencer to kick another proposer
		seqs := f.AllSequencers(ctx)
		var candidates []types.Sequencer
		for _, s := range seqs {
			if s.Bonded() && !s.Sentinel() {
				candidates = append(candidates, s)
			}
		}
		if len(candidates) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, "MsgKickProposer", "no bonded seq"), nil, nil
		}
		kicker := dymsimtypes.RandChoice(r, candidates)
		kickerAddr, err := sdk.AccAddressFromBech32(kicker.Address)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, "MsgKickProposer", "invalid addr"), nil, nil
		}
		acc, found := simtypes.FindAccount(accs, kickerAddr)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, "MsgKickProposer", "acc not found"), nil, nil
		}

		msg := &types.MsgKickProposer{Creator: acc.Address.String()}
		return f.deliverTx(r, app, ctx, cdc, msg, acc, accs)
	}
}

func (f OpFactory) simulateMsgUpdateOptInStatus(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp,
		ctx sdk.Context, accs []simtypes.Account, _ string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		seqs := f.AllSequencers(ctx)
		if len(seqs) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, "MsgUpdateOptInStatus", "no seq"), nil, nil
		}
		seq := dymsimtypes.RandChoice(r, seqs)
		if seq.Sentinel() {
			return simtypes.NoOpMsg(types.ModuleName, "MsgUpdateOptInStatus", "sentinel"), nil, nil
		}
		creatorAddr, err := sdk.AccAddressFromBech32(seq.Address)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, "MsgUpdateOptInStatus", "invalid addr"), nil, nil
		}
		acc, found := simtypes.FindAccount(accs, creatorAddr)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, "MsgUpdateOptInStatus", "acc not found"), nil, nil
		}

		optIn := r.Intn(2) == 0
		msg := &types.MsgUpdateOptInStatus{
			Creator: acc.Address.String(),
			OptedIn: optIn,
		}

		return f.deliverTx(r, app, ctx, cdc, msg, acc, accs)
	}
}

func (f OpFactory) simulateMsgUpdateRewardAddress(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp,
		ctx sdk.Context, accs []simtypes.Account, _ string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		seqs := f.AllSequencers(ctx)
		if len(seqs) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, "MsgUpdateRewardAddress", "no seq"), nil, nil
		}
		seq := dymsimtypes.RandChoice(r, seqs)
		if seq.Sentinel() {
			return simtypes.NoOpMsg(types.ModuleName, "MsgUpdateRewardAddress", "sentinel"), nil, nil
		}

		creatorAddr, err := sdk.AccAddressFromBech32(seq.Address)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, "MsgUpdateRewardAddress", "invalid addr"), nil, nil
		}
		acc, found := simtypes.FindAccount(accs, creatorAddr)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, "MsgUpdateRewardAddress", "acc not found"), nil, nil
		}

		newAddr, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgUpdateRewardAddress{
			Creator:    acc.Address.String(),
			RewardAddr: newAddr.Address.String(),
		}

		return f.deliverTx(r, app, ctx, cdc, msg, acc, accs)
	}
}

func (f OpFactory) simulateMsgUpdateWhitelistedRelayers(cdc *codec.ProtoCodec) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp,
		ctx sdk.Context, accs []simtypes.Account, _ string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		seqs := f.AllSequencers(ctx)
		if len(seqs) == 0 {
			return simtypes.NoOpMsg(types.ModuleName, "MsgUpdateWhitelistedRelayers", "no seq"), nil, nil
		}
		seq := dymsimtypes.RandChoice(r, seqs)
		if seq.Sentinel() {
			return simtypes.NoOpMsg(types.ModuleName, "MsgUpdateWhitelistedRelayers", "sentinel"), nil, nil
		}

		creatorAddr, err := sdk.AccAddressFromBech32(seq.Address)
		if err != nil {
			return simtypes.NoOpMsg(types.ModuleName, "MsgUpdateWhitelistedRelayers", "invalid addr"), nil, nil
		}
		acc, found := simtypes.FindAccount(accs, creatorAddr)
		if !found {
			return simtypes.NoOpMsg(types.ModuleName, "MsgUpdateWhitelistedRelayers", "acc not found"), nil, nil
		}

		// random subset of accounts as relayers
		numRelayers := r.Intn(3) + 1
		var relayers []string
		for i := 0; i < numRelayers; i++ {
			ar, _ := simtypes.RandomAcc(r, accs)
			relayers = append(relayers, ar.Address.String())
		}
		msg := &types.MsgUpdateWhitelistedRelayers{
			Creator:  acc.Address.String(),
			Relayers: relayers,
		}

		return f.deliverTx(r, app, ctx, cdc, msg, acc, accs)
	}
}

func (f OpFactory) deliverTx(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
	cdc *codec.ProtoCodec, msg sdk.Msg, simAccount simtypes.Account, accs []simtypes.Account,
) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
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
