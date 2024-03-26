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
	err := k.fraudProofVerifier.Init(&fp)
	if err != nil {
		return err
	}
	err = k.fraudProofVerifier.Verify(&fp)
	if err != nil {
		return err
	}

	return nil
}

// ValidateFraudProof validates fraud proof preState Hash against the state update posted on the hub
func (k *Keeper) ValidateFraudProof(ctx sdk.Context, rollappID string, fp fraudtypes.FraudProof) error {
	// validate the fp struct and witnesses
	_, err := fp.ValidateBasic()
	if err != nil {
		return fmt.Errorf("validate basic: %w", err)
	}

	// validate the fraud proof against the committed state
	stateInfo, err := k.FindStateInfoByHeight(ctx, rollappID, uint64(fp.GetFraudulentBlockHeight()))
	if err != nil {
		return fmt.Errorf("find state info by height: %w", err)
	}
	blockDescriptor := stateInfo.BlockDescriptorByHeight(uint64(fp.GetFraudulentBlockHeight()))

	err = validateBlockDescriptor(blockDescriptor, fp.PreStateAppHash, fp.ExpectedValidAppHash)
	if err != nil {
		return fmt.Errorf("validate block descriptor: %w", err)
	}
	// TODO: Validate the fraudulent state transition is contained in the block header (TODO: ask Michael)
	// TODO: anything else to do?
	return nil
}

// expectedValidAppHash is the app hash that the fraud proof creator thinks the app hash should be after
// the invalid state transition. Therefore, we shall only process the proof if it's different from what we had before
func validateBlockDescriptor(bd types.BlockDescriptor, preAppStateHash, expectedValidAppHash []byte) error {
	if len(bd.IntermediateStatesRoots) == 0 {
		return types.ErrMissingIntermediateStatesRoots
	}

	for i, isr := range bd.IntermediateStatesRoots[:len(bd.IntermediateStatesRoots)-1] { // go over all but the last
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
