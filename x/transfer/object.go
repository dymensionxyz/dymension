package keeper // have to call it keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/dymensionxyz/dymension/v3/x/transfer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

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

type EIBCK interface {
	PendingOrderByPacket(ctx sdk.Context, p *commontypes.RollappPacket) (*eibctypes.DemandOrder, error)
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

	if o.IsFulfilled() {
		// done
		return
	}

	if o.CompletionHook == nil {
		// done
		return
	}

	// TODO: do it

}

type FulfillArgs struct {
	FundsSource          sdk.AccAddress
	NewTransferRecipient sdk.AccAddress
	Fulfiller            sdk.AccAddress
}

func (m *TransferHooks) Fulfill(ctx sdk.Context, o *eibctypes.DemandOrder, args FulfillArgs) error {
	return nil
}

// assumed already passed validate basic
func (h *TransferHooks) Validate(info types.CompletionHook) error {
	f, ok := h.hooks[info.HookName]
	if !ok {
		return gerrc.ErrNotFound.Wrapf("hook: name: %s", info.HookName)
	}
	return f.ValidateData(info.HookData)
}

func (h *TransferHooks) Exec(ctx sdk.Context, order *eibctypes.DemandOrder, args FulfillArgs) error {
	f, ok := h.hooks[order.CompletionHook.HookName]
	if !ok {
		return gerrc.ErrNotFound.Wrap("hook")
	}
	return f.Run(ctx, order, args.FundsSource, args.NewTransferRecipient, args.Fulfiller, order.CompletionHook.HookData)
}
