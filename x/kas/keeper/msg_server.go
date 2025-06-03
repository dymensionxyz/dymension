package keeper

import (
	"context"
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/kas/types"

	hypercoretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	errorsmod "cosmossdk.io/errors"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
)

type msgServer struct {
	*Keeper
}

func (m msgServer) Foo(context.Context, *types.MsgFoo) (*types.MsgFooResponse, error) {
	panic("unimplemented")
}

func (k *Keeper) IndicateProgress(goCtx context.Context, req *types.MsgIndicateProgress) (*types.MsgIndicateProgressResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	metadata, err := hypercoretypes.NewMessageIdMultisigRawMetadata(req.Metadata)
	if err != nil {
		return nil, errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "metadata")
	}

	payload := req.GetPayload()

	if payload == nil {
		return nil, errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "payload")
	}

	digest, err := req.GetPayload().SignBytes()
	if err != nil {
		return nil, errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "payload digest")
	}

	threshold, vals := k.MustValidators(ctx)

	ok, err := hypercoretypes.VerifyMultisig(vals, threshold, metadata.Signatures, digest)
	if err != nil {
		return nil, errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "verify multisig")
	}
	if !ok {
		return nil, errorsmod.Wrap(errors.Join(gerrc.ErrUnauthenticated, err), "verify multisig")
	}

	return &types.MsgIndicateProgressResponse{}, nil
}

/* TODO: we will need a way to seed the hub state with the appropriate stuff
- the ISM and mailbox for the kaspa warp route
*/

// returns threshold and validator set
func (k *Keeper) MustValidators(ctx sdk.Context) (uint32, []string) {
	var ismID hyperutil.HexAddress
	ism, err := k.hypercoreK.IsmKeeper.Get(ctx, ismID)
	if err != nil {
		panic(err)
	}
	conc, ok := ism.(*hypercoretypes.MessageIdMultisigISMRaw)
	if !ok {
		panic("ism is not a MessageIdMultisigISMRaw")
	}
	return conc.GetThreshold(), conc.GetValidators()
}

func NewMsgServerImpl(keeper *Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}
