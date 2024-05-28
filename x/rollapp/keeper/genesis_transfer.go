package keeper

import (
	"encoding/json"

	sdkerrs "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type genesisTransferDenomMemo struct {
	// If the packet originates from the chain itself, and not a user, this will be true
	DoesNotOriginateFromUser bool           `json:"does_not_originate_from_user"`
	DenomMetadata            types.Metadata `json:"denom_metadata"`
}

func ParseGenesisTransferDenom(memo string) (*types.Metadata, error) {
	var t genesisTransferDenomMemo
	err := json.Unmarshal([]byte(memo), &t)
	if err != nil {
		return nil, sdkerrs.ErrJSONUnmarshal
	}
	if !t.DoesNotOriginateFromUser {
		return nil, sdkerrs.ErrUnauthorized
	}
	return &t.DenomMetadata, nil
}
