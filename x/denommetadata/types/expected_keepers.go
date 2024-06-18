package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

// TransferKeeper defines the expected transfer keeper
type TransferKeeper interface {
	HasDenomTrace(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) bool
	SetDenomTrace(ctx sdk.Context, denomTrace transfertypes.DenomTrace)
}

// BankKeeper defines the expected interface needed
type BankKeeper interface {
	HasDenomMetaData(ctx sdk.Context, denom string) bool
	SetDenomMetaData(ctx sdk.Context, denomMetaData types.Metadata)
}

type DenomMetadataKeeper interface {
	CreateDenomMetadata(ctx sdk.Context, metadata types.Metadata) error
}
