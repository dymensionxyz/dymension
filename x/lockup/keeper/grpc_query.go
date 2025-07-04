package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

var _ types.QueryServer = Querier{}

// Querier defines a wrapper around the x/lockup keeper providing gRPC method
// handlers.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

// ModuleBalance Return full balance of the module.
func (q Querier) ModuleBalance(goCtx context.Context, _ *types.ModuleBalanceRequest) (*types.ModuleBalanceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.ModuleBalanceResponse{Coins: q.GetModuleBalance(ctx)}, nil
}

// ModuleLockedAmount returns locked balance of the module,
// which are all the tokens not unlocking + tokens that are not finished unlocking.
func (q Querier) ModuleLockedAmount(goCtx context.Context, _ *types.ModuleLockedAmountRequest) (*types.ModuleLockedAmountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.ModuleLockedAmountResponse{Coins: q.GetModuleLockedCoins(ctx)}, nil
}

// AccountUnlockableCoins returns unlockable coins which are not withdrawn yet.
func (q Querier) AccountUnlockableCoins(goCtx context.Context, req *types.AccountUnlockableCoinsRequest) (*types.AccountUnlockableCoinsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Owner) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty owner")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}

	return &types.AccountUnlockableCoinsResponse{Coins: q.GetAccountUnlockableCoins(ctx, owner)}, nil
}

// AccountUnlockingCoins returns the total amount of unlocking coins for a specific account.
func (q Querier) AccountUnlockingCoins(goCtx context.Context, req *types.AccountUnlockingCoinsRequest) (*types.AccountUnlockingCoinsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Owner) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty owner")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}

	return &types.AccountUnlockingCoinsResponse{Coins: q.GetAccountUnlockingCoins(ctx, owner)}, nil
}

// AccountLockedCoins returns the total amount of locked coins that can't be withdrawn for a specific account.
func (q Querier) AccountLockedCoins(goCtx context.Context, req *types.AccountLockedCoinsRequest) (*types.AccountLockedCoinsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Owner) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty owner")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}

	return &types.AccountLockedCoinsResponse{Coins: q.GetAccountLockedCoins(ctx, owner)}, nil
}

// AccountLockedPastTime returns the locks of an account whose unlock time is beyond provided timestamp.
func (q Querier) AccountLockedPastTime(goCtx context.Context, req *types.AccountLockedPastTimeRequest) (*types.AccountLockedPastTimeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Owner) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty owner")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}

	return &types.AccountLockedPastTimeResponse{Locks: q.GetAccountLockedPastTime(ctx, owner, req.Timestamp)}, nil
}

// AccountUnlockedBeforeTime returns locks of an account of which unlock time is before the provided timestamp.
func (q Querier) AccountUnlockedBeforeTime(goCtx context.Context, req *types.AccountUnlockedBeforeTimeRequest) (*types.AccountUnlockedBeforeTimeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Owner) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty owner")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}

	return &types.AccountUnlockedBeforeTimeResponse{Locks: q.GetAccountUnlockedBeforeTime(ctx, owner, req.Timestamp)}, nil
}

// AccountLockedPastTimeDenom returns the locks of an account whose unlock time is beyond provided timestamp, limited to locks with
// the specified denom. Equivalent to `AccountLockedPastTime` but denom specific.
func (q Querier) AccountLockedPastTimeDenom(goCtx context.Context, req *types.AccountLockedPastTimeDenomRequest) (*types.AccountLockedPastTimeDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Owner) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty owner")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}

	return &types.AccountLockedPastTimeDenomResponse{Locks: q.GetAccountLockedPastTimeDenom(ctx, owner, req.Denom, req.Timestamp)}, nil
}

// LockedByID returns lock by lock ID.
func (q Querier) LockedByID(goCtx context.Context, req *types.LockedRequest) (*types.LockedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	lock, err := q.GetLockByID(ctx, req.LockId)
	return &types.LockedResponse{Lock: lock}, err
}

// NextLockID returns next lock ID to be created.
func (q Querier) NextLockID(goCtx context.Context, req *types.NextLockIDRequest) (*types.NextLockIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	lastLockID := q.GetLastLockID(ctx)
	nextLockID := lastLockID + 1

	return &types.NextLockIDResponse{LockId: nextLockID}, nil
}

// AccountLockedLongerDuration returns locks of an account with duration longer than specified.
func (q Querier) AccountLockedLongerDuration(goCtx context.Context, req *types.AccountLockedLongerDurationRequest) (*types.AccountLockedLongerDurationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Owner) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty owner")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}

	locks := q.GetAccountLockedLongerDuration(ctx, owner, req.Duration)
	return &types.AccountLockedLongerDurationResponse{Locks: locks}, nil
}

// AccountLockedLongerDurationDenom returns locks of an account with duration longer than specified with specific denom.
func (q Querier) AccountLockedLongerDurationDenom(goCtx context.Context, req *types.AccountLockedLongerDurationDenomRequest) (*types.AccountLockedLongerDurationDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Owner) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty owner")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}

	locks := q.GetAccountLockedLongerDurationDenom(ctx, owner, req.Denom, req.Duration)
	return &types.AccountLockedLongerDurationDenomResponse{Locks: locks}, nil
}

// AccountLockedDuration returns the account locked with the specified duration.
func (q Querier) AccountLockedDuration(goCtx context.Context, req *types.AccountLockedDurationRequest) (*types.AccountLockedDurationResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Owner) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty owner")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}

	locks := q.GetAccountLockedDuration(ctx, owner, req.Duration)
	return &types.AccountLockedDurationResponse{Locks: locks}, nil
}

// AccountLockedPastTimeNotUnlockingOnly returns locks of an account with unlock time beyond
// given timestamp excluding locks that has started unlocking.
func (q Querier) AccountLockedPastTimeNotUnlockingOnly(goCtx context.Context, req *types.AccountLockedPastTimeNotUnlockingOnlyRequest) (*types.AccountLockedPastTimeNotUnlockingOnlyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Owner) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty owner")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}

	return &types.AccountLockedPastTimeNotUnlockingOnlyResponse{Locks: q.GetAccountLockedPastTimeNotUnlockingOnly(ctx, owner, req.Timestamp)}, nil
}

// AccountLockedLongerDurationNotUnlockingOnly returns locks of an account with longer duration
// than the specified duration, excluding tokens that has started unlocking.
func (q Querier) AccountLockedLongerDurationNotUnlockingOnly(goCtx context.Context, req *types.AccountLockedLongerDurationNotUnlockingOnlyRequest) (*types.AccountLockedLongerDurationNotUnlockingOnlyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Owner) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty owner")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromBech32(req.Owner)
	if err != nil {
		return nil, err
	}

	return &types.AccountLockedLongerDurationNotUnlockingOnlyResponse{Locks: q.GetAccountLockedLongerDurationNotUnlockingOnly(ctx, owner, req.Duration)}, nil
}

// LockedDenom returns the total amount of denom locked throughout all locks.
func (q Querier) LockedDenom(goCtx context.Context, req *types.LockedDenomRequest) (*types.LockedDenomResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if len(req.Denom) == 0 {
		return nil, errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "empty denom")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.LockedDenomResponse{Amount: q.GetLockedDenom(ctx, req.Denom, req.Duration)}, nil
}

// Params returns module params
func (q Querier) Params(goCtx context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	return &types.QueryParamsResponse{Params: q.GetParams(ctx)}, nil
}
