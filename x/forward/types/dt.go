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
	HookNameForward = "dym-forward"
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
) *HookEIBCtoHL {
	return &HookEIBCtoHL{
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
	recoveryAddr string,
) *HookHLtoIBC {

	arbSender, _ := sample.AccFromSecret("foo")

	return &HookHLtoIBC{
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

// WARNING: assumes the memo is entirely dedicated to the HL->IBC forwarder
// TODO: also extract and then forward the rest of the memo, so that it can be used for other things later (include memo in ibc transfer so rollapp can use it)
func UnpackAppMemoFromHyperlaneMemo(bz []byte) (*HookHLtoIBC, []byte, error) {
	var d HookHLtoIBC
	err := proto.Unmarshal(bz, &d)
	if err != nil {
		return nil, nil, errorsmod.Wrap(err, "unmarshal forward hook")
	}
	if err := d.ValidateBasic(); err != nil {
		return nil, nil, errorsmod.Wrap(err, "validate basic")
	}
	var memo []byte
	return &d, memo, nil
}

func (h *HookHLtoIBC) ValidateBasic() error {
	if h.Transfer == nil {
		return gerrc.ErrInvalidArgument.Wrap("transfer is nil")
	}
	err := h.Transfer.ValidateBasic()
	if err != nil {
		return errorsmod.Wrap(err, "transfer")
	}
	return nil
}

func NewEIBCFulfillHook(payload *HookEIBCtoHL) (*transfertypes.CompletionHookCall, error) {
	bz, err := proto.Marshal(payload)
	if err != nil {
		return &transfertypes.CompletionHookCall{}, errorsmod.Wrap(err, "marshal forward hook")
	}

	return &transfertypes.CompletionHookCall{
		HookName: HookNameForward,
		HookData: bz,
	}, nil
}

func NewEIBCToHLMemoRaw(
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

// note, potentially expensive
func NewHyperlaneToIBCHyperlaneMessage(
	hyperlaneNonce uint32,
	hyperlaneSrcDomain uint32, // e.g. 1 for Ethereum
	hyperlaneSrcContract hyperutil.HexAddress, // e.g. Ethereum token contract as defined in token remote router
	hyperlaneDstDomain uint32, // e.g. 0 for Dymension
	hyperlaneTokenID hyperutil.HexAddress,
	hyperlaneRecipient sdk.AccAddress, // TODO: explain, ignored?
	hyperlaneTokenAmt math.Int, // must be at least hub token amount
	ibcSourceChan string, // e.g. channel-0
	ibcRecipient string, // address e.g. ethm1wqg8227q0p7pgp7lj7z6cu036l6eg34d9cp6lk
	hubToken sdk.Coin, // e.g. 50ibc/9A1EACD53A6A197ADC81DF9A49F0C4A26F7FF685ACF415EE726D7D59796E71A7
	ibcTimeoutTimestamp uint64, // e.g. 1000000000000000000
	recoveryAddr string, // funds recovery address on hub, e.g. dym1yecvrgz7yp26keaxa4r00554uugatxfegk76hz
) (hyperutil.HyperlaneMessage, error) {

	hook := NewHookHLtoIBC(
		"transfer",
		ibcSourceChan,
		hubToken,
		ibcRecipient,
		ibcTimeoutTimestamp,
		recoveryAddr,
	)

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
func decodeHyperlaneMessageEthHexToHyperlaneToEIBCMemo(s string) (*HookHLtoIBC, error) {
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
	d, _, err := UnpackAppMemoFromHyperlaneMemo(pl.Memo)
	if err != nil {
		return nil, errorsmod.Wrap(err, "unpack memo from hl message")
	}
	return d, nil
}
