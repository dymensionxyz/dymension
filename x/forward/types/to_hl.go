package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	hyperutil "github.com/dymensionxyz/hyperlane-cosmos/util"
	warptypes "github.com/dymensionxyz/hyperlane-cosmos/x/warp/types"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	ibcompletiontypes "github.com/dymensionxyz/dymension/v3/x/ibc_completion/types"
)

// sender is computed
func NewHookForwardToHL(
	tokenId hyperutil.HexAddress,
	destinationDomain uint32,
	recipient hyperutil.HexAddress,
	amount math.Int,
	maxFee sdk.Coin,
	gasLimit math.Int, // can be zero
	customHookId *hyperutil.HexAddress, // optional
	customHookMetadata string, // can be empty
) *HookForwardToHL {
	return &HookForwardToHL{
		HyperlaneTransfer: &warptypes.MsgRemoteTransfer{
			TokenId:            tokenId,
			DestinationDomain:  destinationDomain,
			Recipient:          recipient,
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
func MakeHookForwardToHLCallBytes(
	tokenId hyperutil.HexAddress,
	destinationDomain uint32,
	recipient hyperutil.HexAddress,
	amount math.Int,
	maxFee sdk.Coin,
	gasLimit math.Int,
	customHookId *hyperutil.HexAddress,
	customHookMetadata string) ([]byte, error) {

	hook := NewHookForwardToHL(
		tokenId,
		destinationDomain,
		recipient,
		amount,
		maxFee,
		gasLimit,
		customHookId,
		customHookMetadata,
	)
	if err := hook.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic")
	}

	return NewHookForwardToHLCallBz(hook)
}

// returns memo as string to be directly included in outbound eibc transfer from rollapp
func MakeRolForwardToHLMemoString(
	eibcFee string,
	tokenId hyperutil.HexAddress,
	destinationDomain uint32,
	recipient hyperutil.HexAddress,
	amount math.Int,
	maxFee sdk.Coin,
	gasLimit math.Int,
	customHookId *hyperutil.HexAddress,
	customHookMetadata string) (string, error) {

	bz, err := MakeHookForwardToHLCallBytes(
		tokenId,
		destinationDomain,
		recipient,
		amount,
		maxFee,
		gasLimit,
		customHookId,
		customHookMetadata,
	)
	if err != nil {
		return "", errorsmod.Wrap(err, "make hook forward to hl call bytes")
	}

	return delayedacktypes.CreateMemo(eibcFee, bz), nil
}

// returns memo as string to be directly included in outbound eibc transfer from rollapp
func MakeIBCForwardToHLMemoString(
	tokenId hyperutil.HexAddress,
	destinationDomain uint32,
	recipient hyperutil.HexAddress,
	amount math.Int,
	maxFee sdk.Coin,
	gasLimit math.Int,
	customHookId *hyperutil.HexAddress,
	customHookMetadata string) (string, error) {

	bz, err := MakeHookForwardToHLCallBytes(
		tokenId,
		destinationDomain,
		recipient,
		amount,
		maxFee,
		gasLimit,
		customHookId,
		customHookMetadata,
	)
	if err != nil {
		return "", errorsmod.Wrap(err, "make hook forward to hl call bytes")
	}

	return ibcompletiontypes.MakeMemo(bz)
}
