package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/dymensionxyz/dymension/v3/utils/utransfer"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

const (
	HookNameForward = "forward"
)

// sender is computed
func NewHookEIBCtoHL(
	recovery *Recovery,
	tokenId hyperutil.HexAddress,
	destinationDomain uint32,
	recipient hyperutil.HexAddress,
	amount math.Int,
	maxFee sdk.Coin,

	gasLimit math.Int, // can be zero
	customHookId *hyperutil.HexAddress, // optional
	customHookMetadata string, // can be empty
) *HookEIBCtoHL {
	return &HookEIBCtoHL{
		Recovery: recovery,
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

func (h *HookEIBCtoHL) ValidateBasic() error {
	err := h.Recovery.ValidateBasic()
	if err != nil {
		return err
	}
	if h.HyperlaneTransfer == nil {
		return gerrc.ErrInvalidArgument
	}
	return nil
}

// token is computed
// sender is computed
// timeout height not supported
// next memo should go together in the top level of the HL memo
func NewHookHLtoIBC(
	sourcePort string,
	sourceChannel string,
	token sdk.Coin,
	receiver string,
	timeoutTimestamp uint64,
) *HookHLtoIBC {
	return &HookHLtoIBC{
		Transfer: &ibctransfertypes.MsgTransfer{
			SourcePort:       sourcePort,
			SourceChannel:    sourceChannel,
			Token:            token,
			Receiver:         receiver,
			TimeoutTimestamp: timeoutTimestamp,
		},
	}
}

// WARNING: assumes the memo is entirely dedicated to the HL->IBC forwarder
// TODO: also extract and then forward the rest of the memo, so that it can be used for other things later
func UnpackMemoFromHyperlane(bz []byte) (*HookHLtoIBC, []byte, error) {
	var d HookHLtoIBC
	err := proto.Unmarshal(bz, &d)
	if err != nil {
		return nil, nil, errorsmod.Wrap(err, "unmarshal forward hook")
	}
	return &d, nil, nil
}

func (h *HookHLtoIBC) ValidateBasic() error {
	err := h.Transfer.ValidateBasic()
	if err != nil {
		return err
	}
	err = h.Recovery.ValidateBasic()
	if err != nil {
		return err
	}
	return nil
}

func NewRecovery(
	address string,
) *Recovery {
	return &Recovery{
		Address: address,
	}
}

func (r *Recovery) ValidateBasic() error {
	_, err := r.AccAddr()
	if err != nil {
		return err
	}
	return nil
}

func (r *Recovery) AccAddr() (sdk.AccAddress, error) {
	addr, err := sdk.AccAddressFromBech32(r.Address)
	return addr, err
}

func (r *Recovery) MustAddr() sdk.AccAddress {
	return sdk.MustAccAddressFromBech32(r.Address)
}

func NewEIBCFulfillHook(payload *HookEIBCtoHL) (*eibctypes.FulfillHook, error) {
	bz, err := proto.Marshal(payload)
	if err != nil {
		return &eibctypes.FulfillHook{}, errorsmod.Wrap(err, "marshal forward hook")
	}

	return &eibctypes.FulfillHook{
		HookName: HookNameForward,
		HookData: bz,
	}, nil
}

func NewForwardMemo(
	eibcFee string,
	tokenId hyperutil.HexAddress,
	destinationDomain uint32,
	recipient hyperutil.HexAddress,
	amount math.Int,
	maxFee sdk.Coin,

	recoveryAddr string,

	gasLimit math.Int,
	customHookId *hyperutil.HexAddress,
	customHookMetadata string) (string, error) {

	hook, err := NewEIBCFulfillHook(
		NewHookEIBCtoHL(
			NewRecovery(recoveryAddr),
			tokenId,
			destinationDomain,
			recipient,
			amount,
			maxFee,
			gasLimit,
			customHookId,
			customHookMetadata,
		),
	)
	if err != nil {
		return "", err
	}
	if err := hook.ValidateBasic(); err != nil {
		return "", err
	}

	bz, err := proto.Marshal(hook)
	if err != nil {
		return "", err
	}

	return utransfer.CreateMemo(eibcFee, bz), nil
}
