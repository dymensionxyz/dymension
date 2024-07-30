package keeper_test

import (
	"fmt"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	flag "github.com/spf13/pflag"
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

func (l livenessMockSequencerKeeper) clear(rollappID string) {
	delete(l.slashes, rollappID)
	delete(l.jails, rollappID)
}

// The protocol works.
func TestLivenessFlow(t *testing.T) {
	_ = flag.Set("rapid.checks", "500")
	_ = flag.Set("rapid.steps", "300")
	rapid.Check(t, func(r *rapid.T) {
		s := new(RollappTestSuite)
		s.SetT(t)
		s.SetS(s)
		s.SetupTest()
		p := s.keeper().GetParams(s.Ctx)
		// adjust params to be more amenable to testing without needing thousands of hub blocks
		p.HubExpectedBlockTime = time.Minute * 25
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
				// 1. check registered invariant
				msg, ok := keeper.LivenessEventInvariant(*s.keeper())(s.Ctx)
				require.True(t, ok, msg)
				// 2. check the right amount of slashing occurred
				for _, ra := range rollapps {
					h := s.Ctx.BlockHeight()
					lastUpdate, ok := hLastUpdate[ra]
					if !ok {
						continue // we can freely assume we will not need to slash a rollapp if it has NEVER had an update
					}
					elapsed := h - lastUpdate
					p := s.keeper().GetParams(s.Ctx)
					elapsedTime := time.Duration(elapsed) * p.HubExpectedBlockTime
					if elapsedTime <= p.LivenessJailTime {
						require.Zero(r, tracker.jails[ra], "expect not jailed")
					} else {
						require.NotZerof(r, tracker.jails[ra], "expect jailed")
						t.Log("check jail")
					}
					if elapsedTime <= p.LivenessSlashTime {
						l := tracker.slashes[ra]
						require.Zero(r, l, "expect not slashed")
					} else {
						t.Log("check slash")
						expectedSlashes := int((elapsedTime-p.LivenessSlashTime)/p.LivenessSlashInterval) + 1
						require.Equal(r, expectedSlashes, tracker.slashes[ra], "expect slashed", "rollapp", ra)
					}
				}
			},
			"set sequencer status": func(r *rapid.T) {
				raID := rapid.SampledFrom(rollapps).Draw(r, "rollapp")
				rollappIsDown[raID] = rapid.Bool().Draw(r, "down")
			},
			"state update": func(r *rapid.T) {
				raID := rapid.SampledFrom(rollapps).Draw(r, "rollapp")
				if !rollappIsDown[raID] {
					ra := s.keeper().MustGetRollapp(s.Ctx, raID)
					s.keeper().IndicateLiveness(s.Ctx, &ra)
					s.keeper().SetRollapp(s.Ctx, ra)
					hLastUpdate[raID] = s.Ctx.BlockHeight()
					tracker.clear(raID)
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
func TestNextSlashOrJailHeightRapid(t *testing.T) {
	_ = flag.Set("rapid.checks", "100")
	_ = flag.Set("rapid.steps", "30000")

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
func TestLivenessEventsStorage(t *testing.T) {
	_ = flag.Set("rapid.checks", "50")
	_ = flag.Set("rapid.steps", "50")

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
