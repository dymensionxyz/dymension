package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	ibccompletiontypes "github.com/dymensionxyz/dymension/v3/x/ibc_completion/types"
)

func NewHookForwardToIBC(
	sourceChannel string,
	receiver string,
	timeoutTimestamp uint64,
) *HookForwardToIBC {

	// sender will be ignored anyway, and replaced by the funds src (eibc fulfiller or HL recipient)
	arbSender, _ := sample.AccFromSecret("foo")

	return &HookForwardToIBC{
		Transfer: &ibctransfertypes.MsgTransfer{
			SourcePort:       "transfer",
			SourceChannel:    sourceChannel,
			Sender:           arbSender.String(),
			Token:            sdk.NewCoin("foo", math.NewInt(1)),
			Receiver:         receiver,
			TimeoutTimestamp: timeoutTimestamp,
		},
	}
}

func (h *HookForwardToIBC) ValidateBasic() error {
	if h.Transfer == nil {
		return gerrc.ErrInvalidArgument.Wrap("transfer is nil")
	}
	err := h.Transfer.ValidateBasic()
	if err != nil {
		return errorsmod.Wrap(err, "transfer")
	}
	return nil
}

func UnpackForwardToIBC(bz []byte) (*HookForwardToIBC, error) {
	var d HookForwardToIBC
	err := proto.Unmarshal(bz, &d)
	if err != nil {
		return nil, errorsmod.Wrap(err, "unmarshal forward hook")
	}
	if err := d.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic")
	}
	return &d, nil
}

func NewHookForwardToIBCCall(payload *HookForwardToIBC) (*commontypes.CompletionHookCall, error) {
	bz, err := proto.Marshal(payload)
	if err != nil {
		return &commontypes.CompletionHookCall{}, errorsmod.Wrap(err, "marshal forward hook")
	}

	return &commontypes.CompletionHookCall{
		Name: HookNameRollToIBC,
		Data: bz,
	}, nil
}

func NewHookForwardToIBCCallBz(payload *HookForwardToIBC) ([]byte, error) {
	h, err := NewHookForwardToIBCCall(payload)
	if err != nil {
		return nil, errorsmod.Wrap(err, "new forward to ibc hook")
	}

	bz, err := proto.Marshal(h)
	if err != nil {
		return nil, errorsmod.Wrap(err, "marshal forward hook")
	}

	return bz, nil
}

// returns memo as string to be directly included in outbound eibc transfer from rollapp
func MakeRolForwardToIBCMemoString(
	eibcFee string,
	data *HookForwardToIBC,
) (string, error) {

	bz, err := NewHookForwardToIBCCallBz(data)
	if err != nil {
		return "", errorsmod.Wrap(err, "new forward to ibc hook")
	}

	memo := delayedacktypes.CreateMemo(eibcFee, bz)
	return memo, nil
}

// returns memo as string to be directly included in outbound ibc transfer from e.g. osmosis
func MakeIBCForwardToIBCMemoString(
	data *HookForwardToIBC,
) (string, error) {

	bz, err := NewHookForwardToIBCCallBz(data)
	if err != nil {
		return "", errorsmod.Wrap(err, "new forward to ibc hook")
	}

	return ibccompletiontypes.MakeMemo(bz)
}
