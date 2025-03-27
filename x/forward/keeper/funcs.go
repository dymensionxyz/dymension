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

// at this point funds have not been sent from the fulfiller/eibc LP/funds provider to the recipient (or anywhere else)
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
	(assume for a moment no escrow within the forwarding module)

EIBC -> HL
	1. Token arrives in delayed ack, order is created
	2. Order is fulfilled, hook is called with funds provider and the other args, tokens are not yet moved
	3. Assume the memo contains the right details for the transfer to hyperlane
	4. Directly transfer funds from fulfiller to hyperlane
	Q: Hyperlane refund
		- There is no such thing as a refund on HL so there is no case to refund
	Q: What happens on finalization if EIBC was fulfilled?
		- Transfer is addressed to the eibc fulfiller
		- EIBC fulfiller will get the IBC tokens
	    TODO, need to test this
	Q: What happens if the transfer fails somehow?
	  1. Error returned to eibc fulfill operation
	  2. Fulfill will fail :(
	  	- Need to work on this. Best thing to do is to reduce chance it can fail on frontend.
		- At the end it can finalize, need to see what happens then. (See below)
		- Can optionally, directly send back to the original rollapp person, but not for MVP
	Q: What happens on finalization if EIBC was not fulfilled?
		TODO

	Investigations:
		- Will the transfer actually be accepted in the first place?
			- Is there somewhere in the code that checks that the transfer will be accepted by the rest of the IBC stack before accepting?
		    - Is this the same behaviour for recv, ack, timeout?
		- What should the IBC transfer recipient be? Can it be some nil value? Will it work? What if we just completely ignore it?
HL -> EIBC
	1. Token arrives on warp module
	2. Hook is called with the coins
	3. Unpack the memo and get the rol target address, this can just be baked into into the embedded IBC transfer object
	4. Send the IBC transfer
	Q: IBC transfer initiation fails
		- No such thing as a refund on HL, so would have to do the reverse transfer on HL
		- Easiest thing is to force specifying some refund address on the hub...
	Q: IBC transfer timeout
		- Sender would be the HL module
		- It should get a refund from ibc/eibc
	    - Then we are back in the transfer initiation failure case
	Q: IBC transfer error ack
		- Not for MVP, we should just make sure it's rare

	Investigations:
		- Can we just make the recipient some ignored value?
		- Does IBC transfer succ acc matter?


Extension:
	EIBC -> IBC
	HL ->IBC


Conclusions:
	Moving forward with the recovery address idea
High level investigations:
	Need to check that osmosis allows you to specify a memo on the ibc transfer
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

	return k.transferTokensHyperlaneToIBC(ctx, d.Transfer)
}

func (k Keeper) transferTokensHyperlaneToIBC(ctx sdk.Context, transfer *ibctransfertypes.MsgTransfer) error {
	return nil

}

// func (k Keeper) transferTokensHyperlaneToIBC(ctx sdk.Context, transfer *ibctransfertypes.MsgTransfer) error {
// 	var (
// 		token         = transfer.Token
// 		amount        = transfer.Amount
// 		sender        = transfer.Sender
// 		recipient     = transfer.Receiver
// 		timeoutHeight = transfer.TimeoutHeight
// 	)
// 	k.transferKeeper.Transfer(
// 		ctx,
// 		&ibctransfertypes.MsgTransfer{
// 			SourcePort:       "transfer",
// 			SourceChannel:    "channel-0",
// 			Token:            token,
// 			Sender:           sender,
// 			Receiver:         recipient,
// 			TimeoutHeight:    &types.Height{},
// 			TimeoutTimestamp: 0,
// 		},
// 	)
// }
