package keeper

import (
	"github.com/dymensionxyz/dymension/v3/x/streamer/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func (k Keeper) NewDistrInfo(ctx sdk.Context, records []types.DistrRecord) (*types.DistrInfo, error) {
	err := k.validateGauges(ctx, records)
	if err != nil {
		return nil, err
	}

	distrInfo, err := types.NewDistrInfo(records)
	if err != nil {
		return nil, err
	}

	return distrInfo, nil
}

// validateGauges validates a list of records to ensure that:
// 1) there are no duplicates,
// 2) the records are in sorted order.
// 3) the records only pay to gauges that exist.
func (k Keeper) validateGauges(ctx sdk.Context, records []types.DistrRecord) error {
	lastGaugeID := uint64(0)
	gaugeIdFlags := make(map[uint64]bool)

	for _, record := range records {
		if gaugeIdFlags[record.GaugeId] {
			return sdkerrors.Wrapf(
				types.ErrDistrRecordRegisteredGauge,
				"Gauge ID #%d has duplications.",
				record.GaugeId,
			)
		}

		// Ensure records are sorted because ~AESTHETIC~
		if record.GaugeId < lastGaugeID {
			return sdkerrors.Wrapf(
				types.ErrDistrRecordNotSorted,
				"Gauge ID #%d came after Gauge ID #%d.",
				record.GaugeId, lastGaugeID,
			)
		}
		lastGaugeID = record.GaugeId

		// don't allow distribution records for gauges that don't exist
		gauge, err := k.ik.GetGaugeByID(ctx, record.GaugeId)
		if err != nil {
			return err
		}
		if !gauge.IsPerpetual {
			return sdkerrors.Wrapf(types.ErrDistrRecordRegisteredGauge,
				"Gauge ID #%d is not perpetual.",
				record.GaugeId)
		}

		gaugeIdFlags[record.GaugeId] = true
	}
	return nil
}
