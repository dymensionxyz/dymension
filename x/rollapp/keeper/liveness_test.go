package keeper_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"
	"pgregory.net/rapid"
)

type livenessMockSequencerKeeper struct {
	slashes map[string]int
	jails   map[string]int
}

func newLivenessMockSequencerKeeper() livenessMockSequencerKeeper {
	return livenessMockSequencerKeeper{
		make(map[string]int),
		make(map[string]int),
	}
}

func (l livenessMockSequencerKeeper) SlashLiveness(ctx sdk.Context, rollappID string) error {
	l.slashes[rollappID]++
	return nil
}

func (l livenessMockSequencerKeeper) JailLiveness(ctx sdk.Context, rollappID string) error {
	l.jails[rollappID]++
	return nil
}

func (l livenessMockSequencerKeeper) clear() {
	l.slashes = make(map[string]int)
	l.jails = make(map[string]int)
}

// The protocol works.
// go test -run=TestLivenessFlow -rapid.checks=1000 -rapid.steps=300
func TestLivenessFlow(t *testing.T) {
	rapid.Check(t, func(r *rapid.T) {
		s := new(RollappTestSuite)
		s.SetT(t)
		s.SetS(s)
		s.SetupTest()
		p := s.keeper().GetParams(s.Ctx)
		// adjust params to be more amenable to testing without needing thousands of hub blocks
		p.HubExpectedBlockTime = time.Minute * 20
		s.keeper().SetParams(s.Ctx, p)

		rollapps := []string{"a", "b"}
		hubBlockGap := rapid.Custom[time.Duration](func(t *rapid.T) time.Duration {
			if rapid.Bool().Draw(t, "hub is down") {
				dt := rapid.IntRange(int(time.Hour), int(time.Hour*24*7)).Draw(t, "dt")
				return time.Duration(dt)
			} else {
				return s.keeper().GetParams(s.Ctx).HubExpectedBlockTime
			}
		})

		tracker := newLivenessMockSequencerKeeper()
		s.keeper().SetSequencerKeeper(tracker)
		for _, ra := range rollapps {
			s.keeper().SetRollapp(s.Ctx, types.Rollapp{RollappId: ra})
		}

		hLastUpdate := map[string]int64{}
		rollappIsDown := map[string]bool{}

		r.Repeat(map[string]func(r *rapid.T){
			"": func(r *rapid.T) { // check
				t.Log("check general", "h", s.Ctx.BlockHeight(), "ha", hLastUpdate["a"], "hb", hLastUpdate["b"])
				for _, ra := range rollapps {
					h := s.Ctx.BlockHeight()
					elapsed := h - hLastUpdate[ra]
					p := s.keeper().GetParams(s.Ctx)
					elapsedTime := time.Duration(elapsed) * p.HubExpectedBlockTime
					if elapsedTime <= p.LivenessJailTime {
						require.Zero(r, tracker.jails[ra])
					} else {
						require.NotZero(r, tracker.jails[ra])
						t.Log("check jail")
					}
					if elapsedTime <= p.LivenessSlashTime {
						require.Zero(r, tracker.slashes[ra])
					} else {
						t.Log("check slash")
						expectedSlashes := (elapsedTime-p.LivenessSlashTime)/p.LivenessSlashInterval + 1
						require.Equal(r, expectedSlashes, tracker.slashes[ra])
					}
				}
			},
			"toggle status": func(r *rapid.T) {
				raID := rapid.SampledFrom(rollapps).Draw(r, "rollapp")
				rollappIsDown[raID] = !rollappIsDown[raID]
			},
			"state update": func(r *rapid.T) {
				// if rapid.IntRange(0, 10).Draw(r, "up") < 1 {
				// if rapid.Bool().Draw(r, "sequencer is up")
				raID := rapid.SampledFrom(rollapps).Draw(r, "rollapp")
				if !rollappIsDown[raID] {
					ra := s.keeper().MustGetRollapp(s.Ctx, raID)
					s.keeper().IndicateLiveness(s.Ctx, &ra)
					s.keeper().SetRollapp(s.Ctx, ra)
					hLastUpdate[raID] = s.Ctx.BlockHeight()
					tracker.clear()
				}
			},
			"hub end block": func(r *rapid.T) {
				dt := hubBlockGap.Draw(r, "dt")
				s.nextBlock(dt)
				s.keeper().CheckLiveness(s.Ctx)
			},
		})
	})
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
			"update rollapp": func(r *rapid.T) {
				lastUpdateHeight = hubHeight
				// delete the scheduled event from the 'queue'
				nextEventHeight = -1
				nextEventIsJail = false
			},
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
		})
	})
}

// Storage and query operations work for the event queue
// go test -run=TestLivenessEventsStorage -rapid.checks=100 -rapid.steps=100
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
