package keeper

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	inclusion "github.com/dymensionxyz/dymension/v3/app/dainclusionproofs"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k *Keeper) VerifyWrongCommitmentBatch(ctx sdk.Context, msg *types.MsgWrongCommitmentBatch, inclusionProof *inclusion.InclusionProof) error {

	stateInfo, found := k.GetStateInfo(ctx, msg.GetRollappId(), msg.GetSlIndex())

	if !found {
		return sdkerrors.ErrInvalidRequest
	}

	DAPath := stateInfo.GetDAPath()
	DAMetaDataSequencer, err := types.NewDAMetaData(DAPath)
	if err != nil {
		return err
	}

	DaMetaDataSubmitted, err := types.NewDAMetaData(msg.GetDAPath())
	if err != nil {
		return err
	}

	if !bytes.Equal(DAMetaDataSequencer.GetNameSpace(), DaMetaDataSubmitted.GetNameSpace()) {
		return sdkerrors.ErrInvalidRequest
	}

	commitment, index, length, err := inclusionProof.VerifyBlobInclusion(DAMetaDataSequencer.GetNameSpace(), DAMetaDataSequencer.GetDataRoot())
	//var namespace []byte
	if err != nil {
		return err
	}
	if !bytes.Equal(commitment, DAMetaDataSequencer.GetCommitment()) && index == DAMetaDataSequencer.GetIndex() && length == DAMetaDataSequencer.GetLength() {
		return nil
	}

	return sdkerrors.ErrInvalidRequest
}
