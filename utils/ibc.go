package utils

import (
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
)

const (
	ibcPort = "transfer"
)

func GetForeignIBCDenom(channelId string, denom string) string {
	return GetForeignDenomTrace(channelId, denom).IBCDenom()
}

func GetForeignDenomTrace(channelId string, denom string) transfertypes.DenomTrace {
	sourcePrefix := transfertypes.GetDenomPrefix(ibcPort, channelId)
	// NOTE: sourcePrefix contains the trailing "/"
	prefixedDenom := sourcePrefix + denom
	// construct the denomination trace from the full raw denomination
	return transfertypes.ParseDenomTrace(prefixedDenom)
}
