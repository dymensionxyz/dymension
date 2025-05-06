package types

import (
	context "context"

	"github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
)

type WarpQuery interface {
	Token(ctx context.Context, request *types.QueryTokenRequest) (*types.QueryTokenResponse, error)
}

type TransferKeeper interface {
	Transfer(ctx context.Context, msg *ibctransfertypes.MsgTransfer) (*ibctransfertypes.MsgTransferResponse, error)
}

type WarpMsgServer interface {
	DymRemoteTransfer(ctx context.Context, msg *types.MsgDymRemoteTransfer) (*types.MsgDymRemoteTransferResponse, error)
}
