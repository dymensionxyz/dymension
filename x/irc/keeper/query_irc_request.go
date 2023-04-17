package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/dymensionxyz/dymension/x/irc/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) IRCRequestAll(goCtx context.Context, req *types.QueryAllIRCRequestRequest) (*types.QueryAllIRCRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var ircRequests []types.IRCRequest
	ctx := sdk.UnwrapSDKContext(goCtx)

	store := ctx.KVStore(k.storeKey)
	ircRequestStore := prefix.NewStore(store, types.KeyPrefix(types.IRCRequestKeyPrefix))

	pageRes, err := query.Paginate(ircRequestStore, req.Pagination, func(key []byte, value []byte) error {
		var ircRequest types.IRCRequest
		if err := k.cdc.Unmarshal(value, &ircRequest); err != nil {
			return err
		}

		ircRequests = append(ircRequests, ircRequest)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllIRCRequestResponse{IrcRequest: ircRequests, Pagination: pageRes}, nil
}

func (k Keeper) IRCRequest(goCtx context.Context, req *types.QueryGetIRCRequestRequest) (*types.QueryGetIRCRequestResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	val, found := k.GetIRCRequest(
		ctx,
		req.ReqId,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetIRCRequestResponse{IrcRequest: val}, nil
}
