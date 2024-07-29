package keeper

import (
	"fmt"
	"testing"
	"time"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func TestNextSlashOrJailHeightExample(t *testing.T) {
	for hHub := range int64(30000) {
		h, jail := NextSlashOrJailHeight(
			types.DefaultHubExpectedBlockTime,
			types.DefaultLivenessSlashTime,
			types.DefaultLivenessSlashInterval,
			types.DefaultLivenessJailTime,
			hHub,
			0,
		)
		if h == hHub {
			elapsed := time.Duration(hHub) * types.DefaultHubExpectedBlockTime
			t.Log(fmt.Sprintf("hub height: %d, elapsed %s, jail: %t", hHub, elapsed, jail))
		}
		if jail {
			break
		}
	}
}
