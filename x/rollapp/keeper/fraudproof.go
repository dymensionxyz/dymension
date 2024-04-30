package keeper

import (
	"bytes"
	"fmt"

	fraudtypes "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k *Keeper) ValidateAndRunFraudProof(ctx sdk.Context, rollappID string, fp fraudtypes.FraudProof) error {
	err := k.ValidateFraudProof(ctx, rollappID, fp)
	if err != nil {
		return err
	}

	err = k.RunFraudProof(fp)
	if err != nil {
		return err
	}

	return nil
}

func (k *Keeper) RunFraudProof(fp fraudtypes.FraudProof) error {
	return k.fraudProofVerifier(fp)
}

// ValidateFraudProof validates that the proof itself is well-formed and that it is valid against the current state of the chain
func (k *Keeper) ValidateFraudProof(ctx sdk.Context, rollappID string, fp fraudtypes.FraudProof) error {
	if err := fp.ValidateBasic(); err != nil {
		return fmt.Errorf("validate basic: %w", err)
	}

	// TODO(danwt): double check everything is done here
	// TODO(danwt): Validate the fraudulent state transition is contained in the block header ( ask Michael)

	stateInfo, err := k.FindStateInfoByHeight(ctx, rollappID, uint64(fp.GetFraudulentBlockHeight()))
	if err != nil {
		return fmt.Errorf("find state info by height: %w", err)
	}
	blockDescriptor := stateInfo.BlockDescriptorByHeight(uint64(fp.GetFraudulentBlockHeight()))
	err = validateBlockDescriptor(blockDescriptor, fp.PreStateAppHash, fp.ExpectedValidAppHash)
	if err != nil {
		return fmt.Errorf("validate against block descriptor: %w", err)
	}
	return nil
}

// expectedValidAppHash is the app hash that the fraud proof creator thinks the app hash should be after
// the invalid state transition. Therefore, we shall only process the proof if it's different from what we had before
func validateBlockDescriptor(bd types.BlockDescriptor, preAppStateHash, expectedValidAppHash []byte) error {
	if len(bd.IntermediateStatesRoots) == 0 {
		return types.ErrMissingIntermediateStatesRoots
	}

	for i, isr := range bd.IntermediateStatesRoots[:len(bd.IntermediateStatesRoots)-1] { // the last one has nothing to follow, so can't be a start point
		if bytes.Equal(isr, preAppStateHash) { // TODO(danwt): what if you have two hashes the same (would require empty block/ no tx?)
			if bytes.Equal(bd.IntermediateStatesRoots[i+1], expectedValidAppHash) {
				// in a valid fraud proof, these would be different, but they're not
				return types.ErrInvalidExpectedAppHash
			}
			return nil
		}
	}

	return types.ErrInvalidPreStateAppHash
}
