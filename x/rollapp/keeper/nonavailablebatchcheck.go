package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k *Keeper) VerifyNonAvailableBatch(ctx sdk.Context, msg *types.MsgNonAvailableBatch) error {

	/*var namespace []byte
		blob, _, err := k.blobsAndCommitments(namespace, msg.GetBlob())
		if err != nil {
			return err
		}
		switch msg.GetCase() {
		case types.NonAvaliableCase_wrongcommitment:
			result, err := k.verifyBlobInclusion(ctx, blob, msg.GetNmtproofs(), msg.GetNmtroots(), msg.GetRproofs(), msg.GetDataroot())
			if err != nil {
				return err
			}
			if result {
				return nil
			} else {
				return types.ErrWrongCommitment
			}
		case types.NonAvaliableCase_notavailable:

			return types.ErrBatchNotAvailable
		case types.NonAvaliableCase_invaliddata:

			return types.ErrInvalidBlobData
		default:
			return types.ErrSubmitNonAvailableBatchWrongCase
		}
	}

	func (k *Keeper) verifyBlobInclusion(ctx sdk.Context, blob *blob.Blob, nProofs [][]byte, rowRoots [][]byte, rProofs [][]byte, dataRoot []byte) (boolean, error) {
		var nmtProofs []*nmt.Proof
		for _, codedNMTProof := range nProofs {
			var unmarshalledProof nmt.Proof
			err := unmarshalledProof.UnmarshalJSON(codedNMTProof)
			if err != nil {
				return false, err
			}
			nmtProofs = append(nmtProofs, &unmarshalledProof)
		}

		//_, rowProofs := merkle.ProofsFromByteSlices(rProofs)*/

	/*for i, nmtProof := range nmtProofs {
		nmtProof.
	}*/
	return nil
}

/*func (k *Keeper) blobsAndCommitments(namespace []byte, daBlob []byte) (*blob.Blob, []byte, error) {
	b, err := blob.NewBlobV0(namespace, daBlob)
	if err != nil {
		return nil, nil, err
	}

	commitment, err := blob.CreateCommitment(b)
	if err != nil {
		return nil, nil, err
	}

	return b, commitment, nil
}*/
