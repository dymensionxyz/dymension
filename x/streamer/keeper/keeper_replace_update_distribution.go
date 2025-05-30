package keeper

import (
	"sort"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/streamer/types"
)

// This is checked for no err when a proposal is made, and executed when a proposal passes.
func (k Keeper) ReplaceDistrRecords(ctx sdk.Context, streamId uint64, records []types.DistrRecord) error {
	stream, err := k.GetStreamByID(ctx, streamId)
	if err != nil {
		return err
	}

	if stream.IsFinishedStream(ctx.BlockTime()) {
		return errorsmod.Wrapf(types.ErrInvalidStreamStatus, "stream %d is already finished", stream.Id)
	}

	distrInfo, err := k.NewDistrInfo(ctx, records)
	if err != nil {
		return err
	}

	stream.DistributeTo = distrInfo

	err = k.SetStream(ctx, stream)
	if err != nil {
		return err
	}

	return nil
}

// UpdateDistrRecords is checked for no err when a proposal is made, and executed when a proposal passes.
func (k Keeper) UpdateDistrRecords(ctx sdk.Context, streamId uint64, records []types.DistrRecord) error {
	recordsMap := make(map[uint64]types.DistrRecord)

	stream, err := k.GetStreamByID(ctx, streamId)
	if err != nil {
		return err
	}

	if stream.IsFinishedStream(ctx.BlockTime()) {
		return errorsmod.Wrapf(types.ErrInvalidStreamStatus, "stream %d is already finished", stream.Id)
	}

	err = k.validateGauges(ctx, records)
	if err != nil {
		return err
	}

	for _, existingRecord := range stream.DistributeTo.Records {
		recordsMap[existingRecord.GaugeId] = existingRecord
	}

	for _, record := range records {
		recordsMap[record.GaugeId] = record
	}

	newRecords := []types.DistrRecord{}
	for _, val := range recordsMap {
		if !val.Weight.Equal(math.ZeroInt()) {
			newRecords = append(newRecords, val)
		}
	}

	sort.SliceStable(newRecords, func(i, j int) bool {
		return newRecords[i].GaugeId < newRecords[j].GaugeId
	})

	distrInfo, err := k.NewDistrInfo(ctx, newRecords)
	if err != nil {
		return err
	}

	stream.DistributeTo = distrInfo

	err = k.SetStream(ctx, stream)
	if err != nil {
		return err
	}

	return nil
}
