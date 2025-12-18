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

	FlagHLAmount       = "hl-amount"
	FlagHLGas          = "hl-gas"
	FlagBridgeFeeMul   = "bridge-fee-mul"
	FlagProtocolFeeMul = "protocol-fee-mul"

	FlagNonce       = "nonce"
	FlagSrcContract = "src-contract"
	FlagDstTokenID  = "dst-token-id" // #nosec G101 - This is a CLI flag name, not a credential
	FlagDstAmount   = "dst-amount"

	FlagKasToken = "kas-token"

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
	cmd.Flags().String(FlagEIBCFee, "", "EIBC fee amount for fast finality (required for EIBC source, paid to order fulfiller)")

	return cmd
}

func CmdCreateHLMessage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-hl-message",
		Short: "Create a Hyperlane message for testing or for including in Kaspa payloads or Solana/EVM chain outbound transfers",
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
		Use:   "estimate-fees",
		Short: "Estimate fees for EIBC to HL transfers",
		Long: `Calculate the total IBC transfer amount needed when sending tokens from RollApp through EIBC to Hyperlane.
		
The protocol fee multiplier is a percentage fee taken by the protocol (e.g., 0.01 = 1% fee).
This fee is applied to the Hyperlane amount solely. It is optional and set to 0 if not provided.

The bridge fee multiplier is a percentage fee taken by the bridge operator (e.g., 0.01 = 1% fee).
This fee is applied to the sum of the Hyperlane amount and gas to calculate the total transfer amount.

Formula: transfer_amount = (hl_amount * (1 + protocol_fee) + hl_gas) * (1 + bridge_fee_mul) + eibc_fee`,
		Example: `# Calculate fees with 1% bridge fee (0.01), 2000 EIBC fee
dymd q forward estimate-fees --hl-amount=125000000000000 --hl-gas=200000 --eibc-fee=2000 --bridge-fee-mul=0.01 --protocol-fee-mul=0.01

# With 0.5% bridge fee and 0% protocol fee
dymd q forward estimate-fees --hl-amount=1000000 --hl-gas=50000 --eibc-fee=100 --bridge-fee-mul=0.005`,
		RunE: runEstimateFees,
	}

	cmd.Flags().String(FlagHLAmount, "", "Amount to receive on Hyperlane destination chain")
	cmd.Flags().String(FlagHLGas, "", "Maximum gas fee for Hyperlane execution")
	cmd.Flags().String(FlagEIBCFee, "", "EIBC fee for fast finality (paid to fulfiller)")
	cmd.Flags().String(FlagBridgeFeeMul, "", "Bridge fee multiplier as decimal (e.g., 0.01 for 1% fee)")
	cmd.Flags().String(FlagProtocolFeeMul, "0", "Protocol fee multiplier as decimal (e.g., 0.01 for 1% fee)")
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
	cmd.Flags().String(FlagTokenID, "", "Token ID in hex format (32 bytes, e.g., 0x726f757465725f61707000000000000000000000000000010000000000000005)")
	cmd.Flags().String(FlagAmount, "", "Token amount to transfer (in smallest unit)")
	cmd.Flags().String(FlagMaxFee, "", "Maximum fee for Hyperlane execution including denom (e.g., 20ibc/ABC123...)")
}

func addHyperlaneFlags(cmd *cobra.Command) {
	cmd.Flags().Uint32(FlagNonce, 0, "Message nonce for ordering/uniqueness")
	cmd.Flags().Uint32(FlagDomain, 0, "Domain ID (deprecated, use --dst-domain)")
	cmd.Flags().Uint32(FlagSrcDomain, 0, "Source chain domain ID (e.g., 1260813472 for Dymension Hub)")
	cmd.Flags().Uint32(FlagDstDomain, 0, "Destination chain domain ID (e.g., 11155111 for Ethereum Sepolia)")
	cmd.Flags().String(FlagSrcContract, "", "Source contract address in hex format")
	cmd.Flags().String(FlagDstTokenID, "", "Destination token ID in hex format")
	cmd.Flags().String(FlagRecipientDst, "", "Final recipient address on destination chain (pad to 32 bytes for Ethereum: 0x000...)")
	cmd.Flags().String(FlagDstAmount, "", "Amount to be received on destination after fees")
}

func addIBCFlags(cmd *cobra.Command) {
	cmd.Flags().String(FlagChannel, "", "IBC channel ID (e.g., channel-0)")
	cmd.Flags().String(FlagTimeout, "5m", "IBC packet timeout duration (e.g., 5m, 1h, 30s)")
}

func addKaspaFlags(cmd *cobra.Command) {
	cmd.Flags().String(FlagKasToken, "", "Kaspa token placeholder in hex format")
	cmd.Flags().Uint32(FlagKasDomain, 0, "Kaspa network domain ID (e.g., 80808082)")
	cmd.Flags().Uint32(FlagHubDomain, 0, "Dymension Hub domain ID (e.g., 1260813472)")
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
	return parseTokenFlagsWithContext(cmd, false)
}

func parseTokenFlagsWithContext(cmd *cobra.Command, skipRecipient bool) (*TokenParams, error) {
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

	// Skip recipient parsing when it will be handled by parseHyperlaneFlags
	// This occurs when forwarding from IBC/EIBC to Hyperlane destinations
	if recipientS != "" && !skipRecipient {
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

		// Skip recipient parsing since it's handled by parseHyperlaneFlags
		tokenParams, err := parseTokenFlagsWithContext(cmd, true)
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

		// Validate the created hook to ensure all required fields are populated
		if err := validateHookForwardToHL(hook); err != nil {
			return fmt.Errorf("invalid Hyperlane forward hook: %w", err)
		}

		if common.Src == SrcEIBC {
			eibcFeeS, _ := cmd.Flags().GetString(FlagEIBCFee)
			if eibcFeeS == "" {
				return fmt.Errorf("EIBC fee is required when forwarding from EIBC to Hyperlane")
			}
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

		// Validate the created hook to ensure all required fields are populated
		if err := validateHookForwardToIBC(hook); err != nil {
			return fmt.Errorf("invalid IBC forward hook: %w", err)
		}

		if common.Src == SrcEIBC {
			eibcFeeS, _ := cmd.Flags().GetString(FlagEIBCFee)
			if eibcFeeS == "" {
				return fmt.Errorf("EIBC fee is required when forwarding from EIBC to IBC")
			}
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
			uint64(time.Now().Add(ibcParams.Timeout).UnixNano()), // #nosec G115 - Unix time is always positive
		)

		// Validate the created hook to ensure all required fields are populated
		if err := validateHookForwardToIBC(hook); err != nil {
			return fmt.Errorf("invalid IBC forward hook: %w", err)
		}

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

		// Skip recipient parsing since it's handled by parseHyperlaneFlags
		tokenParams, err := parseTokenFlagsWithContext(cmd, true)
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

	tokenParams, err := parseTokenFlagsWithContext(cmd, true)
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

func runEstimateFees(cmd *cobra.Command, args []string) error {
	hlAmountS, _ := cmd.Flags().GetString(FlagHLAmount)
	hlGasS, _ := cmd.Flags().GetString(FlagHLGas)
	eibcFeeS, _ := cmd.Flags().GetString(FlagEIBCFee)
	bridgeFeeMulS, _ := cmd.Flags().GetString(FlagBridgeFeeMul)
	protocolFeeMulS, _ := cmd.Flags().GetString(FlagProtocolFeeMul)

	hlReceiveAmt, ok := math.NewIntFromString(hlAmountS)
	if !ok {
		return fmt.Errorf("invalid hl amount: %s", hlAmountS)
	}
	if hlReceiveAmt.IsNegative() {
		return fmt.Errorf("hl amount cannot be negative")
	}

	hlMaxGas, ok := math.NewIntFromString(hlGasS)
	if !ok {
		return fmt.Errorf("invalid hl gas: %s", hlGasS)
	}
	if hlMaxGas.IsNegative() {
		return fmt.Errorf("hl gas cannot be negative")
	}

	eibcFee, ok := math.NewIntFromString(eibcFeeS)
	if !ok {
		return fmt.Errorf("invalid eibc fee: %s", eibcFeeS)
	}
	if eibcFee.IsNegative() {
		return fmt.Errorf("eibc fee cannot be negative")
	}

	bridgeFeeMul, err := math.LegacyNewDecFromStr(bridgeFeeMulS)
	if err != nil {
		return fmt.Errorf("invalid bridge fee multiplier: %w", err)
	}
	if bridgeFeeMul.IsNegative() {
		return fmt.Errorf("bridge fee multiplier cannot be negative")
	}
	if bridgeFeeMul.GTE(math.LegacyNewDec(1)) {
		return fmt.Errorf("bridge fee multiplier must be less than 1 (100%%), got %s", bridgeFeeMulS)
	}

	protocolFeeMul, err := math.LegacyNewDecFromStr(protocolFeeMulS)
	if err != nil {
		return fmt.Errorf("invalid protocol fee multiplier: %w", err)
	}
	if protocolFeeMul.IsNegative() {
		return fmt.Errorf("protocol fee multiplier cannot be negative")
	}
	if protocolFeeMul.GTE(math.LegacyNewDec(1)) {
		return fmt.Errorf("protocol fee multiplier must be less than 1 (100%%), got %s", protocolFeeMulS)
	}

	hlReceivePlusProtocolFee := protocolFeeMul.Add(math.LegacyNewDec(1)).MulInt(hlReceiveAmt).TruncateInt()
	needForHl := hlReceivePlusProtocolFee.Add(hlMaxGas)
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

// validateHookForwardToIBC validates that all required fields for IBC forwarding are populated
func validateHookForwardToIBC(hook *types.HookForwardToIBC) error {
	if hook == nil {
		return fmt.Errorf("hook is nil")
	}

	// The hook's ValidateBasic will check if Transfer is nil and call Transfer.ValidateBasic()
	// which validates channel format, receiver format, etc.
	// We only need to check for empty values that ValidateBasic doesn't catch

	if hook.Transfer != nil {
		// MsgTransfer.ValidateBasic checks channel format but allows empty strings
		if hook.Transfer.SourceChannel == "" {
			return fmt.Errorf("channel is required for IBC forwarding")
		}

		// MsgTransfer.ValidateBasic checks receiver is not empty (strings.TrimSpace(msg.Receiver) == "")
		// so we don't need to duplicate that check

		if hook.Transfer.TimeoutTimestamp == 0 {
			return fmt.Errorf("timeout is required for IBC forwarding")
		}
	}

	// Run the built-in validation which handles most checks
	if err := hook.ValidateBasic(); err != nil {
		return fmt.Errorf("hook validation failed: %w", err)
	}

	return nil
}

// validateHookForwardToHL validates that all required fields for Hyperlane forwarding are populated
func validateHookForwardToHL(hook *types.HookForwardToHL) error {
	if hook == nil {
		return fmt.Errorf("hook is nil")
	}

	// The hook's ValidateBasic only checks if HyperlaneTransfer is nil
	// MsgRemoteTransfer doesn't have a ValidateBasic, so we need to check all fields ourselves

	if hook.HyperlaneTransfer != nil {
		// TokenId is a fixed-size array, so check if all bytes are zero
		if isZeroHexAddress(hook.HyperlaneTransfer.TokenId) {
			return fmt.Errorf("token ID is required for Hyperlane forwarding")
		}

		if hook.HyperlaneTransfer.DestinationDomain == 0 {
			return fmt.Errorf("destination domain is required for Hyperlane forwarding")
		}

		// Recipient is also a fixed-size array, check if all bytes are zero
		if isZeroHexAddress(hook.HyperlaneTransfer.Recipient) {
			return fmt.Errorf("recipient address is required for Hyperlane forwarding")
		}

		if hook.HyperlaneTransfer.Amount.IsNil() || hook.HyperlaneTransfer.Amount.IsZero() {
			return fmt.Errorf("amount must be greater than zero for Hyperlane forwarding")
		}

		if hook.HyperlaneTransfer.MaxFee.IsNil() || hook.HyperlaneTransfer.MaxFee.IsZero() {
			return fmt.Errorf("max fee must be greater than zero for Hyperlane forwarding")
		}
	}

	// Run the built-in validation (only checks HyperlaneTransfer != nil)
	if err := hook.ValidateBasic(); err != nil {
		return fmt.Errorf("hook validation failed: %w", err)
	}

	return nil
}

// isZeroHexAddress checks if a HexAddress (32-byte array) is all zeros
func isZeroHexAddress(addr util.HexAddress) bool {
	for _, b := range addr {
		if b != 0 {
			return false
		}
	}
	return true
}
