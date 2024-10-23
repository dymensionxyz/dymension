package keeper

import (
	"sort"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k Keeper) ChooseProposer(ctx sdk.Context, rollappId string) error {
	proposer, err := k.GetProposer(ctx, rollappId)
	if err != nil {
		return errorsmod.Wrap(err, "get proposer")
	}
	if !proposer.Sentinel() {
		if !proposer.Bonded() {
			return gerrc.ErrInternal.Wrap("proposer is unbonded")
		}
	}
	successor := k.GetNextProposer()
}

// ExpectedNextProposer returns the next proposer for a rollapp
// it selects the next proposer from the bonded sequencers by bond amount
// if there are no bonded sequencers, it returns an empty sequencer
func (k Keeper) ExpectedNextProposer(ctx sdk.Context, rollappId string) types.Sequencer {
	// if nextProposer is set, were in the middle of rotation. The expected next proposer cannot change
	seq, ok := k.GetNextProposer(ctx, rollappId)
	if ok {
		return seq
	}

	// take the next bonded sequencer to be the proposer. sorted by bond
	seqs := k.GetSequencersByRollappByStatus(ctx, rollappId, types.Bonded)
	sort.SliceStable(seqs, func(i, j int) bool {
		return seqs[i].Tokens.IsAllGT(seqs[j].Tokens)
	})

	// return the first sequencer that is not the proposer
	proposer, _ := k.GetProposerLegacy(ctx, rollappId)
	for _, s := range seqs {
		if s.Address != proposer.Address {
			return s
		}
	}

	return types.Sequencer{}
}
