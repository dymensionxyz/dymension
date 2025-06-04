package types

import (
	context "context"

	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
)

type WarpQuery interface {
	BridgedSupply(ctx context.Context, request *warptypes.QueryBridgedSupplyRequest) (*warptypes.QueryBridgedSupplyResponse, error)
}
