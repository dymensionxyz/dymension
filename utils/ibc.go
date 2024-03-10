package utils

import (
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
)

const (
	ibcPort = "transfer"
)

func GetForeginIBCDenom(channelId string, denom string) string {
	sourcePrefix := transfertypes.GetDenomPrefix(ibcPort, channelId)
	// NOTE: sourcePrefix contains the trailing "/"
	prefixedDenom := sourcePrefix + denom
	// construct the denomination trace from the full raw denomination
	denomTrace := transfertypes.ParseDenomTrace(prefixedDenom)
	ibcDenom := denomTrace.IBCDenom()

	return ibcDenom
}
