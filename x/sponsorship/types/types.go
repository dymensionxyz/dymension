package types

import "cosmossdk.io/math"

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
