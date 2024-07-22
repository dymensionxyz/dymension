package types

import (
	"slices"
	"sort"

	"cosmossdk.io/math"
)

func ValidateGaugeWeights(w []GaugeWeight) error {
	total := math.ZeroInt()
	for _, g := range w {
		err := g.Validate()
		if err != nil {
			return ErrInvalidDistribution.Wrap(err.Error())
		}
		total = total.Add(g.Weight)
	}
	if !total.Equal(hundred) {
		return ErrInvalidDistribution.Wrapf("total weight must equal 100, got %d", total.Int64())
	}
	return nil
}

func (g GaugeWeight) Validate() error {
	if !g.Weight.IsPositive() {
		return ErrInvalidGaugeWeight.Wrapf("weight must be > 0, got %d", g.Weight.Int64())
	}
	if g.Weight.GT(hundred) {
		return ErrInvalidGaugeWeight.Wrapf("weight must be <= 100, got %d", g.Weight.Int64())
	}
	return nil
}

// ToDistribution multiplies each gauge weight by the voting power to get its absolut voting power.
func (v Vote) ToDistribution() Distribution {
	gauges := make(Gauges, 0, len(v.Weights))
	for _, weight := range v.Weights {
		gauges = append(gauges, Gauge{
			GaugeId: weight.GetGaugeId(),
			Power:   v.VotingPower.Mul(weight.Weight).Quo(hundred),
		})
	}

	// All gauges must be sorted by the gauge ID
	sort.Sort(gauges)
	return Distribution{
		VotingPower: v.VotingPower,
		Gauges:      gauges,
	}
}

// TODO: add tests!
func ApplyUpdate(initial, update Distribution) Distribution {
	// O(n+m) solution based on modified https://leetcode.com/problems/merge-sorted-array.
	gauges := make(Gauges, 0, len(initial.Gauges)+len(update.Gauges))
	var i = 0 // initial iterator
	var j = 0 // update iterator
	var k = 0 // result iterator
	var ini = initial.Gauges
	var upd = update.Gauges
	for i < len(ini) && j < len(upd) {
		switch {
		case ini[i].GaugeId == upd[j].GaugeId:
			gauges[k] = Gauge{
				GaugeId: ini[i].GaugeId,
				Power:   ini[i].Power.Add(upd[j].Power),
			}
		case ini[i].GaugeId < upd[j].GaugeId:
			gauges[k] = initial.Gauges[i]
			i++
		case ini[i].GaugeId > upd[j].GaugeId:
			gauges[k] = initial.Gauges[i]
			j++
		}
		k++
	}
	slices.Clip(gauges)
	return Distribution{
		VotingPower: initial.VotingPower.Add(update.VotingPower),
		Gauges:      gauges,
	}
}

var _ sort.Interface = Gauges{}

type Gauges []Gauge

func (m Gauges) Len() int {
	return len(m)
}

func (m Gauges) Less(i, j int) bool {
	return m[i].GetGaugeId() < m[j].GetGaugeId()
}

func (m Gauges) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}
