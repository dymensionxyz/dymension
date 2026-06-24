package types

const (
	ModuleName = "agent"

	StoreKey = ModuleName

	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

// Store prefixes.
var (
	KeyParams    = []byte{0x00}
	KeyAgents    = []byte{0x01}
	KeyActionLog = []byte{0x02}
)
