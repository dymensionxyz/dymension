package keeper

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	inclusion "github.com/dymensionxyz/dymension/v3/app/dainclusionproofs"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k *Keeper) VerifyNonAvailableBatch(ctx sdk.Context, msg *types.MsgNonAvailableBatch, nonInclusionProof *inclusion.NonInclusionProof) error {

	stateInfo, found := k.GetStateInfo(ctx, msg.GetRollappId(), msg.GetSlIndex())

	if !found {
		return sdkerrors.ErrInvalidRequest
	}

	DAPath := stateInfo.GetDAPath()
	DAMetaDataSequencer, err := types.NewDAMetaData(DAPath)
	if err != nil {
		return err
	}

	err = nonInclusionProof.VerifyNonInclusion(DAMetaDataSequencer.GetIndex(), DAMetaDataSequencer.GetLength(), DAMetaDataSequencer.GetDataRoot())
	//var namespace []byte
	if err != nil {
		return err
	}

	return nil
}
