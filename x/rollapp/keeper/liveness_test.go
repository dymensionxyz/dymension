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

// Storage and query operations work for the event queue
func TestLivenessEventsStorage(t *testing.T) {
	_ = flag.Set("rapid.checks", "50")
	_ = flag.Set("rapid.steps", "50")

	rollapps := rapid.StringMatching("^[a-zA-Z0-9]{1,10}$")
	heights := rapid.Int64Range(0, 10)
	isJail := rapid.Bool()
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
					require.False(r, modelE.HubHeight == h && !slices.Contains(events, modelE), "event in model but not store")
				}
				for _, e := range events {
					require.Contains(r, model, modelKey(e), "event in store but not model")
					require.Equal(r, h, e.HubHeight, "event from store has wrong height")
				}
			},
			"iterAll": func(r *rapid.T) {
				events := k.GetLivenessEvents(ctx, nil)
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

// The protocol works.
func (suite *RollappTestSuite) TestLivenessFlow() {
	_ = flag.Set("rapid.checks", "500")
	_ = flag.Set("rapid.steps", "300")
	rapid.Check(suite.T(), func(r *rapid.T) {
		suite.SetupTest()

		rollapps := []string{urand.RollappID(), urand.RollappID()}

		tracker := newLivenessMockSequencerKeeper()
		suite.keeper().SetSequencerKeeper(tracker)
		for _, ra := range rollapps {
			suite.keeper().SetRollapp(suite.Ctx, types.NewRollapp("", ra, "", "", "", types.Rollapp_Unspecified, nil, false))
		}

		hLastUpdate := map[string]int64{}
		rollappIsDown := map[string]bool{}

		r.Repeat(map[string]func(r *rapid.T){
			"": func(r *rapid.T) { // check
				// 1. check registered invariant
				msg, notOk := keeper.LivenessEventInvariant(*suite.keeper())(suite.Ctx)
				require.False(r, notOk, msg)
				// 2. check the right amount of slashing occurred
				for _, ra := range rollapps {
					h := suite.Ctx.BlockHeight()
					lastUpdate, ok := hLastUpdate[ra]
					if !ok {
						continue // we can freely assume we will not need to slash a rollapp if it has NEVER had an update
					}
					elapsed := uint64(h - lastUpdate)
					p := suite.keeper().GetParams(suite.Ctx)
					if elapsed <= p.LivenessJailBlocks {
						require.Zero(r, tracker.jails[ra], "expect not jailed")
					} else {
						require.NotZero(r, tracker.jails[ra], "expect jailed")
					}
					if elapsed <= p.LivenessSlashBlocks {
						l := tracker.slashes[ra]
						require.Zero(r, l, "expect not slashed")
					} else {
						expectedSlashes := int((elapsed-p.LivenessSlashBlocks)/p.LivenessSlashInterval) + 1
						require.Equal(r, expectedSlashes, tracker.slashes[ra], "expect slashed", "rollapp", ra, "elapsed blocks", elapsed)
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
					ra := suite.keeper().MustGetRollapp(suite.Ctx, raID)
					suite.keeper().IndicateLiveness(suite.Ctx, &ra)
					hLastUpdate[raID] = suite.Ctx.BlockHeight()
					tracker.clear(raID)
				}
			},
			"hub end blocks": func(r *rapid.T) {
				for range rapid.IntRange(0, 100).Draw(r, "num blocks") {
					suite.nextBlock()
				}
			},
		})
	})
}

type livenessMockSequencerKeeper struct {
	slashes map[string]int
	jails   map[string]int
}

func (l livenessMockSequencerKeeper) UnbondingTime(sdk.Context) (res time.Duration) {
	return time.Minute
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
