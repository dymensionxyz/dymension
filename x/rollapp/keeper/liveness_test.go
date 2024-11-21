package keeper_test

import (
	"flag"
	"fmt"
	"slices"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/urand"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	keepertest "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	seqtypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
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
	t.Run("do not schedule for next height", func(t *testing.T) {
		hEvent := keeper.NextSlashHeight(
			8,
			4,
			8,
			0,
		)
		require.Equal(t, 12, int(hEvent))
	})
	t.Run("do not schedule for next height", func(t *testing.T) {
		hEvent := keeper.NextSlashHeight(
			8,
			4,
			12,
			0,
		)
		require.Equal(t, 16, int(hEvent))
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

func (s *RollappTestSuite) TestLivenessFlow2() {
	/*
		What do I want to test?
		- Everything is wired properly
		- Invariants hold
		- They are slashed as they should be

		What is the liveness flow?

	*/

}

// The protocol works.
func (s *RollappTestSuite) TestLivenessFlow() {
	_ = flag.Set("rapid.checks", "500")
	_ = flag.Set("rapid.steps", "300")
	rapid.Check(s.T(), func(r *rapid.T) {
		s.SetupTest()

		rollapps := []string{urand.RollappID(), urand.RollappID()}

		tracker := newLivenessMockSequencerKeeper()
		s.k().SetSequencerKeeper(tracker)
		for _, ra := range rollapps {
			s.k().SetRollapp(s.Ctx, types.NewRollapp("", ra, "", types.Rollapp_Unspecified, nil, types.GenesisInfo{}, false))
		}

		hLastUpdate := map[string]int64{}
		rollappIsDown := map[string]bool{}

		r.Repeat(map[string]func(r *rapid.T){
			"": func(r *rapid.T) { // check
				// 1. check registered invariant
				msg, notOk := keeper.LivenessEventInvariant(*s.k())(s.Ctx)
				require.False(r, notOk, msg)
				// 2. check the right amount of slashing occurred
				for _, ra := range rollapps {
					h := s.Ctx.BlockHeight()
					lastUpdate, ok := hLastUpdate[ra]
					if !ok {
						continue // we can freely assume we will not need to slash a rollapp if it has NEVER had an update
					}
					elapsed := uint64(h - lastUpdate)
					p := s.k().GetParams(s.Ctx)

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
					ra := s.k().MustGetRollapp(s.Ctx, raID)
					s.k().IndicateLiveness(s.Ctx, &ra)
					s.k().SetRollapp(s.Ctx, ra)
					hLastUpdate[raID] = s.Ctx.BlockHeight()
					tracker.clear(raID)
				}
			},
			"hub end blocks": func(r *rapid.T) {
				for range rapid.IntRange(0, 100).Draw(r, "num blocks") {
					h := s.Ctx.BlockHeight()
					s.Ctx = s.Ctx.WithBlockHeight(h + 1)
				}
			},
		})
	})
}

type livenessMockSequencerKeeper struct {
	slashes map[string]int
}

// GetSuccessor implements keeper.SequencerKeeper.
func (l livenessMockSequencerKeeper) GetSuccessor(ctx sdk.Context, rollapp string) seqtypes.Sequencer {
	panic("unimplemented")
}

func (l livenessMockSequencerKeeper) PunishSequencer(ctx sdk.Context, seqAddr string, rewardee *sdk.AccAddress) error {
	// only relevant in HF
	return nil
}

func newLivenessMockSequencerKeeper() livenessMockSequencerKeeper {
	return livenessMockSequencerKeeper{
		make(map[string]int),
	}
}

func (l livenessMockSequencerKeeper) SlashLiveness(ctx sdk.Context, rollappID string) error {
	l.slashes[rollappID]++
	return nil
}

func (l livenessMockSequencerKeeper) GetProposer(ctx sdk.Context, rollappId string) seqtypes.Sequencer {
	return seqtypes.Sequencer{}
}

func (l livenessMockSequencerKeeper) clear(rollappID string) {
	delete(l.slashes, rollappID)
}
