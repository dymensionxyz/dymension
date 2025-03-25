package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	eibckeeper "github.com/dymensionxyz/dymension/v3/x/eibc/keeper"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	types "github.com/dymensionxyz/dymension/v3/x/forward/types"
)

const (
	HookNameForward = "forward"
)

var _ eibckeeper.FulfillHook = Hook{}

func (k Keeper) Hook() Hook {
	return Hook{
		Keeper: &k,
	}
}

type Hook struct {
	*Keeper
}

func (h Hook) ValidateData(data []byte) error {
	return validForward(data)
}

func (h Hook) Run(ctx sdk.Context, order *eibctypes.DemandOrder, fundsSource sdk.AccAddress,
	newTransferRecipient sdk.AccAddress,
	fulfiller sdk.AccAddress, hookData []byte) error {
	return h.doForwardHook(ctx, order, fundsSource, hookData)
}

func (h Hook) Name() string {
	return HookNameForward
}

func validForward(data []byte) error {
	var d types.HookCalldata
	err := proto.Unmarshal(data, &d)
	if err != nil {
		return errorsmod.Wrap(err, "unmarshal forward hook")
	}
	if err := d.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "validate forward hook")
	}
	return nil
}

func (k Keeper) doForwardHook(ctx sdk.Context, order *eibctypes.DemandOrder, fundsSource sdk.AccAddress, data []byte) error {
	var d types.HookCalldata
	err := proto.Unmarshal(order.FulfillHook.HookData, &d)
	if err != nil {
		return errorsmod.Wrap(err, "unmarshal forward hook")
	}
	return k.forwardToHyperlane(ctx, order, fundsSource, d)
}

func (k Keeper) forwardToHyperlane(ctx sdk.Context, order *eibctypes.DemandOrder, fundsSource sdk.AccAddress, d types.HookCalldata) error {

	// m := warptypes.MsgRemoteTransfer{
	// 	TokenId:            tokenId,
	// 	DestinationDomain:  uint32(destinationDomain),
	// 	Sender:             clientCtx.GetFromAddress().String(),
	// 	Recipient:          recipient,
	// 	Amount:             argAmount,
	// 	CustomHookId:       parsedHookId,
	// 	GasLimit:           gasLimitInt,
	// 	MaxFee:             maxFeeCoin,
	// 	CustomHookMetadata: customHookMetadata,
	// }

	res, err := k.warpServer.RemoteTransfer(ctx, d.HyperlaneTransfer)
	if err != nil {
		return errorsmod.Wrap(err, "remote transfer")
	}
	_ = res

	// var token warptypes.HypToken
	// var dst uint32
	// var recipient util.HexAddress
	// var amount math.Int
	// var customHookId *util.HexAddress
	// var gasLimit math.Int
	// var maxFee sdk.Coin
	// var customHookMetadata []byte

	// k.warpKeeper.RemoteTransferCollateral(ctx,
	// 	token,
	// 	fundsSource.String(),
	// 	dst,
	// 	recipient,
	// 	amount,
	// 	customHookId,
	// 	gasLimit,
	// 	maxFee,
	// 	customHookMetadata,
	// )

	return nil

}
