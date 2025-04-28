package keeper // have to call it keeper

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	porttypes "github.com/cosmos/ibc-go/v8/modules/core/05-port/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/osmosis-labs/osmosis/v15/osmoutils"
)

type CompletionHookInstance interface {
	ValidateArg(hookData []byte) error
	Run(ctx sdk.Context, fundSrc sdk.AccAddress, budget sdk.Coin, hookData []byte) error
}

// map name -> instance
func (k Keeper) SetCompletionHooks(hooks map[string]CompletionHookInstance) {
	for name, hook := range hooks {
		k.completionHooks[name] = hook
	}
}

// assumes already passed validate basic
func (k Keeper) ValidateCompletionHook(info commontypes.CompletionHookCall) error {
	f, ok := k.completionHooks[info.Name]
	if !ok {
		return gerrc.ErrNotFound.Wrapf("hook: name: %s", info.Name)
	}
	return f.ValidateArg(info.Data)
}

func (k Keeper) RunOrderCompletionHook(ctx sdk.Context, o *commontypes.DemandOrder, amt math.Int) error {
	fundsSrc := o.GetRecipientBech32Address()
	budget := sdk.NewCoin(o.Denom(), amt)
	return k.RunCompletionHook(ctx, fundsSrc, budget, *o.CompletionHook)
}

func (k Keeper) RunCompletionHook(ctx sdk.Context, fundsSrc sdk.AccAddress, budget sdk.Coin, call commontypes.CompletionHookCall) error {
	f, ok := k.completionHooks[call.Name]
	if !ok {
		return gerrc.ErrInternal.Wrapf("completion hook not registered, should have been checked already: %s", call.Name)
	}
	return f.Run(ctx, fundsSrc, budget, call.Data)
}

// Should be called after packet finalization
// Recipient can either be the fulfiller of a hook that already occurred, or the original recipient still, who probably still wants the hook to happen
// NOTE: there is an asymmetry currently because on fulfill supports multiple hooks, but this finalization onRecv is hardcoded for x/forward atm
func (k Keeper) finalizeOnRecv(ctx sdk.Context, ibc porttypes.IBCModule, p *commontypes.RollappPacket) error { // TODO: rename func
	// Because we intercepted the packet, the core ibc library wasn't able to write the ack when we first
	// got the packet. So we try to write it here.

	ack := ibc.OnRecvPacket(ctx, *p.Packet, p.Relayer)
	/*
			We only write the ack if writing it succeeds:
			1. Transfer fails and writing ack fails - In this case, the funds will never be refunded on the RA.
					non-eibc: sender will never get the funds back
					eibc:     the fulfiller will never get the funds back, the original target has already been paid
			2. Transfer succeeds and writing ack fails - In this case, the packet is never cleared on the RA.
			3. Transfer succeeds and writing succeeds - happy path
			4. Transfer fails and ack succeeds - we write the err ack and the funds will be refunded on the RA
					non-eibc: sender will get the funds back
		            eibc:     effective transfer from fulfiller to original target
	*/
	if ack != nil { // NOTE: in practice ack should not be nil, since ibc transfer core module always returns something
		err := osmoutils.ApplyFuncIfNoError(ctx, k.writeRecvAck(*p, ack))
		if err != nil {
			return err
		}
	}

	/*

		*In general* we want a way to do something whenever an ibc transfer happens ("Hook"). It can happen
			1. on EIBC fulfill
			2. on finalize to the original recipient, for non fulfilled orders
			3. on finalize to the fulfiller, for fulfilled orders

		1. Do the hook on EIBC fulfillment, using immediate funds
		2. On finalize, look up the EIBC demand order to check if it's fulfilled or not.
			a. If it ISN'T, then do the hook AFTER the ibc transfer stack finishes
			b. If it IS, then do nothing

		We can do (2) by finding the eibc order directly using the packet key, because the status has not yet been update to finalized
	*/

	o, err := k.EIBCKeeper.PendingOrderByPacket(ctx, p)
	if errorsmod.IsOf(err, eibctypes.ErrDemandOrderDoesNotExist) {
		// not much we can do here, it should exist...
		k.Logger(ctx).Error("Pending order by packet not found.", "packet", p.LogString())
		return nil
	}
	if err != nil {
		return errorsmod.Wrap(err, "pending order by packet")
	}

	// !! for mvp, we assume all completion hooks are only executable once !!
	if o.IsFulfilled() {
		// done
		return nil
	}

	if o.CompletionHook == nil {
		// done
		return nil
	}

	port, channel := commontypes.PacketHubPortChan(commontypes.RollappPacket_ON_RECV, *p.Packet)
	pTransfer, err := k.rollappKeeper.GetValidTransfer(ctx, p.Packet.Data, port, channel)
	if err != nil {
		return errorsmod.Wrap(err, "get valid transfer")
	}
	amt := pTransfer.MustAmountInt()
	// account for the bridge fee which happened before the receiver got the funds
	amt = amt.Sub(k.BridgingFeeFromAmt(ctx, amt))
	return k.RunOrderCompletionHook(ctx, o, amt)
}
