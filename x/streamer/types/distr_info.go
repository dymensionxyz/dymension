package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	sponsorshiptypes "github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func NewDistrInfo(records []DistrRecord) (DistrInfo, error) {
	distrInfo := DistrInfo{}
	totalWeight := sdk.NewInt(0)

	for _, record := range records {
		if err := record.ValidateBasic(); err != nil {
			return DistrInfo{}, err
		}
		totalWeight = totalWeight.Add(record.Weight)
	}

	distrInfo.Records = records
	distrInfo.TotalWeight = totalWeight

	if !totalWeight.IsPositive() {
		return DistrInfo{}, ErrDistrInfoNotPositiveWeight
	}

	return distrInfo, nil
}

// ValidateBasic is a basic validation test on recordd distribution gauges' weights.
func (r DistrRecord) ValidateBasic() error {
	if r.Weight.IsNegative() {
		return ErrDistrRecordNotPositiveWeight
	}
	return nil
}

// DistrInfoFromDistribution converts sponsorship distribution to the DistrInfo type. Returning an empty
// DistrInfo (with zero DistrInfo.TotalWeight) is a valid scenario. Note that some part of the distribution
// might be abstained. In that case, the sum of gauge powers would be less than the total distribution weight.
//
// Example! Let's say we have the following distribution in the sponsorship:
//
//	Total: 100
//	Gauge1: 30%
//	Gauge2: 50%
//	Abstained: 20%
//
// We want to distribute 100 DYM according to this distribution. Since 20% is abstained, we will normalize the figures
// for Gauge1 and Gauge2. The total "active" voting power is 80%, which comes from 30% (Gauge1) + 50% (Gauge2).
//
//	Gauge1: 30% / 80% = 37.5% (in the new distribution)
//	Gauge2: 50% / 80% = 62.5% (in the new distribution)
//
// So, Gauge1 gets 37.5 DYM, and Gauge2 gets 62.5 DYM.
func DistrInfoFromDistribution(d sponsorshiptypes.Distribution) DistrInfo {
	totalWeight := math.ZeroInt()

	records := make([]DistrRecord, 0, len(d.Gauges))
	for _, g := range d.Gauges {
		records = append(records, DistrRecord{
			GaugeId: g.GaugeId,
			Weight:  g.Power,
		})
		totalWeight = totalWeight.Add(g.Power)
	}

	return DistrInfo{
		TotalWeight: totalWeight,
		Records:     records,
	}
}
