package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/x/sequencer/types"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// CreateSequencer defines a method for creating a new sequencer
func (k msgServer) CreateSequencer(goCtx context.Context, msg *types.MsgCreateSequencer) (*types.MsgCreateSequencerResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return nil, err
	}

	seqAddr, err := sdk.AccAddressFromBech32(msg.SequencerAddress)
	if err != nil {
		return nil, err
	}

	seqAddrStr := seqAddr.String()

	// check to see if the pubkey or sender has been registered before
	if _, found := k.GetSequencer(ctx, seqAddrStr); found {
		return nil, types.ErrSequencerExists
	}

	pk, ok := msg.Pubkey.GetCachedValue().(cryptotypes.PubKey)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "Expecting cryptotypes.PubKey, got %T", pk)
	}

	pkAny, err := codectypes.NewAnyWithValue(pk)
	if err != nil {
		return nil, err
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, err
	}

	sequencer := types.Sequencer{
		Creator:          creator.String(),
		SequencerAddress: seqAddrStr,
		Pubkey:           pkAny,
		Description:      msg.Description,
		RollappId:        msg.RollappId,
	}

	k.SetSequencer(ctx, sequencer)

	return &types.MsgCreateSequencerResponse{}, nil
}
