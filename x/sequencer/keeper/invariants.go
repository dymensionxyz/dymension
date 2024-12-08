package keeper

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/uinv"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

var invs = uinv.NamedFuncsList[Keeper]{
	{Name: "notice", Func: InvariantNotice},
	{Name: "hash-index", Func: InvariantProposerAddrIndex},
	{Name: "status", Func: InvariantStatus},
	{Name: "tokens", Func: InvariantTokens},
	{Name: "do-not-expose-sentinel", Func: InvariantDoNotExposeSentinel},
}

// RegisterInvariants registers the sequencer module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	invs.RegisterInvariants(types.ModuleName, ir, k)
}

// DO NOT DELETE
func AllInvariants(k Keeper) sdk.Invariant {
	return invs.All(types.ModuleName, k)
}

// notice queue should have only proposers who started notice
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

// the lookup proposer hash -> seq should be populated
func InvariantProposerAddrIndex(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		var errs []error
		for _, seq := range k.AllSequencers(ctx) {
			err := checkProposerAddrIndex(ctx, k, seq)
			err = errorsmod.Wrapf(err, "sequencer: %s", seq.Address)
			errs = append(errs, err)
		}
		return errors.Join(errs...)
	})
}

func checkProposerAddrIndex(ctx sdk.Context, k Keeper, exp types.Sequencer) error {
	hash := exp.MustProposerAddr()
	got, err := k.SequencerByDymintAddr(ctx, hash)
	if err != nil {
		return errorsmod.Wrapf(err, "seq by dymint addr: proposer hash: %x", hash)
	}
	if got.Address != exp.Address {
		return fmt.Errorf("hash index mismatch: got addr: %s, exp addr: %s", got.Address, exp.Address)
	}
	return nil
}

// proposer and successor status' should be sensible
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

// module balance must correspond to sequencer stakes, and sequencer stakes should be sensible
func InvariantTokens(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		var errs []error

		for _, seq := range k.AllSequencers(ctx) {
			err := checkSeqTokens(seq)
			err = errorsmod.Wrapf(err, "sequencer: %s", seq.Address)
			errs = append(errs, err)
		}

		if err := errors.Join(errs...); err != nil {
			return err
		}

		total := sdk.NewCoin(commontypes.DYMCoin.Denom, sdk.ZeroInt())
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

func checkSeqTokens(seq types.Sequencer) error {
	if err := seq.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "validate basic")
	}
	if err := validBondDenom(seq.TokensCoin()); err != nil {
		return errorsmod.Wrap(err, "valid bond denom")
	}
	if seq.TokensCoin().Amount.IsNegative() {
		return errors.New("negative seq tokens")
	}
	return nil
}

// sentinel should not be available in the global index or rollapp wise index
// (it's only available in getProposer or getSuccessor)
func InvariantDoNotExposeSentinel(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		var errs []error
		for _, s := range k.AllSequencers(ctx) {
			if s.Sentinel() {
				errs = append(errs, fmt.Errorf("sentinel in global index: %s", s.Address))
			}
		}
		rollapps := k.rollappKeeper.GetAllRollapps(ctx)
		for _, ra := range rollapps {
			for _, s := range k.RollappSequencers(ctx, ra.RollappId) {
				if s.Sentinel() {
					errs = append(errs, fmt.Errorf("sentinel in rollapp index: %s", ra.RollappId))
				}
			}
		}
		return errors.Join(errs...)
	})
}
