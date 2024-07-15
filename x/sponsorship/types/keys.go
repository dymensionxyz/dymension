package types

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
	VoteByte
)

func ParamsPrefix() []byte {
	return []byte{ParamsByte}
}

func DistributionPrefix() []byte {
	return []byte{DistributionByte}
}

func VotePrefix() []byte {
	return []byte{VoteByte}
}
