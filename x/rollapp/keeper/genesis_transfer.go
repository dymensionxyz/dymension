package keeper

import (
	"encoding/json"

	sdkerrs "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type genesisTransferDenomMemo struct {
	// If the packet originates from the chain itself, and not a user, this will be true
	DoesNotOriginateFromUser bool           `json:"does_not_originate_from_user"`
	IsGenesisDenomMetadata   bool           `json:"is_genesis_denom_metadata"`
	DenomMetadata            types.Metadata `json:"denom_metadata"`
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
