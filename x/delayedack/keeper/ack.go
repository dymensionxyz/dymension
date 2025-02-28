package keeper

import (
	"bytes"

	errorsmod "cosmossdk.io/errors"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibcerrors "github.com/cosmos/ibc-go/v7/modules/core/errors"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

func (k Keeper) ReadAck(bz []byte) (*channeltypes.Acknowledgement, error) {

	// TODO: maybe use 	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	//  types.ModuleCdc.UnmarshalJSO ... maybe??
	//  official suggestion is to use ibc 10 codec

	var ack channeltypes.Acknowledgement
	if err := k.Cdc().UnmarshalJSON(bz, &ack); err != nil {
		return nil, errorsmod.Wrapf(types.ErrUnknownRequest, "unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}
	bz2 := k.Cdc().MustMarshalJSON(&ack)
	if !bytes.Equal(bz, bz2) {
		return nil, errorsmod.Wrapf(ibcerrors.ErrInvalidType, "acknowledgement did not marshal to expected bytes: got %X ≠  expect %X", bz2, bz)
	}
	return &ack, nil
}
