package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// Module name and store keys.
const (
	// ModuleName defines the module name
	ModuleName = "sponsorship"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName
)

const (
	ParamsByte uint8 = iota
	DistributionByte
	VotingPowerByte
	VoteByte
)

func ParamsKey() []byte {
	return []byte{ParamsByte}
}

func DistributionKey() []byte {
	return []byte{DistributionByte}
}

func VotingPowerKey(valAddr sdk.ValAddress, voterAddr sdk.AccAddress) []byte {
	key := make([]byte, 0, 1+len(voterAddr)+len(valAddr))
	key = append(key, VotingPowerByte)
	key = append(key, valAddr.Bytes()...)
	key = append(key, voterAddr.Bytes()...)
	return key
}

func VoteKey(voterAddr sdk.AccAddress) []byte {
	return append([]byte{VoteByte}, voterAddr.Bytes()...)
}
