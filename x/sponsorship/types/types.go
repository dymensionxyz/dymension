package types

import (
	"slices"
	"sort"

	"cosmossdk.io/math"
)

func ValidateGaugeWeights(w []GaugeWeight) error {
	total := math.ZeroInt()
	gaugeIDs := make(map[uint64]struct{}, len(w)) // this map helps check for duplicates
	for _, g := range w {
		err := g.Validate()
		if err != nil {
			return ErrInvalidDistribution.Wrap(err.Error())
		}
		if _, ok := gaugeIDs[g.GaugeId]; ok {
			return ErrInvalidDistribution.Wrapf("duplicated gauge id: %d", g.GaugeId)
		}
		gaugeIDs[g.GaugeId] = struct{}{}
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
	return ApplyWeights(v.VotingPower, v.Weights)
}

func ApplyWeights(votingPower math.Int, weights []GaugeWeight) Distribution {
	gauges := make(Gauges, 0, len(weights))
	for _, weight := range weights {
		gauges = append(gauges, Gauge{
			GaugeId: weight.GetGaugeId(),
			Power:   votingPower.Mul(weight.Weight).Quo(hundred),
		})
	}

	// All gauges must be sorted by the gauge ID
	sort.Sort(gauges)
	return Distribution{
		VotingPower: votingPower,
		Gauges:      gauges,
	}
}

func NewDistribution() Distribution {
	return Distribution{
		VotingPower: math.ZeroInt(),
		Gauges:      make([]Gauge, 0),
	}
}

// Merge is a binary associative and commutative operation over Distribution. It takes two
// distributions and applies one to another. Example:
func (d Distribution) Merge(d1 Distribution) Distribution {
	// O(n+m) solution based on modified https://leetcode.com/problems/merge-sorted-array.
	gauges := make(Gauges, 0, len(d.Gauges)+len(d1.Gauges))
	var i = 0           // first iterator
	var j = 0           // second iterator
	var lhs = d.Gauges  // alias
	var rhs = d1.Gauges // alias

	for i < len(lhs) && j < len(rhs) {
		var gauge Gauge
		switch {
		case lhs[i].GaugeId == rhs[j].GaugeId:
			gauge = Gauge{
				GaugeId: lhs[i].GaugeId,
				Power:   lhs[i].Power.Add(rhs[j].Power),
			}
			i++
			j++
		case lhs[i].GaugeId < rhs[j].GaugeId:
			gauge = lhs[i]
			i++
		case lhs[i].GaugeId > rhs[j].GaugeId:
			gauge = rhs[j]
			j++
		}
		// Don't store gauges with zero power
		if !gauge.Power.IsZero() {
			gauges = append(gauges, gauge)
		}
	}

	if i != len(lhs) {
		gauges = append(gauges, lhs[i:]...)
	}
	if j != len(rhs) {
		gauges = append(gauges, rhs[j:]...)
	}

	return Distribution{
		VotingPower: d.VotingPower.Add(d1.VotingPower),
		Gauges:      slices.Clip(gauges),
	}
}

func (d Distribution) Negate() Distribution {
	gauges := make([]Gauge, len(d.Gauges))
	for i, g := range d.Gauges {
		gauges[i] = Gauge{
			GaugeId: g.GaugeId,
			Power:   g.Power.Neg(),
		}
	}
	return Distribution{
		VotingPower: d.VotingPower.Neg(),
		Gauges:      gauges,
	}
}

func (d Distribution) Equal(d1 Distribution) bool {
	return d.VotingPower.Equal(d1.VotingPower) && slices.EqualFunc(d.Gauges, d1.Gauges, func(g1 Gauge, g2 Gauge) bool {
		return g1.GaugeId == g2.GaugeId && g1.Power.Equal(g2.Power)
	})
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
