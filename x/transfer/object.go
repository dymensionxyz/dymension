package keeper // have to call it keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/dymensionxyz/dymension/v3/x/transfer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

type EIBCK interface {
	PendingOrderByPacket(ctx sdk.Context, p *commontypes.RollappPacket) (*eibctypes.DemandOrder, error)
}

type TransferHooks struct {
	Keeper EIBCK
	hooks  map[string]CompletionHookInstance
}

func (h *TransferHooks) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
func NewTransferHooks(keeper EIBCK) *TransferHooks {

	return &TransferHooks{
		Keeper: keeper,
		hooks:  make(map[string]CompletionHookInstance),
	}
}

type CompletionHookInstance interface {
	ValidateData(hookData []byte) error
	Run(ctx sdk.Context, fundSrc sdk.AccAddress, budget sdk.Coin, hookData []byte) error
}

// map name -> instance
func (h TransferHooks) SetHooks(hooks map[string]CompletionHookInstance) {
	for name, hook := range hooks {
		h.hooks[name] = hook
	}
}

// assumes already passed validate basic
func (h *TransferHooks) Validate(info types.CompletionHookCall) error {
	f, ok := h.hooks[info.HookName]
	if !ok {
		return gerrc.ErrNotFound.Wrapf("hook: name: %s", info.HookName)
	}
	return f.ValidateData(info.HookData)
}

// Should be called after packet finalization
// Recipient can either be the fulfiller of a hook that already occurred, or the original recipient still, who probably still wants the hook to happen
// NOTE: there is an asymmetry currently because on fulfill supports multiple hooks, but this finalization onRecv is hardcoded for x/forward atm
func (h *TransferHooks) OnRecvPacket(ctx sdk.Context, p *commontypes.RollappPacket) error {

	o, err := h.Keeper.PendingOrderByPacket(ctx, p)
	if errorsmod.IsOf(err, eibctypes.ErrDemandOrderDoesNotExist) {
		// not much we can do here, it should exist...
		h.Logger(ctx).Error("Pending order by packet not found.", "packet", p.LogString())
		return nil
	}
	if err != nil {
		return errorsmod.Wrap(err, "pending order by packet")
	}

	// for mvp, assume all completion hooks are only executable once
	if o.IsFulfilled() {
		// done
		return nil
	}

	if o.CompletionHookCall == nil {
		// done
		return nil
	}

	f, ok := h.hooks[o.CompletionHookCall.HookName]
	if !ok {
		return gerrc.ErrNotFound.Wrap("hook")
	}

	// the order wasn't fulfilled, so the funds were sent to the original recipient by ibc transfer app
	budget := sdk.NewCoin(o.Denom(), o.PriceAmount()) // TODO: fix amount, need to account for fees and so on
	return f.Run(ctx, o.GetRecipientBech32Address(), budget, o.CompletionHookCall.HookData)

}

func (h *TransferHooks) OnFulfill(ctx sdk.Context, o *eibctypes.DemandOrder) error {
	f, ok := h.hooks[o.CompletionHookCall.HookName]
	if !ok {
		return gerrc.ErrNotFound.Wrap("hook")
	}
	return f.Run(ctx, o.GetRecipientBech32Address(), sdk.NewCoin(o.Denom(), o.PriceAmount()), o.CompletionHookCall.HookData)
}
