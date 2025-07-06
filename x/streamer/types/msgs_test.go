package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
)

func TestMsgCreateStream_ValidateBasic(t *testing.T) {
	validAuthority := sample.AccAddress()
	validCoins := sdk.NewCoins(sdk.NewCoin("udym", math.NewInt(1000000)))
	validDistributeToRecords := []DistrRecord{
		{
			GaugeId: 1,
			Weight:  math.NewInt(100),
		},
	}
	validStartTime := time.Now()
	validEpochIdentifier := "day"
	validNumEpochs := uint64(30)

	tests := []struct {
		name    string
		msg     MsgCreateStream
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message with distribute to records",
			msg: MsgCreateStream{
				Authority:            validAuthority,
				DistributeToRecords:  validDistributeToRecords,
				Coins:                validCoins,
				StartTime:            validStartTime,
				DistrEpochIdentifier: validEpochIdentifier,
				NumEpochsPaidOver:    validNumEpochs,
				Sponsored:            false,
			},
			wantErr: false,
		},
		{
			name: "valid message with sponsored",
			msg: MsgCreateStream{
				Authority:            validAuthority,
				DistributeToRecords:  nil,
				Coins:                validCoins,
				StartTime:            validStartTime,
				DistrEpochIdentifier: validEpochIdentifier,
				NumEpochsPaidOver:    validNumEpochs,
				Sponsored:            true,
			},
			wantErr: false,
		},
		{
			name: "invalid - both distribute to records and sponsored",
			msg: MsgCreateStream{
				Authority:            validAuthority,
				DistributeToRecords:  validDistributeToRecords,
				Coins:                validCoins,
				StartTime:            validStartTime,
				DistrEpochIdentifier: validEpochIdentifier,
				NumEpochsPaidOver:    validNumEpochs,
				Sponsored:            true,
			},
			wantErr: true,
			errMsg:  "either distribute_to_records or sponsored must be defined, but not both",
		},
		{
			name: "invalid - neither distribute to records nor sponsored",
			msg: MsgCreateStream{
				Authority:            validAuthority,
				DistributeToRecords:  nil,
				Coins:                validCoins,
				StartTime:            validStartTime,
				DistrEpochIdentifier: validEpochIdentifier,
				NumEpochsPaidOver:    validNumEpochs,
				Sponsored:            false,
			},
			wantErr: true,
			errMsg:  "either distribute_to_records or sponsored must be defined, but not both",
		},
		{
			name: "invalid authority",
			msg: MsgCreateStream{
				Authority:            "invalid-authority",
				DistributeToRecords:  validDistributeToRecords,
				Coins:                validCoins,
				StartTime:            validStartTime,
				DistrEpochIdentifier: validEpochIdentifier,
				NumEpochsPaidOver:    validNumEpochs,
				Sponsored:            false,
			},
			wantErr: true,
			errMsg:  "authority must be a valid bech32 address",
		},
		{
			name: "empty coins",
			msg: MsgCreateStream{
				Authority:            validAuthority,
				DistributeToRecords:  validDistributeToRecords,
				Coins:                sdk.NewCoins(),
				StartTime:            validStartTime,
				DistrEpochIdentifier: validEpochIdentifier,
				NumEpochsPaidOver:    validNumEpochs,
				Sponsored:            false,
			},
			wantErr: true,
			errMsg:  "coins should not be empty",
		},
		{
			name: "zero start time",
			msg: MsgCreateStream{
				Authority:            validAuthority,
				DistributeToRecords:  validDistributeToRecords,
				Coins:                validCoins,
				StartTime:            time.Time{},
				DistrEpochIdentifier: validEpochIdentifier,
				NumEpochsPaidOver:    validNumEpochs,
				Sponsored:            false,
			},
			wantErr: true,
			errMsg:  "start time should be set",
		},
		{
			name: "empty epoch identifier",
			msg: MsgCreateStream{
				Authority:            validAuthority,
				DistributeToRecords:  validDistributeToRecords,
				Coins:                validCoins,
				StartTime:            validStartTime,
				DistrEpochIdentifier: "",
				NumEpochsPaidOver:    validNumEpochs,
				Sponsored:            false,
			},
			wantErr: true,
			errMsg:  "epoch identifier should be set",
		},
		{
			name: "zero num epochs",
			msg: MsgCreateStream{
				Authority:            validAuthority,
				DistributeToRecords:  validDistributeToRecords,
				Coins:                validCoins,
				StartTime:            validStartTime,
				DistrEpochIdentifier: validEpochIdentifier,
				NumEpochsPaidOver:    0,
				Sponsored:            false,
			},
			wantErr: true,
			errMsg:  "number of epochs paid over should be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgTerminateStream_ValidateBasic(t *testing.T) {
	validAuthority := sample.AccAddress()

	tests := []struct {
		name    string
		msg     MsgTerminateStream
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: MsgTerminateStream{
				Authority: validAuthority,
				StreamId:  1,
			},
			wantErr: false,
		},
		{
			name: "invalid authority",
			msg: MsgTerminateStream{
				Authority: "invalid-authority",
				StreamId:  1,
			},
			wantErr: true,
			errMsg:  "authority must be a valid bech32 address",
		},
		{
			name: "zero stream id",
			msg: MsgTerminateStream{
				Authority: validAuthority,
				StreamId:  0,
			},
			wantErr: true,
			errMsg:  "stream id should be greater than 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgReplaceStream_ValidateBasic(t *testing.T) {
	validAuthority := sample.AccAddress()
	validRecords := []DistrRecord{
		{
			GaugeId: 1,
			Weight:  math.NewInt(100),
		},
	}

	tests := []struct {
		name    string
		msg     MsgReplaceStream
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: MsgReplaceStream{
				Authority: validAuthority,
				StreamId:  1,
				Records:   validRecords,
			},
			wantErr: false,
		},
		{
			name: "invalid authority",
			msg: MsgReplaceStream{
				Authority: "invalid-authority",
				StreamId:  1,
				Records:   validRecords,
			},
			wantErr: true,
			errMsg:  "authority must be a valid bech32 address",
		},
		{
			name: "zero stream id",
			msg: MsgReplaceStream{
				Authority: validAuthority,
				StreamId:  0,
				Records:   validRecords,
			},
			wantErr: true,
			errMsg:  "stream id should be greater than 0",
		},
		{
			name: "empty records",
			msg: MsgReplaceStream{
				Authority: validAuthority,
				StreamId:  1,
				Records:   []DistrRecord{},
			},
			wantErr: true,
			errMsg:  "records should not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgUpdateStream_ValidateBasic(t *testing.T) {
	validAuthority := sample.AccAddress()
	validRecords := []DistrRecord{
		{
			GaugeId: 1,
			Weight:  math.NewInt(100),
		},
	}

	tests := []struct {
		name    string
		msg     MsgUpdateStream
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: MsgUpdateStream{
				Authority: validAuthority,
				StreamId:  1,
				Records:   validRecords,
			},
			wantErr: false,
		},
		{
			name: "invalid authority",
			msg: MsgUpdateStream{
				Authority: "invalid-authority",
				StreamId:  1,
				Records:   validRecords,
			},
			wantErr: true,
			errMsg:  "authority must be a valid bech32 address",
		},
		{
			name: "zero stream id",
			msg: MsgUpdateStream{
				Authority: validAuthority,
				StreamId:  0,
				Records:   validRecords,
			},
			wantErr: true,
			errMsg:  "stream id should be greater than 0",
		},
		{
			name: "empty records",
			msg: MsgUpdateStream{
				Authority: validAuthority,
				StreamId:  1,
				Records:   []DistrRecord{},
			},
			wantErr: true,
			errMsg:  "records should not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}