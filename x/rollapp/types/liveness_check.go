package types

import (
	"time"

	"cosmossdk.io/math"
)

// LivenessCheckParams are the params needed to do a liveness check
type LivenessCheckParams struct {
	// HubExpectedBlockTime	is how long it typically takes to produce a Hub block (e.g. 6 secs). Used to calculate length of time based on height.
	HubExpectedBlockTime time.Duration
	// LivenessSlashTime is the time a sequencer has to post a block, before he will be slashed
	LivenessSlashTime time.Duration
	// LivenessSlashInterval is the min gap between a sequence of slashes if the sequencer continues to be down
	LivenessSlashInterval time.Duration
	// LivenessJailTime	is the time a sequencer can be down after which he will be jailed rather than slashed
	LivenessJailTime time.Duration
	// LivenessSlashMultiplier is a multiplier with the sequencer balance to calculate the slash amountr
	LivenessSlashMultiplier math.LegacyDec
	// LivenessSlashRewardMultiplier is a multiplier for the slashed amount to be sent to the successful slash TX reward addr
	LivenessSlashRewardMultiplier math.LegacyDec
}
