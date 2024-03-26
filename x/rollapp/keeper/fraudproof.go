package keeper

import (
	"bytes"

	fraudtypes "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k *Keeper) VerifyFraudProof(ctx sdk.Context, rollappID string, fp fraudtypes.FraudProof) error {
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
	err := k.fraudProofVerifier.InitFromFraudProof(&fp)
	if err != nil {
		return err
	}
	err = k.fraudProofVerifier.VerifyFraudProof(&fp)
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
		return err
	}

	// validate the fraudproof against the commited state
	blockHeight := fp.BlockHeight + 1
	stateInfo, err := k.FindStateInfoByHeight(ctx, rollappID, uint64(blockHeight))
	if err != nil {
		return err
	}
	idx := blockHeight - int64(stateInfo.StartHeight)
	blockDescriptor := stateInfo.BDs.BD[idx]

	if blockDescriptor.IntermediateStatesRoots == nil {
		return types.ErrMissingIntermediateStatesRoots
	}

	foundIdx := -1
	for idx, isr := range blockDescriptor.IntermediateStatesRoots {
		// skip the last ISR
		if idx == len(blockDescriptor.IntermediateStatesRoots)-1 {
			break
		}
		if bytes.Equal(isr, fp.PreStateAppHash) {
			foundIdx = idx
			break
		}
	}

	if foundIdx == -1 {
		return types.ErrInvalidPreStateAppHash
	}

	if bytes.Equal(blockDescriptor.IntermediateStatesRoots[foundIdx+1], fp.ExpectedValidAppHash) {
		return types.ErrInvalidExpectedAppHash
	}

	// TODO: Validate the fraudulent state transition is contained in the block header

	return nil
}
