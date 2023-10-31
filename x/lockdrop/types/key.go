package types

import (
	"fmt"
	"time"
)

const (
	ModuleName = "lockdrop"

	StoreKey = ModuleName

	RouterKey = ModuleName

	QuerierRoute = ModuleName
)

var (
	DistrInfoKey = []byte("distr_info")
)

// GetPoolGaugeIdStoreKey returns a StoreKey with pool ID and its duration as inputs
func GetPoolGaugeIdStoreKey(poolId uint64, duration time.Duration) []byte {
	return []byte(fmt.Sprintf("lockdrop/%d/%s", poolId, duration.String()))
}
