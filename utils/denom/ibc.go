package denom

import (
	"strings"

	transferTypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
)

// ValidateIBCDenom validates that the given denomination is a valid fungible token representation (i.e 'ibc/{hash}')
// per ADR 001 https://github.com/cosmos/ibc-go/blob/main/docs/architecture/adr-001-coin-source-tracing.md.
// If the denom is valid, return its hash-string part. Inspired by
// https://github.com/cosmos/ibc-go/blob/5d7655684554e4f577be9573ef94ef4ad6c82667/modules/apps/transfer/types/denom.go#L190.
func ValidateIBCDenom(denom string) (string, bool) {
	denomSplit := strings.SplitN(denom, "/", 2)

	if len(denomSplit) == 2 && denomSplit[0] == transferTypes.DenomPrefix && strings.TrimSpace(denomSplit[1]) != "" {
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
