package keeper_test

import (
	"fmt"
	"testing"
	"time"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"golang.org/x/exp/slices"
	"pgregory.net/rapid"
)

func (s *RollappTestSuite) TestLivenessEvents() {
	ctx := &s.Ctx
	k := s.App.RollappKeeper
	_ = ctx
	_ = k
}

// Correct calculation of the next slash or jail event, based on downtime and parameters
// go test -run=TestNextSlashOrJailHeightRapid -rapid.checks=100 -rapid.steps=30000
func TestNextSlashOrJailHeightRapid(t *testing.T) {
	hubBlockInterval := 6 * time.Second
	slashTimeNoUpdate := 12 * time.Hour
	slashInterval := 1 * time.Hour
	jailTime := 48 * time.Hour

	hubBlockGap := rapid.Custom[time.Duration](func(t *rapid.T) time.Duration {
		if rapid.Bool().Draw(t, "hub is down") {
			dt := rapid.IntRange(int(time.Hour), int(time.Hour*24*7)).Draw(t, "dt")
			return time.Duration(dt)
		} else {
			return hubBlockInterval
		}
	})

	rapid.Check(t, func(r *rapid.T) {
		// model data
		hubHeight := int64(0)
		hubBlockTime := time.Time{}
		lastUpdateHeight := int64(0)
		nextEventHeight := int64(-1) // 'queue'
		nextEventIsJail := false     // 'queue'

		// for invariants
		jailed := false
		hubBlockTimeLastSlash := time.Time{}

		r.Repeat(map[string]func(*rapid.T){
			"": func(r *rapid.T) { // Check
			},
			// TODO: add operation for params change
			// TODO: add more invariants
			"hub end block": func(r *rapid.T) {
				if hubHeight == nextEventHeight {
					if nextEventIsJail {
						if jailed {
							r.Fatal("jailed already")
						}
						jailed = true
					} else {
						if !hubBlockTimeLastSlash.IsZero() && hubBlockTime.Sub(hubBlockTimeLastSlash) < slashInterval {
							r.Fatalf("slashed too frequently")
						}
						hubBlockTimeLastSlash = hubBlockTime
						nextEventHeight, nextEventIsJail = keeper.NextSlashOrJailHeight(
							hubBlockInterval,
							slashTimeNoUpdate,
							slashInterval,
							jailTime,
							hubHeight,
							lastUpdateHeight,
						)
					}
				}
				hubHeight += 1
				hubBlockTime.Add(hubBlockGap.Draw(r, "hub time increase"))
			},
			"update rollapp": func(r *rapid.T) {
				lastUpdateHeight = hubHeight
				// delete the scheduled event from the 'queue'
				nextEventHeight = -1
				nextEventIsJail = false
			},
		})
	})
}

// Storage and query operations work for the event queue
// go test -run=TestLivenessEventsStorage -rapid.checks=100 -rapid.steps=10
func TestLivenessEventsStorage(t *testing.T) {
	rollapps := rapid.StringMatching("^[a-zA-Z0-9]{1,10}$")
	heights := rapid.Int64Range(0, 10)
	isJail := rapid.Bool()
	rapid.Check(t, func(r *rapid.T) {
		k, ctx := keepertest.RollappKeeper(t)
		model := make(map[string]types.LivenessEvent)
		modelKey := func(e types.LivenessEvent) string {
			return fmt.Sprintf("%+v", e)
		}
		r.Repeat(map[string]func(r *rapid.T){
			"put": func(r *rapid.T) {
				e := types.LivenessEvent{
					RollappId: rollapps.Draw(r, "rollapp"),
					HubHeight: heights.Draw(r, "h"),
					IsJail:    isJail.Draw(r, "jail"),
				}
				k.PutLivenessEvent(ctx, e)
				model[modelKey(e)] = e
			},
			"deleteForRollapp": func(r *rapid.T) {
				e := types.LivenessEvent{
					RollappId: rollapps.Draw(r, "rollapp"),
					HubHeight: heights.Draw(r, "h"),
				}
				k.DelLivenessEvents(ctx, e.HubHeight, e.RollappId)
				delete(model, modelKey(e))
				e.IsJail = true
				delete(model, modelKey(e))
			},
			"iterHeight": func(r *rapid.T) {
				h := heights.Draw(r, "h")
				events := k.GetLivenessEvents(ctx, &h)
				for _, modelE := range model {
					if modelE.HubHeight == h && !slices.Contains(events, modelE) {
						r.Fatal("event in model but not store")
					}
				}
				for _, e := range events {
					_, ok := model[modelKey(e)]
					if !ok {
						r.Fatal("event in store but not model")
					}
				}
			},
			"iterAll": func(r *rapid.T) {
				events := k.GetLivenessEvents(ctx, nil)
				for _, modelE := range model {
					if !slices.Contains(events, modelE) {
						r.Fatal("event in model but not store")
					}
				}
				for _, e := range events {
					_, ok := model[modelKey(e)]
					if !ok {
						r.Fatal("event in store but not model")
					}
				}
			},
		})
	})
}

// An example, not intended to be run regularly, but rather for debugging
func TestNextSlashOrJailHeightExample(t *testing.T) {
	t.Skip()
	hHub := int64(0)
	jail := false
	for !jail {
		last := hHub
		hHub, jail = keeper.NextSlashOrJailHeight(
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
