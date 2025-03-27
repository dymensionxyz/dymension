package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	warpkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/warp/keeper"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
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
	err = k.escrowFromUser(ctx, fundsSource, order.Price)
	if err != nil {
		// should never happen
		err = errorsmod.Wrap(err, "escrow from user")
		k.Logger(ctx).Error("doForwardHook", "error", err)
		return err
	}
	return k.forwardToHyperlane(ctx, order, fundsSource, d)
}

// for transfers coming from eibc which are being forwarded (to HL)
func (k Keeper) forwardToHyperlane(ctx sdk.Context, order *eibctypes.DemandOrder, fundsSource sdk.AccAddress, d types.HookEIBCtoHL) error {

	// TODO: anything to change?
	m := &warptypes.MsgRemoteTransfer{
		Sender:             d.HyperlaneTransfer.Sender,
		TokenId:            d.HyperlaneTransfer.TokenId,
		DestinationDomain:  d.HyperlaneTransfer.DestinationDomain,
		Recipient:          d.HyperlaneTransfer.Recipient,
		Amount:             d.HyperlaneTransfer.Amount,
		CustomHookId:       d.HyperlaneTransfer.CustomHookId,
		GasLimit:           d.HyperlaneTransfer.GasLimit,
		MaxFee:             d.HyperlaneTransfer.MaxFee,
		CustomHookMetadata: d.HyperlaneTransfer.CustomHookMetadata,
	}

	res, err := k.warpServer.RemoteTransfer(ctx, m)
	if err != nil {
		return errorsmod.Wrap(err, "remote transfer")
	}
	_ = res

	return nil

}

/*
TODO's when I get back:
	- Can write the refunds to the recovery address
		HL -> EIBC, tokens are in module account
		EIBC -> Hyperlane, tokens are in user account
	- Should test the happy path e2e
	- Need to also reanalize the ibc test I wrote to see if I can include finalization or something
*/

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
	// TODO: should allow another level of indirection (e.g. Memo is json containing what we want in bytes?)
	// it would be more flexible and allow memo forwarding
	var d types.HookHLtoIBC
	err := proto.Unmarshal(args.Memo, &d)
	if err != nil {
		return errorsmod.Wrap(err, "unmarshal forward hook")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	d.Transfer.Token = args.Coins[0]

	err = k.escrowFromModule(ctx, warptypes.ModuleName, args.Coins)
	if err != nil {
		return errorsmod.Wrap(err, "escrow from module")
	}

	return k.transferTokensHyperlaneToIBC(ctx, d.Transfer)
}

func (k Keeper) transferTokensHyperlaneToIBC(ctx sdk.Context, transfer *ibctransfertypes.MsgTransfer) error {

	var (
		token            = transfer.Token
		sender           = transfer.Sender
		recipient        = transfer.Receiver
		timeoutTimestamp = transfer.TimeoutTimestamp
		memo             string
	)
	ibctransfertypes.NewMsgTransfer(
		"transfer",
		"channel-0",
		token,
		sender,
		recipient,
		ibcclienttypes.Height{}, // ignore, removed in ibc v2 also
		timeoutTimestamp,
		memo,
	)
	res, err := k.transferKeeper.Transfer(ctx, transfer)
	if err != nil {
		return errorsmod.Wrap(err, "transfer")
	}
	_ = res
	// TODO:
	return nil
}
