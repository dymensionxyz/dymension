package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	warpkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/warp/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
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

func validForward(data []byte) error {
	var d types.HookEIBCtoHL
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
	var d types.HookEIBCtoHL
	err := proto.Unmarshal(order.FulfillHook.HookData, &d)
	if err != nil {
		return errorsmod.Wrap(err, "unmarshal forward hook")
	}
	return k.forwardToHyperlane(ctx, order, fundsSource, d)
}

// for transfers coming from eibc which are being forwarded (to HL)
func (k Keeper) forwardToHyperlane(ctx sdk.Context, order *eibctypes.DemandOrder, fundsSource sdk.AccAddress, d types.HookEIBCtoHL) error {

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

/*
Does it make sense to store funds directly in module?
- It would be closer to PFM design and make refunds easier?

What are all the flows

EIBC -> HL

HL -> EIBC

Extension:
	EIBC -> IBC
	HL ->IBC

*/

// for inbound warp route transfers. At this point, the tokens are in the hyperlane warp module still
func (k Keeper) Handle(goCtx context.Context, args warpkeeper.DymHookArgs) error {
	var d types.HookHLtoIBC
	err := proto.Unmarshal(args.Memo, &d)
	if err != nil {
		return errorsmod.Wrap(err, "unmarshal forward hook")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	d.Transfer.Token = args.Coins[0]

	warpAcc
	d.Transfer.Sender
	return k.transferTokensHyperlaneToIBC(ctx, d.Transfer)
}

func (k Keeper) transferTokensHyperlaneToIBC(ctx sdk.Context, transfer *ibctransfertypes.MsgTransfer) error {
	k.transferKeeper.Transfer(
		ctx,
		&ibctransfertypes.MsgTransfer{
			SourcePort:       "transfer",
			SourceChannel:    "channel-0",
			Token:            sdk.NewCoin(args.Token, args.Amount),
			Sender:           args.Sender,
			Receiver:         args.Recipient,
			TimeoutHeight:    &types.Height{},
			TimeoutTimestamp: 0,
		},
	)
}
