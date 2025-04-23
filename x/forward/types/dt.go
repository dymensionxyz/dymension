package types

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	hypercoretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	warpkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/warp/keeper"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/utils/utransfer"
	transfertypes "github.com/dymensionxyz/dymension/v3/x/transfer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

const (
	// not to be confused with ibc apps PFM which uses 'forward' as the fungible packet json memo key
	HookNameRollToHL  = "dym-fwd-roll-hl"
	HookNameRollToIBC = "dym-fwd-roll-ibc"
)

// sender is computed
func NewHookEIBCtoHL(
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

// token is computed
// sender is computed
// timeout height not supported
func NewHookForwardToIBC(
	sourcePort string,
	sourceChannel string,
	token sdk.Coin,
	receiver string,
	timeoutTimestamp uint64,
) *HookForwardToIBC {

	arbSender, _ := sample.AccFromSecret("foo")

	return &HookForwardToIBC{
		Transfer: &ibctransfertypes.MsgTransfer{
			SourcePort:       sourcePort,
			SourceChannel:    sourceChannel,
			Token:            token,
			Sender:           arbSender.String(),
			Receiver:         receiver,
			TimeoutTimestamp: timeoutTimestamp,
		},
	}
}

func MakeHookForwardToIBC(
	sourceChannel string,
	token sdk.Coin,
	receiver string,
	timeoutTimestamp uint64,
) *HookForwardToIBC {
	hook := NewHookForwardToIBC(
		"transfer",
		sourceChannel,
		token,
		receiver,
		timeoutTimestamp,
	)
	return hook
}

func UnpackForwardToIBCFromHyperlaneMemo(bz []byte) (*HookForwardToIBC, error) {
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

func NewRollToHLHook(payload *HookForwardToHL) (*transfertypes.CompletionHookCall, error) {
	bz, err := proto.Marshal(payload)
	if err != nil {
		return &transfertypes.CompletionHookCall{}, errorsmod.Wrap(err, "marshal forward hook")
	}

	return &transfertypes.CompletionHookCall{
		Name: HookNameRollToHL,
		Data: bz,
	}, nil
}

func NewRollToHLMemoRaw(
	eibcFee string,
	tokenId hyperutil.HexAddress,
	destinationDomain uint32,
	recipient hyperutil.HexAddress,
	amount math.Int,
	maxFee sdk.Coin,

	gasLimit math.Int,
	customHookId *hyperutil.HexAddress,
	customHookMetadata string) (string, error) {

	hook, err := NewRollToHLHook(
		NewHookEIBCtoHL(
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

// note, potentially expensive due to many encodes/decodes, avoid using on chain
func NewForwardToIBCHyperlaneMessage(
	hyperlaneNonce uint32,
	hyperlaneSrcDomain uint32, // e.g. 1 for Ethereum
	hyperlaneSrcContract hyperutil.HexAddress, // e.g. Ethereum token contract as defined in token remote router
	hyperlaneDstDomain uint32, // e.g. 0 for Dymension
	hyperlaneTokenID hyperutil.HexAddress,
	hyperlaneRecipient sdk.AccAddress, // TODO: explain, ignored?
	hyperlaneTokenAmt math.Int, // must be at least hub token amount
	hook *HookForwardToIBC,
) (hyperutil.HyperlaneMessage, error) {

	if err := hook.ValidateBasic(); err != nil {
		return hyperutil.HyperlaneMessage{}, errorsmod.Wrap(err, "validate basic")
	}

	memoBz, err := proto.Marshal(hook)
	if err != nil {
		return hyperutil.HyperlaneMessage{}, err
	}

	hlM, err := warpkeeper.CreateTestHyperlaneMessage(
		hypercoretypes.MESSAGE_VERSION,
		hyperlaneNonce,
		hyperlaneSrcDomain,
		hyperlaneSrcContract,
		hyperlaneDstDomain,
		hyperlaneTokenID,
		hyperlaneRecipient,
		hyperlaneTokenAmt,
		memoBz,
	)
	if err != nil {
		return hyperutil.HyperlaneMessage{}, err
	}

	// sanity
	{
		s := hlM.String()
		_, err := decodeHyperlaneMessageEthHexToHyperlaneToEIBCMemo(s)
		if err != nil {
			return hyperutil.HyperlaneMessage{}, errorsmod.Wrap(err, "decode eth hex")
		}
	}

	return hlM, nil
}

// intended for tests/clients, expensive
func decodeHyperlaneMessageEthHexToHyperlaneToEIBCMemo(s string) (*HookForwardToIBC, error) {
	decoded, err := hyperutil.DecodeEthHex(s)
	if err != nil {
		return nil, errorsmod.Wrap(err, "decode eth hex")
	}
	warpM, err := hyperutil.ParseHyperlaneMessage(decoded)
	if err != nil {
		return nil, errorsmod.Wrap(err, "parse hl message")
	}
	pl, err := warptypes.ParseWarpMemoPayload(warpM.Body)
	if err != nil {
		return nil, errorsmod.Wrap(err, "parse warp memo")
	}
	d, err := UnpackForwardToIBCFromHyperlaneMemo(pl.Memo)
	if err != nil {
		return nil, errorsmod.Wrap(err, "unpack memo from hl message")
	}
	return d, nil
}
