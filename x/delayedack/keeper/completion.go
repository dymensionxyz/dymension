package keeper // have to call it keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

type CompletionHookInstance interface {
	ValidateArg(hookData []byte) error
	Run(ctx sdk.Context, fundSrc sdk.AccAddress, budget sdk.Coin, hookData []byte) error
}

// map name -> instance
func (h Keeper) SetCompletionHooks(hooks map[string]CompletionHookInstance) {
	for name, hook := range hooks {
		h.completionHooks[name] = hook
	}
}

// assumes already passed validate basic
func (h Keeper) Validate(info commontypes.CompletionHookCall) error {
	f, ok := h.completionHooks[info.Name]
	if !ok {
		return gerrc.ErrNotFound.Wrapf("hook: name: %s", info.Name)
	}
	return f.ValidateArg(info.Data)
}

// Should be called after packet finalization
// Recipient can either be the fulfiller of a hook that already occurred, or the original recipient still, who probably still wants the hook to happen
// NOTE: there is an asymmetry currently because on fulfill supports multiple hooks, but this finalization onRecv is hardcoded for x/forward atm
func (h Keeper) OnRecvPacket(ctx sdk.Context, p *commontypes.RollappPacket) error { // TODO: rename func

	o, err := h.EIBCKeeper.PendingOrderByPacket(ctx, p)
	if errorsmod.IsOf(err, eibctypes.ErrDemandOrderDoesNotExist) {
		// not much we can do here, it should exist...
		h.Logger(ctx).Error("Pending order by packet not found.", "packet", p.LogString())
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

	f, ok := h.completionHooks[o.CompletionHook.Name]
	if !ok {
		return gerrc.ErrNotFound.Wrap("hook")
	}

	// the order wasn't fulfilled, so the funds were sent to the original recipient by ibc transfer app
	budget := sdk.NewCoin(o.Denom(), o.PriceAmount()) // TODO: fix amount, need to account for fees and so on
	return f.Run(ctx, o.GetRecipientBech32Address(), budget, o.CompletionHook.Data)

}
