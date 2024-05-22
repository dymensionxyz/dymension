package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// BankKeeper defines the expected interface needed
type BankKeeper interface {
	GetDenomMetaData(ctx sdk.Context, denom string) (types.Metadata, bool)
}

type DelayedAckKeeper interface {
	ExtractRollappAndTransferPacketFromData(
		ctx sdk.Context,
		data []byte,
		rollappPortOnHub string,
		rollappChannelOnHub string,
	) (*rollapptypes.Rollapp, *transfertypes.FungibleTokenPacketData, error)
	SendPacket(
		ctx sdk.Context,
		chanCap *capabilitytypes.Capability,
		sourcePort string,
		sourceChannel string,
		timeoutHeight clienttypes.Height,
		timeoutTimestamp uint64,
		data []byte,
	) (sequence uint64, err error)
}
