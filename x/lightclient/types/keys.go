package types

const (
	// ModuleName defines the module name
	ModuleName = "lightclient"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName

	// TransientKey defines the module's transient store key
	TransientKey = "t_lightclient"
)

// KV Store
var (
	RollappClientKey = []byte{0x01}
)

// Transient Store
var (
	LightClientRegistrationKey = []byte{0x02}
)

func CanonicalClientKey(rollappId string) []byte {
	return append(RollappClientKey, []byte(rollappId)...)
}

func CanonicalLightClientRegistrationKey(rollappId string) []byte {
	return append(LightClientRegistrationKey, []byte(rollappId)...)
}
