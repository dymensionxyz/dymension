package keeper

import "time"

func NextSlashOrJailHeight(
	hubBlockInterval time.Duration, // average time between hub blocks
	slashTimeNoUpdate time.Duration, // time until first slash if not updating
	slashInterval time.Duration, // gap between slash if still not updating
	jailTime time.Duration, // time until jail if not updating
	heightHub int64, // current hub height
	heightLastRollappUpdate int64, // when was the rollapp last updated
) (
	heightEvent int64, // hub height to schedule event
	isJail bool, // is it a jail event? (false -> slash)
) {
	// how long has the rollapp been down already?
	downTime := time.Duration(heightHub-heightLastRollappUpdate) * hubBlockInterval
	// when should we schedule the next slash/jail, in terms of down time duration?
	var targetDuration time.Duration
	if downTime < slashTimeNoUpdate {
		targetDuration = slashTimeNoUpdate
	} else {
		targetDuration = slashTimeNoUpdate + ((downTime-slashTimeNoUpdate+slashInterval-1)/slashInterval)*slashInterval
		// targetDuration = slashTimeNoUpdate + ((downTime-slashTimeNoUpdate)/slashInterval+1)*slashInterval
	}
	heightEvent = heightLastRollappUpdate + int64((targetDuration+hubBlockInterval-1)/hubBlockInterval) // doesn't work
	// heightEvent = heightLastRollappUpdate + int64(targetDuration/hubBlockInterval) + 1                  // works but wrong
	isJail = jailTime <= targetDuration
	return
}