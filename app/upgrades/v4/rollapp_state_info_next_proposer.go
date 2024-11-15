package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	sequencerkeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
)

// migrateRollappStateInfoNextProposer should be called
//   - after migrateSequencerIndices: it uses the proposer index
//   - before migrateRollappFinalizationQueue: it uses the old queue to reduce store reads
func migrateRollappStateInfoNextProposer(ctx sdk.Context, rk *rollappkeeper.Keeper, sk *sequencerkeeper.Keeper) error {
	// use old queue since it requires less store reads
	q := rk.GetAllBlockHeightToFinalizationQueue(ctx)

	// map rollappID to proposer address to avoid multiple reads from the sequencer keeper
	// safe since the map is used only for existence checks
	rollappProposers := make(map[string]string)

	for _, queue := range q {
		for _, stateInfoIndex := range queue.FinalizationQueue {
			// get current proposer address
			proposer, ok := rollappProposers[stateInfoIndex.RollappId]
			if !ok {
				p := sk.GetProposer(ctx, stateInfoIndex.RollappId)
				if p.Sentinel() {
					return gerrc.ErrInternal.Wrapf("proposer cannot be sentinel: rollappID %s", stateInfoIndex.RollappId)
				}
				rollappProposers[stateInfoIndex.RollappId] = p.Address
				proposer = p.Address
			}

			// update state info
			stateInfo, found := rk.GetStateInfo(ctx, stateInfoIndex.RollappId, stateInfoIndex.Index)
			if !found {
				return gerrc.ErrInternal.Wrapf("state info not found: rollappID %s, index %d", stateInfoIndex.RollappId, stateInfoIndex.Index)
			}
			stateInfo.NextProposer = proposer
			rk.SetStateInfo(ctx, stateInfo)
		}
	}

	return nil
}
