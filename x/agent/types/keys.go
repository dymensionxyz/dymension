package types

const (
	ModuleName = "agent"

	StoreKey = ModuleName

	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

// Store prefixes.
var (
	KeyParams     = []byte{0x00}
	KeyAgents     = []byte{0x01}
	KeyActionLog  = []byte{0x02}
	KeyFeedback   = []byte{0x03}
	KeyReputation = []byte{0x04}
)
