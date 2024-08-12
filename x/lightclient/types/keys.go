package types

const (
	// ModuleName defines the module name
	ModuleName = "lightclient"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	SequencersKeyPrefix = []byte{0x00} // prefix/seqAddr

)
