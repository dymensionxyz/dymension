package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/uinv"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

var invs = uinv.NamedFuncsList[Keeper]{
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

func InvariantNotice(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		seqs, err := k.NoticeQueue(ctx, nil)
		if err != nil {
			return err
		}
		var errs []error
		for _, seq := range seqs {
			if !seq.NoticeStarted() {
				errs = append(errs, fmt.Errorf("in notice queue but notice not started: %s", seq.Address))
			}
			if !k.IsProposer(ctx, seq) {
				errs = append(errs, fmt.Errorf("in notice queue but not proposer: %s", seq.Address))
			}
		}
		return errors.Join(errs...)
	})
}

func InvariantHashIndex(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		var errs []error
		for _, seq := range k.AllSequencers(ctx) {
			err := checkSeqHashIndex(ctx, k, seq)
			err = errorsmod.Wrapf(err, "sequencer: %s", seq.Address)
			errs = append(errs, err)
		}
		return errors.Join(errs...)
	})
}

func checkSeqHashIndex(ctx sdk.Context, k Keeper, exp types.Sequencer) error {
	got, err := k.SequencerByDymintAddr(ctx, exp.MustValsetHash())
	if err != nil {
		return err
	}
	if got.Address != exp.Address {
		return errors.New("address mismatch")
	}
	return nil
}

func InvariantStatus(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		var errs []error
		rollapps := k.rollappKeeper.GetAllRollapps(ctx)
		for _, ra := range rollapps {
			err := checkRollappStatus(ctx, k, ra.RollappId)
			err = errorsmod.Wrapf(err, "rollapp: %s", ra.RollappId)
			errs = append(errs, err)
		}
		for _, seq := range k.AllProposers(ctx) {
			if !k.IsProposer(ctx, seq) {
				errs = append(errs, fmt.Errorf("proposer in query is not proposer: %s", seq.Address))
			}
		}
		for _, seq := range k.AllSuccessors(ctx) {
			if !k.IsSuccessor(ctx, seq) {
				errs = append(errs, fmt.Errorf("successor in query is not successor: %s", seq.Address))
			}
		}
		return errors.Join(errs...)
	})
}

func checkRollappStatus(ctx sdk.Context, k Keeper, ra string) error {
	proposer := k.GetProposer(ctx, ra)
	if !proposer.Bonded() {
		return errors.New("proposer not bonded")
	}
	successor := k.GetSuccessor(ctx, ra)
	if !successor.Bonded() {
		return errors.New("successor not bonded")
	}
	if !proposer.Sentinel() && proposer.Address == successor.Address {
		return errors.New("proposer and successor are the same")
	}
	if !successor.Sentinel() && proposer.Sentinel() {
		return errors.New("proposer is sentinel but successor is not")
	}
	all := k.RollappSequencers(ctx, ra)
	bonded := k.RollappSequencersByStatus(ctx, ra, types.Bonded)
	unbonded := k.RollappSequencersByStatus(ctx, ra, types.Unbonded)
	if len(all) != len(bonded)+len(unbonded) {
		return errors.New("sequencer by rollapp length is not equal to sum of bonded, and unbonded")
	}
	return nil
}

func InvariantTokens(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		var errs []error

		for _, seq := range k.AllSequencers(ctx) {
			err := checkSeqTokens(ctx, seq, k)
			err = errorsmod.Wrapf(err, "sequencer: %s", seq.Address)
		}

		if err := errors.Join(errs...); err != nil {
			return err
		}

		total := sdk.NewCoin(k.bondDenom(ctx), sdk.ZeroInt())
		for _, seq := range k.AllSequencers(ctx) {
			total = total.Add(seq.TokensCoin())
		}
		// check module balance is equal
		moduleAcc := k.accountK.GetModuleAccount(ctx, types.ModuleName)
		balances := k.bankKeeper.GetAllBalances(ctx, moduleAcc.GetAddress())
		if 1 < len(balances) {
			return errors.New("module account has more than one coin")
		}
		if !total.IsZero() && len(balances) == 0 {
			return errors.New("module account has no balance")
		}
		if !total.IsZero() && !balances[0].IsEqual(total) {
			return errors.New("module account balance not equal to sum of sequencer tokens")
		}
		return nil
	})
}

func checkSeqTokens(ctx sdk.Context, seq types.Sequencer, k Keeper) error {
	if err := seq.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "validate basic")
	}
	if err := k.validBondDenom(ctx, seq.TokensCoin()); err != nil {
		return errorsmod.Wrap(err, "valid bond denom")
	}
	if seq.TokensCoin().Amount.IsNegative() {
		return errors.New("negative seq tokens")
	}
	return nil
}
