package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	inclusion "github.com/dymensionxyz/dymension/v3/app/dainclusionproofs"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k *Keeper) VerifyWrongCommitmentBatch(ctx sdk.Context, msg *types.MsgWrongCommitmentBatch, inclusionProof *inclusion.InclusionProof) error {

	stateInfo, found := k.GetStateInfo(ctx, msg.GetRollappId(), msg.GetSlIndex())
	if !found {
		return nil
	}
	DAPath := stateInfo.GetDAPath()
	DAMetaDataSequencer, err := types.NewDAMetaData(DAPath)
	if err != nil {
		return nil
	}
	DaMetaDataSubmitted, err := types.NewDAMetaData(msg.GetDAPath())
	if err != nil {
		return nil
	}

	err = inclusionProof.VerifyBlobInclusion(DaMetaDataSubmitted.GetCommitment(), DaMetaDataSubmitted.GetNameSpace(), DAMetaDataSequencer.GetDataRoot())
	//var namespace []byte
	if err != nil {
		return err
	}

	return nil
}
