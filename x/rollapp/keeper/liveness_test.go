package keeper_test

import (
	"flag"
	"fmt"
	"slices"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/urand"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func TestLivenessArithmetic(t *testing.T) {
	t.Run("simple case", func(t *testing.T) {
		hEvent := keeper.NextSlashHeight(
			8,
			4,
			0,
			0,
		)
		require.Equal(t, 8, int(hEvent))
	})
	t.Run("almost at the interval", func(t *testing.T) {
		hEvent := keeper.NextSlashHeight(
			8,
			4,
			7,
			0,
		)
		require.Equal(t, 8, int(hEvent))
	})
	t.Run("do not schedule for next height (1)", func(t *testing.T) {
		hEvent := keeper.NextSlashHeight(
			8,
			4,
			8,
			0,
		)
		require.Equal(t, 12, int(hEvent))
	})
	t.Run("do not schedule for next height (2)", func(t *testing.T) {
		hEvent := keeper.NextSlashHeight(
			8,
			4,
			12,
			0,
		)
		require.Equal(t, 16, int(hEvent))
	})
}

func TestCannotScheduleForPast(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		var (
			noUpdate          = rapid.Uint64Range(1, 100).Draw(t, "noUpdate")
			slashInterval     = rapid.Uint64Range(1, 100).Draw(t, "slashInterval")
			heightHub         = rapid.Int64Range(0, 100).Draw(t, "heightHub")
			lastRollappUpdate = rapid.Int64Range(0, 100).Draw(t, "lastRollappUpdate")
		)
		res := keeper.NextSlashHeight(noUpdate, slashInterval, heightHub, lastRollappUpdate)
		if res <= heightHub {
			t.Fatalf(
				"no update %d, interval %d, hub %d, last update %d, res %d",
				noUpdate, slashInterval, heightHub, lastRollappUpdate, res,
			)
		}
	})
}

// Storage and query operations work for the event queue
func TestLivenessEventsStorage(t *testing.T) {
	_ = flag.Set("rapid.checks", "50")
	_ = flag.Set("rapid.steps", "50")

	rollapps := rapid.StringMatching("^[a-zA-Z0-9]{1,10}$")
	heights := rapid.Int64Range(0, 10)
	rapid.Check(t, func(r *rapid.T) {
		k, ctx := keepertest.RollappKeeper(t)
		model := make(map[string]types.LivenessEvent) // model actual sdk storage
		modelKey := func(e types.LivenessEvent) string {
			return fmt.Sprintf("%+v", e)
		}
		r.Repeat(map[string]func(r *rapid.T){
			"put": func(r *rapid.T) {
				e := types.LivenessEvent{
					RollappId: rollapps.Draw(r, "rollapp"),
					HubHeight: heights.Draw(r, "h"),
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
				delete(model, modelKey(e))
			},
			"iterHeight": func(r *rapid.T) {
				h := heights.Draw(r, "h")
				events := k.GetLivenessEvents(ctx, &h)
				for _, modelE := range model {
					require.False(r, modelE.HubHeight == h && !slices.Contains(events, modelE), "event in model but not store")
				}
				for _, e := range events {
					require.Contains(r, model, modelKey(e), "event in store but not model")
					require.Equal(r, h, e.HubHeight, "event from store has wrong height")
				}
			},
			"iterAll": func(r *rapid.T) {
				events := k.GetLivenessEvents(ctx, nil)
				for i := 1; i < len(events); i++ {
					require.LessOrEqual(r, events[i-1].HubHeight, events[i].HubHeight, "events not sorted")
					i++
				}
				for _, modelE := range model {
					require.Contains(r, events, modelE, "event in model but not store")
				}
				for _, e := range events {
					require.Contains(r, model, modelKey(e), "event in store but not model")
				}
			},
		})
	})
}

func (s *RollappTestSuite) TestLivenessEndBlock() {
	p := s.k().GetParams(s.Ctx)
	p.LivenessSlashBlocks = 2
	s.k().SetParams(s.Ctx, p)
	tracker := newLivenessMockSequencerKeeper(s.k().SequencerK)
	s.k().SetSequencerKeeper(tracker)
	rollapp, proposer := s.CreateDefaultRollappAndProposer()
	_, err := s.PostStateUpdate(s.Ctx, rollapp, proposer, 1, uint64(10))
	s.Require().NoError(err)
	for range p.LivenessSlashBlocks + 1 {
		s.Require().Equal(0, tracker.slashes[rollapp])
		s.checkLiveness(rollapp, false, true)
		s.NextBlock(time.Second)
	}
	s.Require().Equal(1, tracker.slashes[rollapp])
	s.checkLiveness(rollapp, false, true)
}

func (s *RollappTestSuite) checkLiveness(rollappId string, expectClockReset, expectEvent bool) {
	msg, broken := keeper.LivenessEventInvariant(*s.k())(s.Ctx)
	s.Require().False(broken, msg)
	ra := s.k().MustGetRollapp(s.Ctx, rollappId)
	if expectClockReset {
		s.Require().Equal(s.Ctx.BlockHeight(), ra.LivenessCountdownStartHeight)
	} else {
		s.Require().LessOrEqual(ra.LivenessCountdownStartHeight, s.Ctx.BlockHeight())
	}
	s.Require().Equal(expectEvent, ra.LivenessEventHeight != 0)
}

// The protocol works.
func (s *RollappTestSuite) TestLivenessFlow() {
	_ = flag.Set("rapid.checks", "400")
	_ = flag.Set("rapid.steps", "200")
	rapid.Check(s.T(), func(r *rapid.T) {
		s.SetupTest()
		p := s.k().GetParams(s.Ctx)
		p.LivenessSlashBlocks = 5
		p.LivenessSlashInterval = 3
		s.k().SetParams(s.Ctx, p)

		rollapps := []string{urand.RollappID(), urand.RollappID()}

		tracker := newLivenessMockSequencerKeeper(s.k().SequencerK)
		s.k().SetSequencerKeeper(tracker)
		for _, ra := range rollapps {
			s.k().SetRollapp(s.Ctx, types.NewRollapp("", ra, "", types.DefaultMinSequencerBondGlobalCoin, types.Rollapp_Unspecified, nil, types.GenesisInfo{}))
		}

		hClockStart := map[string]int64{}

		r.Repeat(map[string]func(r *rapid.T){
			"": func(r *rapid.T) { // check
				// 1. check registered invariant
				msg, notOk := keeper.LivenessEventInvariant(*s.k())(s.Ctx)
				require.False(r, notOk, msg)
				// 2. check the right amount of slashing occurred
				for _, ra := range rollapps {
					h := s.Ctx.BlockHeight()
					hStart, ok := hClockStart[ra]
					if !ok {
						// clock not started (can be before first state update, or during fork)
						continue
					}
					p := s.k().GetParams(s.Ctx)
					elapsed := h - 1 - hStart // num end blockers that have passed since setting the clock start
					if elapsed <= 0 {
						continue
					}
					if elapsed < int64(p.LivenessSlashBlocks) {
						l := tracker.slashes[ra]
						require.Zero(r, l, "expect not slashed", "elapsed", elapsed, "rollapp", ra, "h", s.Ctx.BlockHeight())
					} else {
						expectedSlashes := int((elapsed-int64(p.LivenessSlashBlocks))/int64(p.LivenessSlashInterval)) + 1
						require.Equal(r, expectedSlashes, tracker.slashes[ra], "expect slashed",
							"rollapp", ra, "elapsed blocks", elapsed, "h", s.Ctx.BlockHeight())
					}
				}
			},
			"indicate liveness (new proposer + state update)": func(r *rapid.T) {
				raID := rapid.SampledFrom(rollapps).Draw(r, "rollapp")
				ra := s.k().MustGetRollapp(s.Ctx, raID)
				// use intrusive method. Wiring is checked in state update test and hook test.
				s.k().IndicateLiveness(s.Ctx, &ra)
				s.k().SetRollapp(s.Ctx, ra)
				hClockStart[raID] = s.Ctx.BlockHeight()
				tracker.clear(raID)
			},
			"fork": func(r *rapid.T) {
				raID := rapid.SampledFrom(rollapps).Draw(r, "rollapp")
				ra := s.k().MustGetRollapp(s.Ctx, raID)
				// use intrusive method. Wiring is checked in fork test.
				s.k().ResetLivenessClock(s.Ctx, &ra)
				s.k().SetRollapp(s.Ctx, ra)
				delete(hClockStart, raID)
				tracker.clear(raID)
			},
			"end blocker": func(r *rapid.T) {
				for range rapid.IntRange(0, 9).Draw(r, "num blocks") {
					r.Log("Doing EB", s.Ctx.BlockHeight(), "slashes", tracker.slashes)
					s.NextBlock(time.Second)
				}
			},
		})
	})
}

type livenessMockSequencerKeeper struct {
	keeper.SequencerKeeper
	slashes map[string]int // rollapp->cnt
}

func newLivenessMockSequencerKeeper(k keeper.SequencerKeeper) livenessMockSequencerKeeper {
	return livenessMockSequencerKeeper{
		SequencerKeeper: k,
		slashes:         make(map[string]int),
	}
}

func (l livenessMockSequencerKeeper) SlashLiveness(ctx sdk.Context, rollappID string) error {
	l.slashes[rollappID]++
	return nil
}

func (l livenessMockSequencerKeeper) clear(rollappID string) {
	delete(l.slashes, rollappID)
}
