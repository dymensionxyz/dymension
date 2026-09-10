package types

import "time"

const (
	ModuleName = "forward"
	// not to be confused with ibc apps PFM which uses 'forward' as the fungible packet json memo key
	HookNameRollToHL  = "dym-fwd-roll-hl"
	HookNameRollToIBC = "dym-fwd-roll-ibc"

	// DefaultForwardIBCTimeout is the fresh timeout applied to a forwarded IBC
	// hop when the composer's absolute timeout is missing or already stale at
	// execution time. Nanoseconds.
	DefaultForwardIBCTimeout = uint64(time.Hour)
	// MinForwardIBCTimeout is the minimum remaining lead time a composer-supplied
	// timeout must have (relative to block time) to be honored as-is; anything
	// tighter is treated as stale and replaced with DefaultForwardIBCTimeout.
	MinForwardIBCTimeout = uint64(10 * time.Minute)
)
