package keeper

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

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
	targetDuration := slashTimeNoUpdate + ((max(0, downTime-slashTimeNoUpdate)+slashInterval-1)/slashInterval)*slashInterval
	heightEvent = heightLastRollappUpdate + int64((targetDuration+hubBlockInterval-1)/hubBlockInterval)
	isJail = jailTime <= targetDuration
	return
}

// CheckLiveness will slash or jail any sequencers for whom their rollapp has been down
// and a slash or jail event is due
func (k Keeper) CheckLiveness(ctx sdk.Context) {
}

func (k Keeper) GetScheduledLivenessEvents(ctx sdk.Context) {
}
