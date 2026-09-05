package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	ibcompletiontypes "github.com/dymensionxyz/dymension/v3/x/ibc_completion/types"
)

// sender is computed
func NewHookForwardToHL(
	tokenId hyperutil.HexAddress,
	destinationDomain uint32,
	recipientFunds hyperutil.HexAddress,
	amount math.Int,
	maxFee sdk.Coin,
	gasLimit math.Int, // can be zero
	customHookId *hyperutil.HexAddress, // optional
	customHookMetadata string, // can be empty
	useFullBudget bool,
	minAmount math.Int,
) *HookForwardToHL {
	return &HookForwardToHL{
		UseFullBudget: useFullBudget,
		MinAmount:     minAmount,
		HyperlaneTransfer: &warptypes.MsgRemoteTransfer{
			TokenId:            tokenId,
			DestinationDomain:  destinationDomain,
			Recipient:          recipientFunds,
			Amount:             amount,
			CustomHookId:       customHookId,
			GasLimit:           gasLimit,
			MaxFee:             maxFee,
			CustomHookMetadata: customHookMetadata,
		},
	}
}

func (h *HookForwardToHL) ValidateBasic() error {
	if h.HyperlaneTransfer == nil {
		return gerrc.ErrInvalidArgument
	}
	return nil
}

func UnpackForwardToHL(bz []byte) (*HookForwardToHL, error) {
	var d HookForwardToHL
	err := proto.Unmarshal(bz, &d)
	if err != nil {
		return nil, errorsmod.Wrap(err, "unmarshal forward hook")
	}
	if err := d.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic")
	}
	return &d, nil
}

func NewHookForwardToHLCall(payload *HookForwardToHL) (*commontypes.CompletionHookCall, error) {
	bz, err := proto.Marshal(payload)
	if err != nil {
		return &commontypes.CompletionHookCall{}, errorsmod.Wrap(err, "marshal forward hook")
	}

	return &commontypes.CompletionHookCall{
		Name: HookNameRollToHL,
		Data: bz,
	}, nil
}

func NewHookForwardToHLCallBz(payload *HookForwardToHL) ([]byte, error) {
	call, err := NewHookForwardToHLCall(payload)
	if err != nil {
		return nil, errorsmod.Wrap(err, "new hook forward to hl call")
	}

	bz, err := proto.Marshal(call)
	if err != nil {
		return nil, errorsmod.Wrap(err, "marshal forward hook")
	}
	return bz, nil
}

// returns memo as string to be directly included in outbound eibc transfer from rollapp
func MakeRolForwardToHLMemoString(
	eibcFee string,
	payload *HookForwardToHL,
) (string, error) {
	bz, err := NewHookForwardToHLCallBz(payload)
	if err != nil {
		return "", errorsmod.Wrap(err, "make hook forward to hl call bytes")
	}

	return delayedacktypes.CreateMemo(eibcFee, bz), nil
}

// returns memo as string to be directly included in outbound eibc transfer from rollapp
func MakeIBCForwardToHLMemoString(
	payload *HookForwardToHL,
) (string, error) {
	bz, err := NewHookForwardToHLCallBz(payload)
	if err != nil {
		return "", errorsmod.Wrap(err, "make hook forward to hl call bytes")
	}

	return ibcompletiontypes.MakeMemo(bz)
}

// returns HLMetadata bytes to be included in hyperlane transfer metadata for HL-to-HL forwarding
func MakeHLForwardToHLMetadata(payload *HookForwardToHL) ([]byte, error) {
	bz, err := proto.Marshal(payload)
	if err != nil {
		return nil, errorsmod.Wrap(err, "marshal forward to hl hook")
	}

	metadata := &HLMetadata{
		HookForwardToHl: bz,
	}

	metadataBz, err := proto.Marshal(metadata)
	if err != nil {
		return nil, errorsmod.Wrap(err, "marshal hl metadata")
	}

	return metadataBz, nil
}

// ResolveHLForwardAmount decides how many base units to forward on the Hyperlane
// leg given the funds that actually arrived (budget).
func ResolveHLForwardAmount(budget sdk.Coin, d *HookForwardToHL) (math.Int, error) {
	mt := d.HyperlaneTransfer
	if mt.MaxFee.Denom != budget.Denom {
		return math.Int{}, gerrc.ErrInvalidArgument.Wrapf("max fee denom does not match allowed denom: %s != %s", mt.MaxFee.Denom, budget.Denom)
	}
	if mt.MaxFee.Amount.IsNil() || mt.MaxFee.Amount.IsNegative() {
		return math.Int{}, gerrc.ErrInvalidArgument.Wrap("max fee amount must be non-negative")
	}
	if !d.UseFullBudget {
		if mt.Amount.IsNil() {
			return math.Int{}, gerrc.ErrInvalidArgument.Wrap("transfer amount must be set")
		}
		maxCost := mt.MaxFee.Amount.Add(mt.Amount)
		if maxCost.GT(budget.Amount) {
			return math.Int{}, gerrc.ErrInvalidArgument.Wrapf("max cost (fee + amount)exceeds max budget %s > %s", maxCost, budget.Amount)
		}
		return mt.Amount, nil
	}
	send := budget.Amount.Sub(mt.MaxFee.Amount)
	if !send.IsPositive() {
		return math.Int{}, gerrc.ErrInvalidArgument.Wrapf("budget %s does not cover max fee %s", budget.Amount, mt.MaxFee.Amount)
	}
	if !d.MinAmount.IsNil() && d.MinAmount.IsPositive() && send.LT(d.MinAmount) {
		return math.Int{}, gerrc.ErrInvalidArgument.Wrapf("resolved send %s below min_amount %s", send, d.MinAmount)
	}
	return send, nil
}
