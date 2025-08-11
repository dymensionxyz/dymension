package cli

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	errorsmod "cosmossdk.io/errors"
	math "cosmossdk.io/math"
	"github.com/bcp-innovations/hyperlane-cosmos/util"
	hypercoretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
)

const (
	FlagReadable = "readable"
	FlagSrc      = "src"
	FlagDst      = "dst"

	FlagTokenID = "token-id"
	FlagAmount  = "amount"
	FlagMaxFee  = "max-fee"

	FlagDomain    = "domain"
	FlagSrcDomain = "src-domain"
	FlagDstDomain = "dst-domain"
	FlagHubDomain = "hub-domain"
	FlagKasDomain = "kas-domain"

	// FlagRecipientDst is the final recipient address on the destination chain (IBC or Hyperlane)
	// This is where the funds will ultimately be delivered
	FlagRecipientDst = "funds-recipient-dst"

	// FlagRecipientHub is the recipient address on the Dymension Hub
	// This serves dual purposes:
	// 1. For Kaspa->Hub transfers: this is the final recipient of funds
	// 2. For Kaspa->IBC/HL transfers: this is the intermediary address that temporarily holds funds
	//    and also serves as the recovery address if forwarding fails
	FlagRecipientHub = "funds-recipient-hub"

	FlagChannel = "channel"
	FlagTimeout = "timeout"

	FlagEIBCFee = "eibc-fee"

	FlagHLAmount     = "hl-amount"
	FlagHLGas        = "hl-gas"
	FlagBridgeFeeMul = "bridge-fee-mul"

	FlagNonce       = "nonce"
	FlagSrcContract = "src-contract"
	FlagDstTokenID  = "dst-token-id" // #nosec G101 - This is a CLI flag name, not a credential
	FlagDstAmount   = "dst-amount"

	FlagKasToken = "kas-token"

	FlagDecodeMemo = "decode-memo"

	SrcIBC   = "ibc"
	SrcEIBC  = "eibc"
	SrcHL    = "hl"
	SrcKaspa = "kaspa"

	DstIBC = "ibc"
	DstHL  = "hl"
	DstHub = "hub"
)

type CommonParams struct {
	Readable bool
	Src      string
	Dst      string
}

type TokenParams struct {
	TokenID   util.HexAddress
	Amount    math.Int
	Recipient sdk.AccAddress
}

type HyperlaneParams struct {
	TokenID        util.HexAddress
	Amount         math.Int
	RecipientFunds util.HexAddress
	Domain         uint32
	MaxFee         sdk.Coin
	Nonce          uint32
	SrcDomain      uint32
	SrcContract    util.HexAddress
	DstDomain      uint32
}

type IBCParams struct {
	Channel   string
	Recipient string
	Timeout   time.Duration
}

type KaspaParams struct {
	TokenPlaceholder util.HexAddress
	KasDomain        uint32
	HubDomain        uint32
	FundsRecipient   sdk.AccAddress
}

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Utility commands for %s logic", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdCreateMemo())
	cmd.AddCommand(CmdCreateHLMessage())
	cmd.AddCommand(CmdDecodeHL())
	cmd.AddCommand(CmdCosmosAddrToHLAddr())
	cmd.AddCommand(CmdEstimateFees())

	// Backward compatibility
	cmd.AddCommand(CmdMemoIBCtoHL())
	cmd.AddCommand(CmdMemoIBCtoIBC())
	cmd.AddCommand(CmdMemoEIBCtoHL())
	cmd.AddCommand(CmdMemoEIBCtoIBC())
	cmd.AddCommand(CmdMemoHLtoIBCRaw())
	cmd.AddCommand(CmdMemoHLtoHLRaw())
	cmd.AddCommand(CmdTestHLtoIBCMessage())
	cmd.AddCommand(CmdHLMessageKaspaToHub())
	cmd.AddCommand(CmdHLMessageKaspaToIBC())
	cmd.AddCommand(CmdHLMessageKaspaToHL())

	return cmd
}

func CmdCreateMemo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-memo",
		Short: "Create a memo for forwarding between different protocols",
		Long: `Create a memo for forwarding tokens between IBC, EIBC, and Hyperlane.
The src and destination determine which type of memo is created.`,
		Example: `# IBC to Hyperlane
dymd q forward create-memo --src=ibc --dst=hl \
  --token-id=0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0 \
  --dst-domain=1 \
  --funds-recipient-dst=0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0 \
  --amount=10000 \
  --max-fee=20foo

# EIBC to IBC
dymd q forward create-memo --src=eibc --dst=ibc \
  --eibc-fee=100 \
  --channel=channel-0 \
  --funds-recipient-dst=ethm1a30y0h95a7p38plnv5s02lzrgcy0m0xumq0ymn \
  --timeout=5m`,
		RunE: runCreateMemo,
	}

	addCommonFlags(cmd)
	addTokenFlags(cmd)
	addHyperlaneFlags(cmd)
	addIBCFlags(cmd)
	cmd.Flags().String(FlagEIBCFee, "", "EIBC fee (required for EIBC src)")

	return cmd
}

func CmdCreateHLMessage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-hl-message",
		Short: "Create a Hyperlane message for testing or for including in Kaspa payloads",
		Long: `Create a Hyperlane message from various srcs (HL, Kaspa) to various destination (Hub, IBC, HL).
The src and destination determine the message format and required parameters.`,
		Example: `# Kaspa to Hub (no forwarding - hub recipient is final recipient)
dymd q forward create-hl-message --src=kaspa --dst=hub \
  --token-id=0x0000000000000000000000000000000000000000000000000000000000000000 \
  --funds-recipient-hub=dym139mq752delxv78jvtmwxhasyrycufsvrw4aka9 \
  --amount=1000000000000000000 \
  --kas-token=0x0000000000000000000000000000000000000000000000000000000000000000 \
  --kas-domain=80808082 \
  --hub-domain=1260813472

# Kaspa to IBC (hub recipient is intermediary/recovery address)
dymd q forward create-hl-message --src=kaspa --dst=ibc \
  --token-id=0x0000000000000000000000000000000000000000000000000000000000000000 \
  --funds-recipient-hub=dym139mq752delxv78jvtmwxhasyrycufsvrw4aka9 \
  --amount=1000000000000000000 \
  --kas-token=0x0000000000000000000000000000000000000000000000000000000000000000 \
  --kas-domain=80808082 \
  --hub-domain=1260813472 \
  --channel=channel-0 \
  --funds-recipient-dst=ethm1a30y0h95a7p38plnv5s02lzrgcy0m0xumq0ymn \
  --timeout=5m

# HL to HL forwarding
dymd q forward create-hl-message --src=hl --dst=hl \
  --nonce=123 \
  --src-domain=1 \
  --src-contract=0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0 \
  --dst-domain=2 \
  --token-id=0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0 \
  --funds-recipient-dst=0x1234567890123456789012345678901234567890 \
  --amount=1000000000000000000 \
  --dst-amount=900000000000000000 \
  --max-fee=20foo`,
		RunE: runCreateHLMessage,
	}

	addCommonFlags(cmd)
	addTokenFlags(cmd)
	addHyperlaneFlags(cmd)
	addIBCFlags(cmd)
	addKaspaFlags(cmd)

	return cmd
}

func CmdDecodeHL() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "decode-hl [body|message] [hexsing]",
		Args:    cobra.ExactArgs(2),
		Short:   "Decode a Hyperlane message or body from hex string",
		Example: `dymd q forward decode-hl message 0x00000000... --decode-memo`,
		RunE:    runDecodeHL,
	}

	cmd.Flags().Bool(FlagDecodeMemo, false, "Decode the memo from the payload")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdCosmosAddrToHLAddr() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cosmos-addr-to-hl-addr [address]",
		Args:    cobra.ExactArgs(1),
		Short:   "Convert a Cosmos address to Hyperlane recipient format",
		Example: `dymd q forward cosmos-addr-to-hl-addr dym139mq752delxv78jvtmwxhasyrycufsvrw4aka9`,
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, err := EthRecipient(args[0])
			if err != nil {
				return fmt.Errorf("eth recipient: %w", err)
			}
			fmt.Println(addr)
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdEstimateFees() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "estimate-fees",
		Short:   "Estimate fees for EIBC to HL transfers",
		Example: `dymd q forward estimate-fees --hl-amount=125000000000000 --hl-gas=200000 --eibc-fee=2000 --bridge-fee-mul=0.01`,
		RunE:    runEstimateFees,
	}

	cmd.Flags().String(FlagHLAmount, "", "Amount to receive on Hyperlane")
	cmd.Flags().String(FlagHLGas, "", "Max gas for Hyperlane")
	cmd.Flags().String(FlagEIBCFee, "", "EIBC fee")
	cmd.Flags().String(FlagBridgeFeeMul, "", "Bridge fee multiplier")
	_ = cmd.MarkFlagRequired(FlagHLAmount)
	_ = cmd.MarkFlagRequired(FlagHLGas)
	_ = cmd.MarkFlagRequired(FlagEIBCFee)
	_ = cmd.MarkFlagRequired(FlagBridgeFeeMul)

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(FlagReadable, false, "Show output in human-readable format")
	cmd.Flags().String(FlagSrc, "", "Src protocol (ibc, eibc, hl, kaspa)")
	cmd.Flags().String(FlagDst, "", "destination protocol (ibc, hl, hub)")
	_ = cmd.MarkFlagRequired(FlagSrc)
	_ = cmd.MarkFlagRequired(FlagDst)
}

func addTokenFlags(cmd *cobra.Command) {
	cmd.Flags().String(FlagTokenID, "", "Token ID (hex)")
	cmd.Flags().String(FlagAmount, "", "Token amount")
	cmd.Flags().String(FlagMaxFee, "", "Maximum fee (e.g., 20foo)")
}

func addHyperlaneFlags(cmd *cobra.Command) {
	cmd.Flags().Uint32(FlagNonce, 0, "Message nonce")
	cmd.Flags().Uint32(FlagDomain, 0, "Domain ID")
	cmd.Flags().Uint32(FlagSrcDomain, 0, "Source domain ID")
	cmd.Flags().Uint32(FlagDstDomain, 0, "Destination domain ID")
	cmd.Flags().String(FlagSrcContract, "", "Source contract (hex)")
	cmd.Flags().String(FlagDstTokenID, "", "Destination token ID (hex)")
	cmd.Flags().String(FlagRecipientDst, "", "Final recipient address on destination chain")
	cmd.Flags().String(FlagDstAmount, "", "Destination amount")
}

func addIBCFlags(cmd *cobra.Command) {
	cmd.Flags().String(FlagChannel, "", "IBC channel")
	cmd.Flags().String(FlagTimeout, "5m", "IBC timeout duration")
}

func addKaspaFlags(cmd *cobra.Command) {
	cmd.Flags().String(FlagKasToken, "", "Kaspa token placeholder (hex)")
	cmd.Flags().Uint32(FlagKasDomain, 0, "Kaspa domain ID")
	cmd.Flags().Uint32(FlagHubDomain, 0, "Hub domain ID")
	cmd.Flags().String(FlagRecipientHub, "", "Hub recipient address (final recipient for Kaspa->Hub, or intermediary/recovery address for Kaspa->IBC/HL)")
}

func parseCommonFlags(cmd *cobra.Command) (*CommonParams, error) {
	readable, _ := cmd.Flags().GetBool(FlagReadable)
	src, _ := cmd.Flags().GetString(FlagSrc)
	dst, _ := cmd.Flags().GetString(FlagDst)

	return &CommonParams{
		Readable: readable,
		Src:      src,
		Dst:      dst,
	}, nil
}

func parseTokenFlags(cmd *cobra.Command) (*TokenParams, error) {
	tokenIDS, _ := cmd.Flags().GetString(FlagTokenID)
	amountS, _ := cmd.Flags().GetString(FlagAmount)
	recipientS, _ := cmd.Flags().GetString(FlagRecipientDst)

	var params TokenParams
	var err error

	if tokenIDS != "" {
		params.TokenID, err = util.DecodeHexAddress(tokenIDS)
		if err != nil {
			return nil, fmt.Errorf("invalid token ID: %w", err)
		}
	}

	if amountS != "" {
		var ok bool
		params.Amount, ok = math.NewIntFromString(amountS)
		if !ok {
			return nil, fmt.Errorf("invalid amount: %s", amountS)
		}
	}

	if recipientS != "" {
		params.Recipient, err = sdk.AccAddressFromBech32(recipientS)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient: %w", err)
		}
	}

	return &params, nil
}

func parseHyperlaneFlags(cmd *cobra.Command) (*HyperlaneParams, error) {
	var params HyperlaneParams
	var err error

	params.Nonce, _ = cmd.Flags().GetUint32(FlagNonce)
	params.Domain, _ = cmd.Flags().GetUint32(FlagDomain)
	params.SrcDomain, _ = cmd.Flags().GetUint32(FlagSrcDomain)
	params.DstDomain, _ = cmd.Flags().GetUint32(FlagDstDomain)

	tokenIDS, _ := cmd.Flags().GetString(FlagTokenID)
	if tokenIDS != "" {
		params.TokenID, err = util.DecodeHexAddress(tokenIDS)
		if err != nil {
			return nil, fmt.Errorf("invalid token ID: %w", err)
		}
	}

	srcContractS, _ := cmd.Flags().GetString(FlagSrcContract)
	if srcContractS != "" {
		params.SrcContract, err = util.DecodeHexAddress(srcContractS)
		if err != nil {
			return nil, fmt.Errorf("invalid src contract: %w", err)
		}
	}

	recipientFundsS, _ := cmd.Flags().GetString(FlagRecipientDst)
	if recipientFundsS != "" {
		params.RecipientFunds, err = util.DecodeHexAddress(recipientFundsS)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient: %w", err)
		}
	}

	amountS, _ := cmd.Flags().GetString(FlagAmount)
	if amountS != "" {
		var ok bool
		params.Amount, ok = math.NewIntFromString(amountS)
		if !ok {
			return nil, fmt.Errorf("invalid amount: %s", amountS)
		}
	}

	maxFeeS, _ := cmd.Flags().GetString(FlagMaxFee)
	if maxFeeS != "" {
		params.MaxFee, err = sdk.ParseCoinNormalized(maxFeeS)
		if err != nil {
			return nil, fmt.Errorf("invalid max fee: %w", err)
		}
	}

	return &params, nil
}

func parseIBCFlags(cmd *cobra.Command) (*IBCParams, error) {
	channel, _ := cmd.Flags().GetString(FlagChannel)
	recipient, _ := cmd.Flags().GetString(FlagRecipientDst)

	timeoutS, _ := cmd.Flags().GetString(FlagTimeout)
	timeout, err := time.ParseDuration(timeoutS)
	if err != nil {
		return nil, fmt.Errorf("invalid timeout: %w", err)
	}

	return &IBCParams{
		Channel:   channel,
		Recipient: recipient,
		Timeout:   timeout,
	}, nil
}

func parseKaspaFlags(cmd *cobra.Command) (*KaspaParams, error) {
	var params KaspaParams
	var err error

	kasTokenS, _ := cmd.Flags().GetString(FlagKasToken)
	if kasTokenS != "" {
		params.TokenPlaceholder, err = util.DecodeHexAddress(kasTokenS)
		if err != nil {
			return nil, fmt.Errorf("invalid kas token: %w", err)
		}
	}

	params.KasDomain, _ = cmd.Flags().GetUint32(FlagKasDomain)
	params.HubDomain, _ = cmd.Flags().GetUint32(FlagHubDomain)

	hubFundsRecipientS, _ := cmd.Flags().GetString(FlagRecipientHub)
	if hubFundsRecipientS != "" {
		params.FundsRecipient, err = sdk.AccAddressFromBech32(hubFundsRecipientS)
		if err != nil {
			return nil, fmt.Errorf("invalid hub recipient: %w", err)
		}
	}

	return &params, nil
}

func runCreateMemo(cmd *cobra.Command, args []string) error {
	common, err := parseCommonFlags(cmd)
	if err != nil {
		return err
	}

	switch common.Src {
	case SrcIBC, SrcEIBC:
		return runCreateMemoFromIBC(cmd, common)
	case SrcHL:
		return runCreateMemoFromHL(cmd, common)
	default:
		return fmt.Errorf("unsupported src: %s", common.Src)
	}
}

func runCreateMemoFromIBC(cmd *cobra.Command, common *CommonParams) error {
	var memo string

	switch common.Dst {
	case DstHL:
		hlParams, err := parseHyperlaneFlags(cmd)
		if err != nil {
			return err
		}

		tokenParams, err := parseTokenFlags(cmd)
		if err != nil {
			return err
		}

		hook := types.NewHookForwardToHL(
			hlParams.TokenID,
			hlParams.DstDomain,
			hlParams.RecipientFunds,
			tokenParams.Amount,
			hlParams.MaxFee,
			math.ZeroInt(),
			nil,
			"",
		)

		if common.Src == SrcEIBC {
			eibcFeeS, _ := cmd.Flags().GetString(FlagEIBCFee)
			memo, err = types.MakeRolForwardToHLMemoString(eibcFeeS, hook)
		} else {
			memo, err = types.MakeIBCForwardToHLMemoString(hook)
		}

		if err != nil {
			return fmt.Errorf("create memo: %w", err)
		}

	case DstIBC:
		ibcParams, err := parseIBCFlags(cmd)
		if err != nil {
			return err
		}

		hook := types.NewHookForwardToIBC(
			ibcParams.Channel,
			ibcParams.Recipient,
			uint64(time.Now().Add(ibcParams.Timeout).UnixNano()), // #nosec G115 - Unix time is always positive
		)

		if common.Src == SrcEIBC {
			eibcFeeS, _ := cmd.Flags().GetString(FlagEIBCFee)
			memo, err = types.MakeRolForwardToIBCMemoString(eibcFeeS, hook)
		} else {
			memo, err = types.MakeIBCForwardToIBCMemoString(hook)
		}

		if err != nil {
			return fmt.Errorf("create memo: %w", err)
		}

	default:
		return fmt.Errorf("unsupported destination for IBC/EIBC src: %s", common.Dst)
	}

	fmt.Println(memo)
	return nil
}

func runCreateMemoFromHL(cmd *cobra.Command, common *CommonParams) error {
	switch common.Dst {
	case DstIBC:
		ibcParams, err := parseIBCFlags(cmd)
		if err != nil {
			return err
		}

		hook := types.NewHookForwardToIBC(
			ibcParams.Channel,
			ibcParams.Recipient,
			uint64(time.Now().Add(ibcParams.Timeout).UnixNano()), // #nosec G115 - Unix time is always positive // #nosec G115 - Unix time is always positive
		)

		if common.Readable {
			fmt.Printf("hyperlane message: %+v\n", hook)
		} else {
			bz, err := proto.Marshal(hook)
			if err != nil {
				return fmt.Errorf("marshal: %w", err)
			}
			hlMetadata := &types.HLMetadata{
				HookForwardToIbc: bz,
			}
			bz, err = proto.Marshal(hlMetadata)
			if err != nil {
				return fmt.Errorf("marshal: %w", err)
			}
			fmt.Printf("%s\n", util.EncodeEthHex(bz))
		}

	case DstHL:
		hlParams, err := parseHyperlaneFlags(cmd)
		if err != nil {
			return err
		}

		tokenParams, err := parseTokenFlags(cmd)
		if err != nil {
			return err
		}

		hook := types.NewHookForwardToHL(
			hlParams.TokenID,
			hlParams.DstDomain,
			hlParams.RecipientFunds,
			tokenParams.Amount,
			hlParams.MaxFee,
			math.ZeroInt(),
			nil,
			"",
		)

		if common.Readable {
			fmt.Printf("hyperlane message: %+v\n", hook)
		} else {
			bz, err := proto.Marshal(hook)
			if err != nil {
				return fmt.Errorf("marshal: %w", err)
			}
			hlMetadata := &types.HLMetadata{
				HookForwardToHl: bz,
			}
			bz, err = proto.Marshal(hlMetadata)
			if err != nil {
				return fmt.Errorf("marshal: %w", err)
			}
			fmt.Printf("%s\n", util.EncodeEthHex(bz))
		}

	default:
		return fmt.Errorf("unsupported destination for HL src: %s", common.Dst)
	}

	return nil
}

func runCreateHLMessage(cmd *cobra.Command, args []string) error {
	common, err := parseCommonFlags(cmd)
	if err != nil {
		return err
	}

	switch common.Src {
	case SrcKaspa:
		return runCreateHLMessageFromKaspa(cmd, common)
	case SrcHL:
		return runCreateHLMessageFromHL(cmd, common)
	default:
		return fmt.Errorf("unsupported src: %s", common.Src)
	}
}

func runCreateHLMessageFromKaspa(cmd *cobra.Command, common *CommonParams) error {
	kaspaParams, err := parseKaspaFlags(cmd)
	if err != nil {
		return err
	}

	tokenParams, err := parseTokenFlags(cmd)
	if err != nil {
		return err
	}

	var memo []byte

	switch common.Dst {
	case DstHub:
		// For Kaspa->Hub transfers, create minimal metadata to satisfy validation
		// The hub recipient (kaspaParams.FundsRecipient) is the final recipient
		hlMetadata := &types.HLMetadata{
			// All fields are empty as no forwarding is needed
			Kaspa:            nil,
			HookForwardToHl:  nil,
			HookForwardToIbc: nil,
		}
		memo, err = proto.Marshal(hlMetadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}
	case DstIBC:
		// For Kaspa->IBC transfers:
		// - kaspaParams.FundsRecipient (hub recipient) is the intermediary/recovery address
		// - ibcParams.Recipient is the final recipient on the IBC destination chain
		ibcParams, err := parseIBCFlags(cmd)
		if err != nil {
			return err
		}

		hook := types.NewHookForwardToIBC(
			ibcParams.Channel,
			ibcParams.Recipient,
			uint64(time.Now().Add(ibcParams.Timeout).UnixNano()), // #nosec G115 - Unix time is always positive
		)

		hookBz, err := proto.Marshal(hook)
		if err != nil {
			return fmt.Errorf("marshal hook: %w", err)
		}

		hlMetadata := &types.HLMetadata{
			HookForwardToIbc: hookBz,
		}
		memo, err = proto.Marshal(hlMetadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}

	case DstHL:
		// For Kaspa->HL transfers:
		// - kaspaParams.FundsRecipient (hub recipient) is the intermediary/recovery address
		// - hlParams.RecipientFunds is the final recipient on the Hyperlane destination chain
		hlParams, err := parseHyperlaneFlags(cmd)
		if err != nil {
			return err
		}

		dstAmountS, _ := cmd.Flags().GetString(FlagDstAmount)
		dstAmount, ok := math.NewIntFromString(dstAmountS)
		if !ok {
			return fmt.Errorf("invalid destination amount")
		}

		hook := types.NewHookForwardToHL(
			hlParams.TokenID,
			hlParams.DstDomain,
			hlParams.RecipientFunds,
			dstAmount,
			hlParams.MaxFee,
			math.ZeroInt(),
			nil,
			"",
		)

		hookBz, err := proto.Marshal(hook)
		if err != nil {
			return fmt.Errorf("marshal hook: %w", err)
		}

		hlMetadata := &types.HLMetadata{
			HookForwardToHl: hookBz,
		}
		memo, err = proto.Marshal(hlMetadata)
		if err != nil {
			return fmt.Errorf("marshal metadata: %w", err)
		}

	default:
		return fmt.Errorf("unsupported destination: %s", common.Dst)
	}

	// Create the Hyperlane message
	// - tokenParams.TokenID: becomes HyperlaneMessage.Recipient (the warp router contract)
	// - kaspaParams.FundsRecipient: becomes the recipient in the warp payload
	//   This is the hub address that will receive funds (either as final recipient or intermediary)
	m, err := createHyperlaneMessage(
		0, // Fixed nonce for Kaspa
		kaspaParams.KasDomain,
		kaspaParams.TokenPlaceholder,
		kaspaParams.HubDomain,
		tokenParams.TokenID,
		kaspaParams.FundsRecipient,
		tokenParams.Amount,
		memo,
	)
	if err != nil {
		return fmt.Errorf("create hyperlane message: %w", err)
	}

	if common.Readable {
		fmt.Printf("hyperlane message: %+v\n", m)
		switch common.Dst {
		case DstIBC:
			fmt.Printf("ibc forward details: channel=%s, recipient=%s\n",
				cmd.Flag(FlagChannel).Value, cmd.Flag(FlagRecipientDst).Value)
		case DstHL:
			fmt.Printf("hl forward details: domain=%s, recipient=%s\n",
				cmd.Flag(FlagDstDomain).Value, cmd.Flag(FlagRecipientDst).Value)
		}
	} else {
		fmt.Print(m)
	}

	return nil
}

func runCreateHLMessageFromHL(cmd *cobra.Command, common *CommonParams) error {
	hlParams, err := parseHyperlaneFlags(cmd)
	if err != nil {
		return err
	}

	tokenParams, err := parseTokenFlags(cmd)
	if err != nil {
		return err
	}

	var m util.HyperlaneMessage

	switch common.Dst {
	case DstIBC:
		ibcParams, err := parseIBCFlags(cmd)
		if err != nil {
			return err
		}

		hook := types.NewHookForwardToIBC(
			ibcParams.Channel,
			ibcParams.Recipient,
			uint64(time.Now().Add(ibcParams.Timeout).UnixNano()), // #nosec G115 - Unix time is always positive
		)

		m, err = MakeForwardToIBCHyperlaneMessage(
			hlParams.Nonce,
			hlParams.SrcDomain,
			hlParams.SrcContract,
			hlParams.DstDomain,
			hlParams.TokenID,
			tokenParams.Recipient,
			tokenParams.Amount,
			hook,
		)
		if err != nil {
			return fmt.Errorf("create hl message: %w", err)
		}

	case DstHL:
		// For HL->HL transfers:
		// - tokenParams.Recipient is the hub address that receives funds
		// - hlParams.RecipientFunds is the final recipient on the destination Hyperlane chain
		// Need to get destination amount (what the final recipient will receive)
		dstAmountS, _ := cmd.Flags().GetString(FlagDstAmount)
		dstAmount, ok := math.NewIntFromString(dstAmountS)
		if !ok {
			return fmt.Errorf("invalid destination amount")
		}

		hook := types.NewHookForwardToHL(
			hlParams.TokenID,
			hlParams.DstDomain,
			hlParams.RecipientFunds,
			dstAmount,
			hlParams.MaxFee,
			math.ZeroInt(),
			nil,
			"",
		)

		m, err = MakeForwardToHLHyperlaneMessage(
			hlParams.Nonce,
			hlParams.SrcDomain,
			hlParams.SrcContract,
			hlParams.DstDomain,
			hlParams.TokenID,
			tokenParams.Recipient,
			tokenParams.Amount,
			hook,
		)
		if err != nil {
			return fmt.Errorf("create hl message: %w", err)
		}

	default:
		return fmt.Errorf("unsupported destination for HL src: %s", common.Dst)
	}

	if common.Readable {
		fmt.Printf("hyperlane message: %+v\n", m)
	} else {
		fmt.Print(m)
	}

	return nil
}

func runDecodeHL(cmd *cobra.Command, args []string) error {
	kind := args[0]
	if kind != "body" && kind != "message" {
		return fmt.Errorf("unsupported message type: %s (use 'body' or 'message')", kind)
	}

	d := args[1]
	d = strings.TrimSpace(d)
	d = strings.ReplaceAll(d, `\`, "")
	d = strings.ReplaceAll(d, " ", "")

	fmt.Printf("input: %s\n", d)

	bz, err := util.DecodeEthHex(d)
	if err != nil {
		return fmt.Errorf("decode eth hex: %w", err)
	}

	body := bz
	if kind == "message" {
		m, err := util.ParseHyperlaneMessage(bz)
		if err != nil {
			return fmt.Errorf("parse message: %w", err)
		}
		body = m.Body
		fmt.Printf("hyperlane message: %+v\n", m)
	}

	decodeMemo, _ := cmd.Flags().GetBool(FlagDecodeMemo)

	warpPL, err := warptypes.ParseWarpPayload(body)
	if err != nil {
		return fmt.Errorf("parse warp payload: %w", err)
	}

	if decodeMemo {
		hlMetadata, err := types.UnpackHLMetadata(warpPL.Metadata())
		if err != nil {
			return fmt.Errorf("unpack hl metadata: %w", err)
		}
		fmt.Printf("hl metadata: %+v\n", hlMetadata)

		if len(hlMetadata.HookForwardToIbc) > 0 {
			m, err := types.UnpackForwardToIBC(hlMetadata.HookForwardToIbc)
			if err != nil {
				return fmt.Errorf("unpack ibc forward: %w", err)
			}
			fmt.Printf("ibc forward: %+v\n", m)
		}

		if len(hlMetadata.HookForwardToHl) > 0 {
			m, err := types.UnpackForwardToHL(hlMetadata.HookForwardToHl)
			if err != nil {
				return fmt.Errorf("unpack hl forward: %w", err)
			}
			fmt.Printf("hl forward: %+v\n", m)
		}
	}

	fmt.Printf("warp payload message: %+v\n", warpPL)
	fmt.Printf("cosmos account: %s\n", warpPL.GetCosmosAccount().String())
	fmt.Printf("amount: %s\n", warpPL.Amount().String())

	return nil
}

func runEstimateFees(cmd *cobra.Command, args []string) error {
	hlAmountS, _ := cmd.Flags().GetString(FlagHLAmount)
	hlGasS, _ := cmd.Flags().GetString(FlagHLGas)
	eibcFeeS, _ := cmd.Flags().GetString(FlagEIBCFee)
	bridgeFeeMulS, _ := cmd.Flags().GetString(FlagBridgeFeeMul)

	hlReceiveAmt, ok := math.NewIntFromString(hlAmountS)
	if !ok {
		return fmt.Errorf("invalid hl amount")
	}

	hlMaxGas, ok := math.NewIntFromString(hlGasS)
	if !ok {
		return fmt.Errorf("invalid hl gas")
	}

	eibcFee, ok := math.NewIntFromString(eibcFeeS)
	if !ok {
		return fmt.Errorf("invalid eibc fee")
	}

	bridgeFeeMul, err := math.LegacyNewDecFromStr(bridgeFeeMulS)
	if err != nil {
		return fmt.Errorf("invalid bridge fee multiplier: %w", err)
	}

	needForHl := hlReceiveAmt.Add(hlMaxGas)
	transferAmt := eibctypes.CalcTargetPriceAmt(needForHl, eibcFee, bridgeFeeMul)

	fmt.Print(transferAmt)
	return nil
}

func MakeForwardToIBCHyperlaneMessage(
	hyperlaneNonce uint32,
	hyperlaneSrcDomain uint32,
	hyperlaneSrcContract util.HexAddress,
	hyperlaneDstDomain uint32,
	hyperlaneTokenID util.HexAddress,
	fundsRecipient sdk.AccAddress,
	hyperlaneTokenAmt math.Int,
	hook *types.HookForwardToIBC,
) (util.HyperlaneMessage, error) {
	if err := hook.ValidateBasic(); err != nil {
		return util.HyperlaneMessage{}, errorsmod.Wrap(err, "hook validate basic")
	}

	memoBz, err := proto.Marshal(hook)
	if err != nil {
		return util.HyperlaneMessage{}, errorsmod.Wrap(err, "marshal memo")
	}

	hlMetadata := &types.HLMetadata{
		HookForwardToIbc: memoBz,
	}
	memoBz, err = proto.Marshal(hlMetadata)
	if err != nil {
		return util.HyperlaneMessage{}, errorsmod.Wrap(err, "marshal hl metadata")
	}

	hlM, err := createHyperlaneMessage(
		hyperlaneNonce,
		hyperlaneSrcDomain,
		hyperlaneSrcContract,
		hyperlaneDstDomain,
		hyperlaneTokenID,
		fundsRecipient,
		hyperlaneTokenAmt,
		memoBz,
	)
	if err != nil {
		return util.HyperlaneMessage{}, err
	}

	{
		s := hlM.String()
		_, err := decodeHyperlaneMessageEthHexToHyperlaneToEIBCMemo(s)
		if err != nil {
			return util.HyperlaneMessage{}, errorsmod.Wrap(err, "decode eth hex")
		}
	}

	return hlM, nil
}

func MakeForwardToHLHyperlaneMessage(
	hyperlaneNonce uint32,
	hyperlaneSrcDomain uint32,
	hyperlaneSrcContract util.HexAddress,
	hyperlaneDstDomain uint32,
	hyperlaneTokenID util.HexAddress,
	fundsRecipient sdk.AccAddress,
	hyperlaneTokenAmt math.Int,
	hook *types.HookForwardToHL,
) (util.HyperlaneMessage, error) {
	if err := hook.ValidateBasic(); err != nil {
		return util.HyperlaneMessage{}, errorsmod.Wrap(err, "hook validate basic")
	}

	memoBz, err := proto.Marshal(hook)
	if err != nil {
		return util.HyperlaneMessage{}, errorsmod.Wrap(err, "marshal memo")
	}

	hlMetadata := &types.HLMetadata{
		HookForwardToHl: memoBz,
	}
	memoBz, err = proto.Marshal(hlMetadata)
	if err != nil {
		return util.HyperlaneMessage{}, errorsmod.Wrap(err, "marshal hl metadata")
	}

	hlM, err := createHyperlaneMessage(
		hyperlaneNonce,
		hyperlaneSrcDomain,
		hyperlaneSrcContract,
		hyperlaneDstDomain,
		hyperlaneTokenID,
		fundsRecipient,
		hyperlaneTokenAmt,
		memoBz,
	)
	if err != nil {
		return util.HyperlaneMessage{}, err
	}

	return hlM, nil
}

func EthRecipient(addr string) (string, error) {
	bz, err := sdk.GetFromBech32(addr, sdk.GetConfig().GetBech32AccountAddrPrefix())
	if err != nil {
		return "", fmt.Errorf("addr address from bech32: %w", err)
	}

	ret := util.EncodeEthHex(bz)
	ret = strings.TrimPrefix(ret, "0x")

	// https://github.com/dymensionxyz/hyperlane-cosmos/blob/bed4e0313aeddb8d2c59ab98d5e805955e88b300/util/hex_address.go#L29-L30
	prefix := "0x000000000000000000000000"
	ret = prefix + ret

	return ret, nil
}

func createHyperlaneMessage(
	nonce uint32,
	srcDomain uint32,
	srcContract util.HexAddress,
	dstDomain uint32,
	tokenID util.HexAddress,
	fundsRecipient sdk.AccAddress,
	amt math.Int,
	memo []byte,
) (util.HyperlaneMessage, error) {
	// Important: In Hyperlane's design, there are two different recipient concepts:
	// 1. HyperlaneMessage.Recipient = the warp router contract address (tokenID parameter)
	// 2. WarpPayload.recipient = the actual end user who receives funds (fundsRecipient parameter)

	p := sdk.GetConfig().GetBech32AccountAddrPrefix()
	bech32, err := sdk.Bech32ifyAddressBytes(p, fundsRecipient)
	if err != nil {
		return util.HyperlaneMessage{}, errorsmod.Wrap(err, "bech32ify address bytes")
	}
	recip, err := sdk.GetFromBech32(bech32, p)
	if err != nil {
		return util.HyperlaneMessage{}, errorsmod.Wrap(err, "get from bech32")
	}

	// Create warp payload with the actual funds recipient
	wmpl, err := warptypes.NewWarpPayload(recip, *big.NewInt(amt.Int64()), memo)
	if err != nil {
		return util.HyperlaneMessage{}, errorsmod.Wrap(err, "new warp payload")
	}

	body := wmpl.Bytes()
	return util.HyperlaneMessage{
		Version:     hypercoretypes.MESSAGE_VERSION,
		Nonce:       nonce,
		Origin:      srcDomain,
		Sender:      srcContract,
		Destination: dstDomain,
		Recipient:   tokenID, // This is the warp router contract address
		Body:        body,
	}, nil
}

func decodeHyperlaneMessageEthHexToHyperlaneToEIBCMemo(s string) (*types.HookForwardToIBC, error) {
	decoded, err := util.DecodeEthHex(s)
	if err != nil {
		return nil, errorsmod.Wrap(err, "decode eth hex")
	}
	warpM, err := util.ParseHyperlaneMessage(decoded)
	if err != nil {
		return nil, errorsmod.Wrap(err, "parse hl message")
	}
	pl, err := warptypes.ParseWarpPayload(warpM.Body)
	if err != nil {
		return nil, errorsmod.Wrap(err, "parse warp memo")
	}

	hlMetadata, err := types.UnpackHLMetadata(pl.Metadata())
	if err != nil {
		return nil, errorsmod.Wrap(err, "unpack hl metadata")
	}

	d, err := types.UnpackForwardToIBC(hlMetadata.HookForwardToIbc)
	if err != nil {
		return nil, errorsmod.Wrap(err, "unpack memo from hl message")
	}
	return d, nil
}
