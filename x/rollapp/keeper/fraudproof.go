package keeper

import (
	"fmt"

	fraudtypes "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k *Keeper) VerifyFraudProof(ctx sdk.Context, rollappID string, fp fraudtypes.FraudProof) error {
	err := k.fraudProofVerifier.InitFromFraudProof(&fp)
	if err != nil {
		return err
	}

	err = k.ValidateFraudProof(ctx, rollappID, fp)
	if err != nil {
		return err
	}

	err = k.fraudProofVerifier.VerifyFraudProof(&fp)
	if err != nil {
		return err
	}

	return nil
}

// validate fraud proof preState Hash against the state update posted on the hub
func (k *Keeper) ValidateFraudProof(ctx sdk.Context, rollappID string, fp fraudtypes.FraudProof) error {
	stateInfo, err := k.FindStateInfoByHeight(ctx, rollappID, uint64(fp.BlockHeight))
	if err != nil {
		return err
	}
	idx := fp.BlockHeight - int64(stateInfo.StartHeight)
	blockDescriptor := stateInfo.BDs.BD[idx]
	if blockDescriptor.Height != uint64(fp.BlockHeight) {
		return fmt.Errorf("invalid block height")
	}
	//FIXME: validate fraud proof initial state against the current state posted on the hub
	// if blockDescriptor.IntermediateStatesRoot[0] != fp.PreStateAppHash {
	// 	return ErrInvalidPreStateAppHash
	// }

	return nil
}
