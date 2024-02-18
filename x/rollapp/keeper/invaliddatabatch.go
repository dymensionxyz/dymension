package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k *Keeper) VerifyInvalidDataBatch(ctx sdk.Context, msg *types.MsgInvalidDataBatch) error {

	/*stateInfo, found := k.GetStateInfo(ctx, msg.GetRollappId(), msg.GetSlIndex())
	if !found {
		return nil
	}
	DAPath := stateInfo.GetDAPath()
	DAMetaData, err := types.NewDAMetaData(DAPath)
	if err != nil {
		return nil
	}
	//var namespace []byte

	err = k.verifyBlobNonInclusion(ctx, DAMetaData.GetNameSpace(), msg.GetRproofs(), msg.GetDataroot())
	if err != nil {
		return err
	}*/

	return nil
}

/*func (k *Keeper) validateIncludedData(ctx sdk.Context, b *blob.Blob) error {
	//TODO (srene): Implement check invalid data blob
	return types.ErrUnableToVerifyProof
}*/
