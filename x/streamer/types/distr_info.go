package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func NewDistrInfo(records []DistrRecord) (*DistrInfo, error) {
	distrInfo := DistrInfo{}
	totalWeight := sdk.NewInt(0)

	for _, record := range records {
		if err := record.ValidateBasic(); err != nil {
			return nil, err
		}
		totalWeight = totalWeight.Add(record.Weight)
	}

	distrInfo.Records = records
	distrInfo.TotalWeight = totalWeight

	if !totalWeight.IsPositive() {
		return nil, ErrDistrInfoNotPositiveWeight
	}

	return &distrInfo, nil
}

// ValidateBasic is a basic validation test on recordd distribution gauges' weights.
func (r DistrRecord) ValidateBasic() error {
	if r.Weight.IsNegative() {
		return ErrDistrRecordNotPositiveWeight
	}
	return nil
}

var hundred = math.NewInt(100)

func DistrInfoFromDistribution(d sponsorshiptypes.Distribution) *DistrInfo {
	totalWeight := math.ZeroInt()
	records := make([]DistrRecord, 0, len(d.Gauges))
	for _, g := range d.Gauges {
		weight := g.Power.Quo(d.VotingPower).Mul(hundred)

		totalWeight = totalWeight.Add(weight)
		records = append(records, DistrRecord{
			GaugeId: g.GaugeId,
			Weight:  weight,
		})
	}
	return &DistrInfo{
		TotalWeight: totalWeight,
		Records:     records,
	}
}
