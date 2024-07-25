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
	VotedDelegationsByte
	VoteByte
)

func ParamsKey() []byte {
	return []byte{ParamsByte}
}

func DistributionKey() []byte {
	return []byte{DistributionByte}
}

func VotedDelegationKey(valAddr sdk.ValAddress, voterAddr sdk.AccAddress) []byte {
	key := make([]byte, 0, 1+len(voterAddr)+len(valAddr))
	key = append(key, VotedDelegationsByte)
	key = append(key, valAddr.Bytes()...)
	key = append(key, voterAddr.Bytes()...)
	return key
}

func VotedDelegationsByValidatorKey(valAddr sdk.ValAddress) []byte {
	key := make([]byte, 0, 1+len(valAddr))
	key = append(key, VotedDelegationsByte)
	key = append(key, valAddr.Bytes()...)
	return key
}

func VoteKey(voterAddr sdk.AccAddress) []byte {
	return append([]byte{VoteByte}, voterAddr.Bytes()...)
}
