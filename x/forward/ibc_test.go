package forward

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"
)

type capturingTransferKeeper struct {
	msg *ibctransfertypes.MsgTransfer
}

func (k *capturingTransferKeeper) Transfer(_ context.Context, msg *ibctransfertypes.MsgTransfer) (*ibctransfertypes.MsgTransferResponse, error) {
	k.msg = msg
	return &ibctransfertypes.MsgTransferResponse{}, nil
}

func TestForwardToIBCClampsTimeoutTimestamp(t *testing.T) {
	blockTime := time.Date(2026, time.August, 22, 12, 0, 0, 0, time.UTC)
	blockTimeNs := uint64(blockTime.UnixNano())
	healthyTimeout := uint64(blockTime.Add(2 * time.Hour).UnixNano())

	testCases := []struct {
		name             string
		timeoutTimestamp uint64
		expectedTimeout  uint64
	}{
		{
			name:             "past timeout",
			timeoutTimestamp: uint64(blockTime.Add(-time.Minute).UnixNano()),
			expectedTimeout:  uint64(blockTime.Add(time.Hour).UnixNano()),
		},
		{
			name:             "zero timeout",
			timeoutTimestamp: 0,
			expectedTimeout:  uint64(blockTime.Add(time.Hour).UnixNano()),
		},
		{
			name:             "healthy timeout",
			timeoutTimestamp: healthyTimeout,
			expectedTimeout:  healthyTimeout,
		},
		{
			name:             "minimum lead time boundary",
			timeoutTimestamp: blockTimeNs + uint64(10*time.Minute),
			expectedTimeout:  uint64(blockTime.Add(time.Hour).UnixNano()),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			transferKeeper := &capturingTransferKeeper{}
			forward := New(transferKeeper, nil, nil)
			transfer := &ibctransfertypes.MsgTransfer{
				SourcePort:       "transfer",
				SourceChannel:    "channel-0",
				Receiver:         "destination",
				TimeoutTimestamp: tc.timeoutTimestamp,
				Memo:             "memo",
			}
			fundsSource := sdk.AccAddress([]byte("funds-source-address"))
			budget := sdk.NewCoin("adym", sdkmath.NewInt(10))

			err := forward.forwardToIBC(sdk.Context{}.WithBlockTime(blockTime), transfer, fundsSource, budget)

			require.NoError(t, err)
			require.NotNil(t, transferKeeper.msg)
			require.Equal(t, tc.expectedTimeout, transferKeeper.msg.TimeoutTimestamp)
		})
	}
}
