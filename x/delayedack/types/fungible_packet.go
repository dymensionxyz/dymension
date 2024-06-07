package types

import (
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
)

type TransferData struct {
	transfertypes.FungibleTokenPacketData
	RollappID string
}
