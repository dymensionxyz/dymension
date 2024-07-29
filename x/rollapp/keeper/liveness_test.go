package keeper

import (
	"fmt"
	"testing"
	"time"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"golang.org/x/exp/slices"
	"pgregory.net/rapid"
)

// TODO: I need to rethink, because you can have more tha one event for a rollapp at a given height

// An example, not intended to be run regularly
func TestNextSlashOrJailHeightExample(t *testing.T) {
	t.Skip()
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

// go test -run=TestNextSlashOrJailHeightRapid -rapid.checks=100 -rapid.steps=30000
func TestNextSlashOrJailHeightRapid(t *testing.T) {
	/*
	  -rapid.checks int
	    	rapid: number of checks to perform (default 100)
	  -rapid.debug
	    	rapid: debugging output
	  -rapid.debugvis
	    	rapid: debugging visualization
	  -rapid.failfile string
	    	rapid: fail file to use to reproduce test failure
	  -rapid.log
	    	rapid: eager verbose output to stdout (to aid with unrecoverable test failures)
	  -rapid.nofailfile
	    	rapid: do not write fail files on test failures
	  -rapid.seed uint
	    	rapid: PRNG seed to start with (0 to use a random one)
	  -rapid.shrinktime duration
	    	rapid: maximum time to spend on test case minimization (default 30s)
	  -rapid.steps int
	    	rapid: average number of Repeat actions to execute (default 30)
	  -rapid.v
	    	rapid: verbose output
	*/
	rapid.Check(t, testWithRapid)
}

func testWithRapid(t *rapid.T) {
	hubHeight := int64(0)
	hubBlockTime := time.Time{}
	lastUpdateHeight := int64(0)
	nextEventHeight := int64(-1)
	nextEventIsJail := false
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

	// used for invariants
	jailed := false
	numSlashes := 0
	hubBlockTimeLastSlash := time.Time{}

	ops := map[string]func(*rapid.T){
		"": func(t *rapid.T) { // Check
		},
		"hub end block": func(t *rapid.T) {
			if hubHeight == nextEventHeight {
				if nextEventIsJail {
					if jailed {
						t.Fatal("jailed already")
					}
					jailed = true
				} else {
					if !hubBlockTimeLastSlash.IsZero() && hubBlockTime.Sub(hubBlockTimeLastSlash) < slashInterval {
						t.Fatalf("slashed too frequently")
					}
					hubBlockTimeLastSlash = hubBlockTime
					numSlashes += 1
					nextEventHeight, nextEventIsJail = NextSlashOrJailHeight(
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
			hubBlockTime.Add(hubBlockGap.Draw(t, "hub time increase"))
		},
		"update rollapp": func(t *rapid.T) {
			lastUpdateHeight = hubHeight
			// delete the scheduled event from the 'queue'
			nextEventHeight = -1
			nextEventIsJail = false
		},
		// TODO: add capability to change params on the fly
	}

	t.Repeat(ops)
}

// go test -run=TestLivenessEventsStorage -rapid.checks=100 -rapid.steps=10
func TestLivenessEventsStorage(t *testing.T) {
	k, ctx := keepertest.RollappKeeper(t)
	rollapps := rapid.SampledFrom([]string{"a", "b", "c"})
	heights := rapid.Int64Range(0, 10)
	isJail := rapid.Bool()

	model := make(map[string]types.LivenessEvent)
	modelKey := func(e types.LivenessEvent) string {
		return fmt.Sprintf("%+v", e)
	}

	f := func(r *rapid.T) {
		ops := map[string]func(r *rapid.T){
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
		}
		r.Repeat(ops)
	}
	rapid.Check(t, f)
}
