package types

// HistoricalSellOrderMinExpiry defines the structure for the historical sell order minimum expiry of each Dym-Name.
type HistoricalSellOrderMinExpiry struct {
	DymName   string
	MinExpiry int64
}
