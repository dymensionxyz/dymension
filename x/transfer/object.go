package keeper // have to call it keeper

import (
	errorsmod "cosmossdk.io/errors"
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

func NewTransferHooks(keeper EIBCK) *TransferHooks {
	return &TransferHooks{
		Keeper: keeper,
		hooks:  make(map[string]CompletionHookInstance),
	}
}

type CompletionHookInstance interface {
	ValidateData(hookData []byte) error
	Run(ctx sdk.Context, order *eibctypes.DemandOrder,
		fundsSource sdk.AccAddress,
		newTransferRecipient sdk.AccAddress,
		fulfiller sdk.AccAddress,
		hookData []byte) error
}

// map name -> instance
func (k TransferHooks) SetHooks(hooks map[string]CompletionHookInstance) {
	for name, hook := range hooks {
		k.hooks[name] = hook
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
func (m *TransferHooks) AfterRecvPacket(ctx sdk.Context, p *commontypes.RollappPacket) {

	o, err := m.Keeper.PendingOrderByPacket(ctx, p)
	if errorsmod.IsOf(err, eibctypes.ErrDemandOrderDoesNotExist) {
		// not much we can do here, it should exist...
		return
	}
	if err != nil {
		// TODO: something
		return
	}

	// for mvp, assume all completion hooks are only executable once
	if o.IsFulfilled() {
		// done
		return
	}

	if o.CompletionHookCall == nil {
		// done
		return
	}

	// TODO: do it

}

type EIBCFulfillArgs struct {
	FundsSource          sdk.AccAddress
	NewTransferRecipient sdk.AccAddress
	Fulfiller            sdk.AccAddress
}

func (h *TransferHooks) Fulfill(ctx sdk.Context, o *eibctypes.DemandOrder, args EIBCFulfillArgs) error {
	f, ok := h.hooks[o.CompletionHookCall.HookName]
	if !ok {
		return gerrc.ErrNotFound.Wrap("hook")
	}
	return f.Run(ctx, o, args.FundsSource, args.NewTransferRecipient, args.Fulfiller, o.CompletionHookCall.HookData)
}
