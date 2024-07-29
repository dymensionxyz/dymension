package keeper

import "time"

type CalculateNextSlashOrJailHeight func(
	HubBlockInterval time.Duration, // average time between hub blocks
	SlashTimeNoUpdate time.Duration, // time until first slash if not updating
	SlashInterval time.Duration, // gap between slash if still not updating
	JailTime time.Duration, // time until jail if not updating
	HubHeight int64, // current hub height
	LastRollappUpdateHeight int64, // when was the rollapp last updated
) (
	hubHeight int64, // height to schedule event
	isJail bool, // is it a jail event? (false -> slash)
)
