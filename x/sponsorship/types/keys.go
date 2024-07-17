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

func ParamsKey() []byte {
	return []byte{ParamsByte}
}

func DistributionKey() []byte {
	return []byte{DistributionByte}
}

func VoteKey(voterAddr string) []byte {
	return append([]byte{VoteByte}, []byte(voterAddr)...)
}
