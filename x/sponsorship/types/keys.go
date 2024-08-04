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
	ParamsByte                  uint8 = iota // Module params: Params
	DistributionByte                         // Current distribution: Distribution
	DelegatorValidatorPowerByte              // Delegator voting power by the validator: math.Int
	VoteByte                                 // User's vote: Vote
	InactiveVoterByte                        // Votes having less than the min voting power to be removed at the end of the block
)

func ParamsKey() []byte {
	return []byte{ParamsByte}
}

func DistributionKey() []byte {
	return []byte{DistributionByte}
}

func DelegatorValidatorPowerKey(voterAddr sdk.AccAddress, valAddr sdk.ValAddress) []byte {
	key := make([]byte, 0, 1+len(voterAddr)+len(valAddr))
	key = append(key, DelegatorValidatorPowerByte)
	key = append(key, voterAddr.Bytes()...)
	key = append(key, valAddr.Bytes()...)
	return key
}

func AllDelegatorValidatorPowersKey(voterAddr sdk.AccAddress) []byte {
	key := make([]byte, 0, 1+len(voterAddr))
	key = append(key, DelegatorValidatorPowerByte)
	key = append(key, voterAddr.Bytes()...)
	return key
}

func VoteKey(voterAddr sdk.AccAddress) []byte {
	return append([]byte{VoteByte}, voterAddr.Bytes()...)
}

func InactiveVoterKey() []byte {
	return []byte{InactiveVoterByte}
}
