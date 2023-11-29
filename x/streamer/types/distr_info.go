package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	return &distrInfo, nil
}

func (d DistrInfo) Validate() error {
	totalWeight := sdk.NewInt(0)
	for _, record := range d.Records {
		totalWeight = totalWeight.Add(record.Weight)
	}

	if !totalWeight.Equal(d.TotalWeight) {
		return ErrDistrInfoTotalWeightNotEqual
	}
	return nil
}

// ValidateBasic is a basic validation test on recordd distribution gauges' weights.
func (r DistrRecord) ValidateBasic() error {
	if r.Weight.IsNegative() {
		return ErrDistrRecordNotPositiveWeight
	}
	return nil
}
