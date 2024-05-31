package keeper

import (
	"encoding/json"

	sdkerrs "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type genesisTransferDenomMemo struct {
	GenesisTransfer struct {
		Data types.Metadata `json:"data"`
	} `json:"genesis_transfer"`
}

func ParseGenesisTransferDenom(memo string) (types.Metadata, error) {
	var t genesisTransferDenomMemo
	err := json.Unmarshal([]byte(memo), &t)
	if err != nil || !t.IsGenesisDenomMetadata {
		return types.Metadata{}, sdkerrs.ErrNotSupported
	}
	if !t.DoesNotOriginateFromUser {
		return types.Metadata{}, sdkerrs.ErrUnauthorized
	}
	return t.DenomMetadata, nil
}
