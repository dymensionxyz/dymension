package types

import (
	"time"
)

// LivenessCheckParams are the params needed to do a liveness check
// This is a utility struct to make dealing with these params which are commonly used together
// easier.
// TODO: do I really need this?
type LivenessCheckParams struct {
	// HubExpectedBlockTime	is how long it typically takes to produce a Hub block (e.g. 6 secs). Used to calculate length of time based on height.
	HubExpectedBlockTime time.Duration
	// SlashTime is the time a sequencer has to post a block, before he will be slashed
	SlashTime time.Duration
	// SlashInterval is the min gap between a sequence of slashes if the sequencer continues to be down
	SlashInterval time.Duration
	// JailTime	is the time a sequencer can be down after which he will be jailed rather than slashed
	JailTime time.Duration
}
