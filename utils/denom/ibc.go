package denom

import (
	"strings"

	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/dymensionxyz/sdk-utils/utils/uibc"
)

// ValidateIBCDenom validates that the given denomination is a valid fungible token representation (i.e 'ibc/{hash}')
// per ADR 001 https://github.com/cosmos/ibc-go/blob/main/docs/architecture/adr-001-coin-source-tracing.md.
// If the denom is valid, return its hash-string part. Inspired by
// https://github.com/cosmos/ibc-go/blob/5d7655684554e4f577be9573ef94ef4ad6c82667/modules/apps/transfer/types/denom.go#L190.
func ValidateIBCDenom(denom string) (string, bool) {
	denomSplit := strings.SplitN(denom, "/", 2)

	if len(denomSplit) == 2 && denomSplit[0] == transfertypes.DenomPrefix && strings.TrimSpace(denomSplit[1]) != "" {
		return denomSplit[1], true
	}

	return "", false
}

// SourcePortChanFromTracePath extracts source port and channel from the provided IBC denom trace path.
// References:
//   - https://github.com/cosmos/ibc-go/blob/main/docs/architecture/adr-001-coin-source-tracing.md
//   - https://github.com/cosmos/relayer/issues/288
func SourcePortChanFromTracePath(tracePath string) (sourcePort, sourceChannel string, validTrace bool) {
	sp := strings.Split(tracePath, "/")
	if len(sp) < 2 {
		return "", "", false
	}
	sourcePort, sourceChannel = sp[len(sp)-2], sp[len(sp)-1]
	if sourcePort == "" || sourceChannel == "" {
		return "", "", false
	}
	return sourcePort, sourceChannel, true
}

// GetIncomingTransferDenom returns the denom that will be credited on the hub.
// The denom logic follows the transfer middleware's logic and is necessary in order to prefix/non-prefix the denom
// based on the original chain it was sent from.
func GetIncomingTransferDenom(packet channeltypes.Packet, fungibleTokenPacketData transfertypes.FungibleTokenPacketData) string {
	var denom string

	if transfertypes.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), fungibleTokenPacketData.Denom) {
		// remove prefix added by sender chain
		voucherPrefix := transfertypes.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
		unprefixedDenom := fungibleTokenPacketData.Denom[len(voucherPrefix):]
		// coin denomination used in sending from the escrow address
		denom = unprefixedDenom
		// The denomination used to send the coins is either the native denom or the hash of the path
		// if the denomination is not native.
		denomTrace := transfertypes.ParseDenomTrace(unprefixedDenom)
		if denomTrace.Path != "" {
			denom = denomTrace.IBCDenom()
		}
	} else {
		denom = uibc.GetForeignDenomTrace(packet.GetDestChannel(), fungibleTokenPacketData.Denom).IBCDenom()
	}
	return denom
}
