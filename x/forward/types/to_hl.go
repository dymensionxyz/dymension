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
