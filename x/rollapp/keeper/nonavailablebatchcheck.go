package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	inclusion "github.com/dymensionxyz/dymension/v3/app/dainclusionproofs"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k *Keeper) VerifyNonAvailableBatch(ctx sdk.Context, msg *types.MsgNonAvailableBatch, nonInclusionProof *inclusion.NonInclusionProof) error {

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

	err = nonInclusionProof.VerifyNonInclusion(DaMetaDataSubmitted.GetNameSpace(), DaMetaDataSubmitted.GetIndex(), DAMetaDataSequencer.GetLength(), DAMetaDataSequencer.GetDataRoot())
	//var namespace []byte
	if err != nil {
		return err
	}

	return nil
}
