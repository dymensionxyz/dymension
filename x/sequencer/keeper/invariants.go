package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/invar"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

var invs = invar.NamedFuncsList[Keeper]{
	{"notice", InvariantNotice},
	{"hash-index", InvariantHashIndex},
	{"status", InvariantStatus},
	{"tokens", InvariantTokens},
}

// RegisterInvariants registers the sequencer module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	invs.RegisterInvariants(types.ModuleName, ir, k)
}

// DO NOT DELETE
func AllInvariants(k Keeper) sdk.Invariant {
	return invs.All(types.ModuleName, k)
}

func InvariantNotice(k Keeper) invar.Func {
	return func(ctx sdk.Context) (error, bool) {
		seqs, err := k.NoticeQueue(ctx, nil)
		if err != nil {
			return err, true
		}
		for _, seq := range seqs {
			if !seq.NoticeStarted() {
				return errors.New("sequencer not started notice"), true
			}
		}
		return nil, false
	}
}

func InvariantHashIndex(k Keeper) invar.Func {
	return func(ctx sdk.Context) (error, bool) {
		for _, exp := range k.AllSequencers(ctx) {
			got, err := k.SequencerByDymintAddr(ctx, exp.MustValsetHash())
			if err != nil {
				return err, true
			}
			if got.Address != exp.Address {
				return errors.New("address mismatch"), true
			}
		}
		return nil, false
	}
}

func InvariantStatus(k Keeper) invar.Func {
	return func(ctx sdk.Context) (error, bool) {
		rollapps := k.rollappKeeper.GetAllRollapps(ctx)
		for _, ra := range rollapps {
			proposer := k.GetProposer(ctx, ra.RollappId)
			if !proposer.Bonded() {
				return errors.New("proposer not bonded"), true
			}
			successor := k.GetSuccessor(ctx, ra.RollappId)
			if !successor.Bonded() {
				return errors.New("successor not bonded"), true
			}
			if !proposer.Sentinel() && proposer.Address == successor.Address {
				return errors.New("proposer and successor are the same"), true
			}
			if !successor.Sentinel() && proposer.Sentinel() {
				return errors.New("proposer is sentinel but successor is not"), true
			}
			all := k.RollappSequencers(ctx, ra.RollappId)
			bonded := k.RollappSequencersByStatus(ctx, ra.RollappId, types.Bonded)
			unbonded := k.RollappSequencersByStatus(ctx, ra.RollappId, types.Unbonded)
			if len(all) != len(bonded)+len(unbonded) {
				return errors.New("sequencer by rollapp length is not equal to sum of bonded, and unbonded"), true
			}
		}
		for _, seq := range k.AllProposers(ctx) {
			if !k.IsProposer(ctx, seq) {
				return errors.New("sequencer in proposers query is not proposer"), true
			}
		}
		for _, seq := range k.AllSuccessors(ctx) {
			if !k.IsSuccessor(ctx, seq) {
				return errors.New("sequencer in successor query is not successor"), true
			}
		}
		return nil, false
	}
}

func InvariantTokens(k Keeper) invar.Func {
	return func(ctx sdk.Context) (error, bool) {

		for _, seq := range k.AllSequencers(ctx) {
			if err := seq.ValidateBasic(); err != nil {
				return err, true
			}
			if err := k.validBondDenom(ctx, seq.TokensCoin()); err != nil {
				return err, true
			}
			if seq.TokensCoin().Amount.IsNegative() {
				return errors.New("negative seq tokens"), true
			}
		}

		total := sdk.NewCoin(k.bondDenom(ctx), sdk.ZeroInt())
		for _, seq := range k.AllSequencers(ctx) {
			total = total.Add(seq.TokensCoin())
		}
		// check module balance is equal
		moduleAcc := k.accountK.GetModuleAccount(ctx, types.ModuleName)
		balances := k.bankKeeper.GetAllBalances(ctx, moduleAcc.GetAddress())
		if 1 < len(balances) {
			return errors.New("module account has more than one coin"), true
		}
		if !total.IsZero() && len(balances) == 0 {
			return errors.New("module account has no balance"), true
		}
		if !total.IsZero() && !balances[0].IsEqual(total) {
			return errors.New("module account balance not equal to sum of sequencer tokens"), true
		}
		return nil, false
	}
}
