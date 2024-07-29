package keeper

import (
	"fmt"
	"testing"
	"time"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func TestNextSlashOrJailHeightExample(t *testing.T) {
	hHub := int64(0)
	jail := false
	for !jail {
		last := hHub
		hHub, jail = NextSlashOrJailHeight(
			types.DefaultHubExpectedBlockTime,
			types.DefaultLivenessSlashTime,
			types.DefaultLivenessSlashInterval,
			types.DefaultLivenessJailTime,
			hHub,
			0,
		)
		elapsed := time.Duration(hHub) * types.DefaultHubExpectedBlockTime
		t.Log(fmt.Sprintf("hub height: %d, elapsed %s, jail: %t", hHub, elapsed, jail))
		if last == hHub {
			hHub++
		}
	}
}
