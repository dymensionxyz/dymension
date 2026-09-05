package forward

import (
	"context"
	"strconv"
	"testing"

	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warpkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/warp/keeper"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
	"github.com/stretchr/testify/require"
)

type forwardWarpStub struct {
	sent []*warptypes.MsgRemoteTransfer
}

func (s *forwardWarpStub) Token(context.Context, *warptypes.QueryTokenRequest) (*warptypes.QueryTokenResponse, error) {
	return &warptypes.QueryTokenResponse{Token: &warptypes.WrappedHypToken{OriginDenom: "adym"}}, nil
}

func (s *forwardWarpStub) RemoteTransfer(_ context.Context, msg *warptypes.MsgRemoteTransfer) (*warptypes.MsgRemoteTransferResponse, error) {
	s.sent = append(s.sent, msg)
	return &warptypes.MsgRemoteTransferResponse{}, nil
}

func TestHyperlaneForwardEntryPaths(t *testing.T) {
	for _, path := range []string{"rollapp completion", "hyperlane inbound"} {
		for _, tc := range []struct {
			name                string
			full                bool
			budget, floor, want int64
		}{
			{"post eibc fee", true, 90, 80, 80},
			{"surplus", true, 120, 0, 110},
			{"floor violation", true, 90, 81, 0},
			{"fee exhausts budget", true, 10, 0, 0},
			{"legacy insufficient", false, 90, 0, 0},
			{"legacy unchanged", false, 120, 0, 100},
		} {
			t.Run(path+"/"+tc.name, func(t *testing.T) {
				ctx := testutil.DefaultContext(storetypes.NewKVStoreKey("forward"), storetypes.NewTransientStoreKey("transient"))
				stub := &forwardWarpStub{}
				f := New(nil, stub, stub)
				sender := sdk.AccAddress(make([]byte, 20))
				hook := types.NewHookForwardToHL(hyperutil.HexAddress{}, 1, hyperutil.HexAddress{}, math.NewInt(100), sdk.NewInt64Coin("adym", 10), math.ZeroInt(), nil, "", tc.full, math.NewInt(tc.floor))
				budget := sdk.NewInt64Coin("adym", tc.budget)
				var err error
				if path == "rollapp completion" {
					bz, marshalErr := proto.Marshal(hook)
					require.NoError(t, marshalErr)
					err = f.RollToHLHook().Run(ctx, sender, budget, bz)
				} else {
					bz, marshalErr := types.MakeHLForwardToHLMetadata(hook)
					require.NoError(t, marshalErr)
					err = f.OnHyperlaneMessage(ctx, warpkeeper.OnHyperlaneMessageArgs{Account: sender, Coins: sdk.NewCoins(budget), Metadata: bz})
				}
				require.NoError(t, err) // Failures are reported through EventForward.
				events := ctx.EventManager().Events()
				require.Len(t, events, 1)
				require.Equal(t, proto.MessageName(&types.EventForward{}), events[0].Type)
				attrs := make(map[string]string)
				for _, attr := range events[0].Attributes {
					attrs[attr.Key] = attr.Value
				}
				require.Equal(t, "true", attrs["was_forwarded"])
				require.Equal(t, strconv.FormatBool(tc.want > 0), attrs["ok"], attrs["err"])
				if tc.want == 0 {
					require.Empty(t, stub.sent)
					require.NotEmpty(t, attrs["err"])
					return
				}
				require.Len(t, stub.sent, 1)
				sent := stub.sent[0]
				require.Equal(t, math.NewInt(tc.want), sent.Amount)
				require.Equal(t, sender.String(), sent.Sender)
				require.Equal(t, hook.HyperlaneTransfer.MaxFee, sent.MaxFee)
				if tc.full {
					require.Equal(t, budget.Amount, sent.Amount.Add(sent.MaxFee.Amount))
				} else {
					legacy := *hook.HyperlaneTransfer
					legacy.Sender = sender.String()
					before, err := proto.Marshal(&legacy)
					require.NoError(t, err)
					after, err := proto.Marshal(sent)
					require.NoError(t, err)
					require.Equal(t, before, after)
				}
			})
		}
	}
}
