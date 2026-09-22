package forward

import (
	"context"
	"strconv"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	warpkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/warp/keeper"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
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

			err := forward.forwardToIBC(sdk.Context{}.WithBlockTime(blockTime), transfer, fundsSource, budget, sdkmath.Int{})

			require.NoError(t, err)
			require.NotNil(t, transferKeeper.msg)
			require.Equal(t, tc.expectedTimeout, transferKeeper.msg.TimeoutTimestamp)
		})
	}
}

func TestIBCForwardMinimum(t *testing.T) {
	for _, path := range []string{"direct", "rollapp completion", "hyperlane inbound"} {
		for _, tc := range []struct {
			name   string
			floor  sdkmath.Int
			reject bool
		}{
			{"below floor", sdkmath.NewInt(91), true},
			{"at floor", sdkmath.NewInt(90), false},
			{"above floor", sdkmath.NewInt(89), false},
			{"zero", sdkmath.ZeroInt(), false},
			{"absent", sdkmath.Int{}, false},
		} {
			t.Run(path+"/"+tc.name, func(t *testing.T) {
				key := storetypes.NewKVStoreKey("forward")
				ctx := testutil.DefaultContext(key, storetypes.NewTransientStoreKey("transient")).WithBlockTime(time.Now())
				keeper := &capturingTransferKeeper{}
				f := New(keeper, nil, nil)
				sender := sdk.AccAddress(make([]byte, 20))
				budget := sdk.NewInt64Coin("adym", 90)
				hook := types.NewHookForwardToIBC("channel-0", "destination", 1, tc.floor)
				var err error
				switch path {
				case "direct":
					err = f.forwardToIBC(ctx, hook.Transfer, sender, budget, tc.floor)
					if tc.reject {
						require.ErrorContains(t, err, "forwardable budget 90 below min_amount 91")
					} else {
						require.NoError(t, err)
					}
				default:
					bz, marshalErr := proto.Marshal(hook)
					require.NoError(t, marshalErr)
					if tc.name == "absent" {
						// Encode only field 1 to reproduce a pre-floor composer's payload.
						transferBz, e := proto.Marshal(hook.Transfer)
						require.NoError(t, e)
						require.Less(t, len(transferBz), 128)
						bz = append([]byte{0x0a, byte(len(transferBz))}, transferBz...)
					}
					if path == "rollapp completion" {
						err = f.RollToIBCHook().Run(ctx, sender, budget, bz)
					} else {
						metadata, e := proto.Marshal(&types.HLMetadata{HookForwardToIbc: bz})
						require.NoError(t, e)
						err = f.OnHyperlaneMessage(ctx, warpkeeper.OnHyperlaneMessageArgs{Account: sender, Coins: sdk.NewCoins(budget), Metadata: metadata})
					}
					require.NoError(t, err)
					events := ctx.EventManager().Events()
					require.Len(t, events, 1)
					require.Equal(t, proto.MessageName(&types.EventForward{}), events[0].Type)
					attrs := map[string]string{}
					for _, attr := range events[0].Attributes {
						attrs[attr.Key] = attr.Value
					}
					require.Equal(t, "true", attrs["was_forwarded"])
					require.Equal(t, strconv.FormatBool(!tc.reject), attrs["ok"], attrs["err"])
					if tc.reject {
						require.Contains(t, attrs["err"], "forwardable budget 90 below min_amount 91")
					}
				}
				if tc.reject {
					require.Nil(t, keeper.msg)
				} else {
					require.NotNil(t, keeper.msg)
					require.Equal(t, budget, keeper.msg.Token)
					require.Equal(t, sender.String(), keeper.msg.Sender)
				}
			})
		}
	}
}
