package keeper

import (
	"context"
	"fmt"
	"strings"

	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ types.QueryServer = Querier{}

type Querier struct {
	Keeper
}

// NewQuerier creates a new Querier struct.
func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

func (q Querier) IBCDenomByDenomTrace(goCtx context.Context, req *types.QueryGetIBCDenomByDenomTraceRequest) (*types.QueryIBCDenomByDenomTraceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	traces := strings.Split(req.DenomTrace, "/")
	if len(traces) < 3 {
		return nil, status.Error(codes.InvalidArgument, "input denom traces invalid, need to have at least 3 elements")
	}

	if len(traces)%2 == 0 {
		return nil, status.Error(codes.InvalidArgument, "denom traces must follow this format [port-id-1]/[channel-id-1]/.../[port-id-n]/[channel-id-n]/[denom]")
	}

	denom := traces[len(traces)-1]

	for i := 0; i < len(traces)-1; i += 2 {
		portID := traces[i]
		channelID := traces[i+1]
		if !strings.Contains(channelID, "channel-") {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("channel %s must contain channel-", channelID))
		}
		tokenDenom := transfertypes.GetPrefixedDenom(portID, channelID, denom)
		denom = transfertypes.ParseDenomTrace(tokenDenom).IBCDenom()
	}

	ibcDenomResponse := &types.QueryIBCDenomByDenomTraceResponse{IbcDenom: denom}
	return ibcDenomResponse, nil
}
