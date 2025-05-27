package cli

import (
	"time"

	flag "github.com/spf13/pflag"
)

// Flags for incentives module tx commands.
const (
	FlagDuration  = "duration"
	FlagStartTime = "start-time"
	FlagEpochs    = "epochs"
	FlagPerpetual = "perpetual"
	FlagTimestamp = "timestamp"
	FlagOwner     = "owner"
	FlagLockIds   = "lock-ids"
	FlagEndEpoch  = "end-epoch"
	FlagLockAge   = "lock-age"
)

// FlagSetCreateGauge returns flags for creating gauges.
func FlagSetCreateGauge() *flag.FlagSet {
	fs := flag.NewFlagSet("", flag.ContinueOnError)

	dur, _ := time.ParseDuration("24h")
	fs.Duration(FlagDuration, dur, "The duration token to be locked, default 1d(24h). Other examples are 7d(168h), 14d(336h). Maximum unit is hour.")
	fs.Duration(FlagLockAge, 0, "The minimum age of the lock to qualify. Examples: 1h, 24h, 7d.")
	fs.String(FlagStartTime, "", "Timestamp to begin distribution")
	fs.Uint64(FlagEpochs, 0, "Total epochs to distribute tokens")
	fs.Bool(FlagPerpetual, false, "Perpetual distribution")
	return fs
}
