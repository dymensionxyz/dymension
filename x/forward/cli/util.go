package cli

import (
	"fmt"
	"math/big"
	"strconv"
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

// Flag constants
const (
	// Common flags
	FlagReadable = "readable"
	FlagSource   = "source"
	FlagDest     = "dest"

	// Token/Amount flags
	FlagTokenID = "token-id"
	FlagAmount  = "amount"
	FlagMaxFee  = "max-fee"

	// Domain flags
	FlagDomain     = "domain"
	FlagSrcDomain  = "src-domain"
	FlagDestDomain = "dest-domain"
	FlagHubDomain  = "hub-domain"
	FlagKasDomain  = "kas-domain"

	// Recipient flags
	FlagRecipient     = "recipient"
	FlagDestRecipient = "dest-recipient"
	FlagHubRecipient  = "hub-recipient"

	// IBC specific flags
	FlagChannel = "channel"
	FlagTimeout = "timeout"

	// EIBC specific flags
	FlagEIBCFee = "eibc-fee"

	// Hyperlane specific flags
	FlagNonce        = "nonce"
	FlagSrcContract  = "src-contract"
	FlagDestTokenID  = "dest-token-id"
	FlagDestAmount   = "dest-amount"
	FlagRecoveryAddr = "recovery-address"

	// Kaspa specific flags
	FlagKasToken = "kas-token"

	// Decode specific
	FlagDecodeMemo = "decode-memo"
)

// Source/Destination types
const (
	SourceIBC   = "ibc"
	SourceEIBC  = "eibc"
	SourceHL    = "hl"
	SourceKaspa = "kaspa"

	DestIBC = "ibc"
	DestHL  = "hl"
	DestHub = "hub"
)

// Parameter structs
type CommonParams struct {
	Readable bool
	Source   string
	Dest     string
}

type TokenParams struct {
	TokenID   util.HexAddress
	Amount    math.Int
	Recipient sdk.AccAddress
}

type HyperlaneParams struct {
	TokenID     util.HexAddress
	Amount      math.Int
	Recipient   util.HexAddress
	Domain      uint32
	MaxFee      sdk.Coin
	Nonce       uint32
	SrcDomain   uint32
	SrcContract util.HexAddress
	DstDomain   uint32
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
	HubRecipient     sdk.AccAddress
}

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Utility commands for %s logic", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	// New unified commands
	cmd.AddCommand(CmdCreateMemo())
	cmd.AddCommand(CmdCreateHLMessage())
	cmd.AddCommand(CmdDecodeHL())
	cmd.AddCommand(CmdEthRecipient())
	cmd.AddCommand(CmdEstimateFees())

	// Keep old commands for backward compatibility (can add deprecation warnings)
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

// Unified memo creation command
func CmdCreateMemo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-memo",
		Short: "Create a memo for forwarding between different protocols",
		Long: `Create a memo for forwarding tokens between IBC, EIBC, and Hyperlane.
The source and destination determine which type of memo is created.`,
		Example: `# IBC to Hyperlane
dymd q forward create-memo --source=ibc --dest=hl \
  --token-id=0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0 \
  --dest-domain=1 \
  --dest-recipient=0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0 \
  --amount=10000 \
  --max-fee=20foo

# EIBC to IBC
dymd q forward create-memo --source=eibc --dest=ibc \
  --eibc-fee=100 \
  --channel=channel-0 \
  --recipient=ethm1a30y0h95a7p38plnv5s02lzrgcy0m0xumq0ymn \
  --timeout=5m`,
		RunE: runCreateMemo,
	}

	// Add all relevant flags
	addCommonFlags(cmd)
	addTokenFlags(cmd)
	addHyperlaneFlags(cmd)
	addIBCFlags(cmd)
	cmd.Flags().String(FlagEIBCFee, "", "EIBC fee (required for EIBC source)")

	return cmd
}

// Unified hyperlane message creation command
func CmdCreateHLMessage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-hl-message",
		Short: "Create a Hyperlane message for testing",
		Long: `Create a Hyperlane message from various sources (HL, Kaspa) to various destinations (Hub, IBC, HL).
The source and destination determine the message format and required parameters.`,
		Example: `# Kaspa to IBC
dymd q forward create-hl-message --source=kaspa --dest=ibc \
  --token-id=0x0000000000000000000000000000000000000000000000000000000000000000 \
  --hub-recipient=dym139mq752delxv78jvtmwxhasyrycufsvrw4aka9 \
  --amount=1000000000000000000 \
  --kas-token=0x0000000000000000000000000000000000000000000000000000000000000000 \
  --kas-domain=80808082 \
  --hub-domain=1260813472 \
  --channel=channel-0 \
  --dest-recipient=ethm1a30y0h95a7p38plnv5s02lzrgcy0m0xumq0ymn \
  --timeout=5m`,
		RunE: runCreateHLMessage,
	}

	// Add all relevant flags
	addCommonFlags(cmd)
	addTokenFlags(cmd)
	addHyperlaneFlags(cmd)
	addIBCFlags(cmd)
	addKaspaFlags(cmd)

	return cmd
}

// Decode hyperlane messages
func CmdDecodeHL() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "decode-hl [body|message] [hexstring]",
		Args:    cobra.ExactArgs(2),
		Short:   "Decode a Hyperlane message or body from hex string",
		Example: `dymd q forward decode-hl message 0x00000000... --decode-memo`,
		RunE:    runDecodeHL,
	}

	cmd.Flags().Bool(FlagDecodeMemo, false, "Decode the memo from the payload")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// Convert address for Ethereum
func CmdEthRecipient() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "eth-recipient [address]",
		Args:    cobra.ExactArgs(1),
		Short:   "Convert a Cosmos address to Ethereum recipient format",
		Example: `dymd q forward eth-recipient dym139mq752delxv78jvtmwxhasyrycufsvrw4aka9`,
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

// Estimate fees
func CmdEstimateFees() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "estimate-fees",
		Short:   "Estimate fees for EIBC to HL transfers",
		Example: `dymd q forward estimate-fees --hl-amount=125000000000000 --hl-gas=200000 --eibc-fee=2000 --bridge-fee-mul=0.01`,
		RunE:    runEstimateFees,
	}

	cmd.Flags().String("hl-amount", "", "Amount to receive on Hyperlane")
	cmd.Flags().String("hl-gas", "", "Max gas for Hyperlane")
	cmd.Flags().String("eibc-fee", "", "EIBC fee")
	cmd.Flags().String("bridge-fee-mul", "", "Bridge fee multiplier")
	cmd.MarkFlagRequired("hl-amount")
	cmd.MarkFlagRequired("hl-gas")
	cmd.MarkFlagRequired("eibc-fee")
	cmd.MarkFlagRequired("bridge-fee-mul")

	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

// Helper functions to add flag groups
func addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().Bool(FlagReadable, false, "Show output in human-readable format")
	cmd.Flags().String(FlagSource, "", "Source protocol (ibc, eibc, hl, kaspa)")
	cmd.Flags().String(FlagDest, "", "Destination protocol (ibc, hl, hub)")
	cmd.MarkFlagRequired(FlagSource)
	cmd.MarkFlagRequired(FlagDest)
}

func addTokenFlags(cmd *cobra.Command) {
	cmd.Flags().String(FlagTokenID, "", "Token ID (hex)")
	cmd.Flags().String(FlagAmount, "", "Token amount")
	cmd.Flags().String(FlagRecipient, "", "Recipient address")
	cmd.Flags().String(FlagMaxFee, "", "Maximum fee (e.g., 20foo)")
}

func addHyperlaneFlags(cmd *cobra.Command) {
	cmd.Flags().Uint32(FlagNonce, 0, "Message nonce")
	cmd.Flags().Uint32(FlagDomain, 0, "Domain ID")
	cmd.Flags().Uint32(FlagSrcDomain, 0, "Source domain ID")
	cmd.Flags().Uint32(FlagDestDomain, 0, "Destination domain ID")
	cmd.Flags().String(FlagSrcContract, "", "Source contract (hex)")
	cmd.Flags().String(FlagDestTokenID, "", "Destination token ID (hex)")
	cmd.Flags().String(FlagDestRecipient, "", "Destination recipient")
	cmd.Flags().String(FlagDestAmount, "", "Destination amount")
	cmd.Flags().String(FlagRecoveryAddr, "", "Recovery address")
}

func addIBCFlags(cmd *cobra.Command) {
	cmd.Flags().String(FlagChannel, "", "IBC channel")
	cmd.Flags().String(FlagTimeout, "5m", "IBC timeout duration")
}

func addKaspaFlags(cmd *cobra.Command) {
	cmd.Flags().String(FlagKasToken, "", "Kaspa token placeholder (hex)")
	cmd.Flags().Uint32(FlagKasDomain, 0, "Kaspa domain ID")
	cmd.Flags().Uint32(FlagHubDomain, 0, "Hub domain ID")
	cmd.Flags().String(FlagHubRecipient, "", "Hub recipient address")
}

// Parse helper functions
func parseCommonFlags(cmd *cobra.Command) (*CommonParams, error) {
	readable, _ := cmd.Flags().GetBool(FlagReadable)
	source, _ := cmd.Flags().GetString(FlagSource)
	dest, _ := cmd.Flags().GetString(FlagDest)

	return &CommonParams{
		Readable: readable,
		Source:   source,
		Dest:     dest,
	}, nil
}

func parseTokenFlags(cmd *cobra.Command) (*TokenParams, error) {
	tokenIDStr, _ := cmd.Flags().GetString(FlagTokenID)
	amountStr, _ := cmd.Flags().GetString(FlagAmount)
	recipientStr, _ := cmd.Flags().GetString(FlagRecipient)

	var params TokenParams
	var err error

	if tokenIDStr != "" {
		params.TokenID, err = util.DecodeHexAddress(tokenIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid token ID: %w", err)
		}
	}

	if amountStr != "" {
		var ok bool
		params.Amount, ok = math.NewIntFromString(amountStr)
		if !ok {
			return nil, fmt.Errorf("invalid amount: %s", amountStr)
		}
	}

	if recipientStr != "" {
		params.Recipient, err = sdk.AccAddressFromBech32(recipientStr)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient: %w", err)
		}
	}

	return &params, nil
}

func parseHyperlaneFlags(cmd *cobra.Command) (*HyperlaneParams, error) {
	var params HyperlaneParams
	var err error

	// Parse basic fields
	params.Nonce, _ = cmd.Flags().GetUint32(FlagNonce)
	params.Domain, _ = cmd.Flags().GetUint32(FlagDomain)
	params.SrcDomain, _ = cmd.Flags().GetUint32(FlagSrcDomain)
	params.DstDomain, _ = cmd.Flags().GetUint32(FlagDestDomain)

	// Parse hex addresses
	tokenIDStr, _ := cmd.Flags().GetString(FlagTokenID)
	if tokenIDStr != "" {
		params.TokenID, err = util.DecodeHexAddress(tokenIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid token ID: %w", err)
		}
	}

	srcContractStr, _ := cmd.Flags().GetString(FlagSrcContract)
	if srcContractStr != "" {
		params.SrcContract, err = util.DecodeHexAddress(srcContractStr)
		if err != nil {
			return nil, fmt.Errorf("invalid src contract: %w", err)
		}
	}

	recipientStr, _ := cmd.Flags().GetString(FlagDestRecipient)
	if recipientStr != "" {
		params.Recipient, err = util.DecodeHexAddress(recipientStr)
		if err != nil {
			return nil, fmt.Errorf("invalid recipient: %w", err)
		}
	}

	// Parse amount
	amountStr, _ := cmd.Flags().GetString(FlagAmount)
	if amountStr != "" {
		var ok bool
		params.Amount, ok = math.NewIntFromString(amountStr)
		if !ok {
			return nil, fmt.Errorf("invalid amount: %s", amountStr)
		}
	}

	// Parse max fee
	maxFeeStr, _ := cmd.Flags().GetString(FlagMaxFee)
	if maxFeeStr != "" {
		params.MaxFee, err = sdk.ParseCoinNormalized(maxFeeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid max fee: %w", err)
		}
	}

	return &params, nil
}

func parseIBCFlags(cmd *cobra.Command) (*IBCParams, error) {
	channel, _ := cmd.Flags().GetString(FlagChannel)
	recipient, _ := cmd.Flags().GetString(FlagRecipient)
	if recipient == "" {
		recipient, _ = cmd.Flags().GetString(FlagDestRecipient)
	}

	timeoutStr, _ := cmd.Flags().GetString(FlagTimeout)
	timeout, err := time.ParseDuration(timeoutStr)
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

	kasTokenStr, _ := cmd.Flags().GetString(FlagKasToken)
	if kasTokenStr != "" {
		params.TokenPlaceholder, err = util.DecodeHexAddress(kasTokenStr)
		if err != nil {
			return nil, fmt.Errorf("invalid kas token: %w", err)
		}
	}

	params.KasDomain, _ = cmd.Flags().GetUint32(FlagKasDomain)
	params.HubDomain, _ = cmd.Flags().GetUint32(FlagHubDomain)

	hubRecipientStr, _ := cmd.Flags().GetString(FlagHubRecipient)
	if hubRecipientStr != "" {
		params.HubRecipient, err = sdk.AccAddressFromBech32(hubRecipientStr)
		if err != nil {
			return nil, fmt.Errorf("invalid hub recipient: %w", err)
		}
	}

	return &params, nil
}

// Implementation functions for commands
func runCreateMemo(cmd *cobra.Command, args []string) error {
	common, err := parseCommonFlags(cmd)
	if err != nil {
		return err
	}

	switch common.Source {
	case SourceIBC, SourceEIBC:
		return runCreateMemoFromIBC(cmd, common)
	case SourceHL:
		return runCreateMemoFromHL(cmd, common)
	default:
		return fmt.Errorf("unsupported source: %s", common.Source)
	}
}

func runCreateMemoFromIBC(cmd *cobra.Command, common *CommonParams) error {
	var memo string

	if common.Dest == DestHL {
		// IBC/EIBC -> HL
		hlParams, err := parseHyperlaneFlags(cmd)
		if err != nil {
			return err
		}

		// Also get token params for amount
		tokenParams, err := parseTokenFlags(cmd)
		if err != nil {
			return err
		}

		hook := types.NewHookForwardToHL(
			hlParams.TokenID,
			hlParams.DstDomain,
			hlParams.Recipient,
			tokenParams.Amount,
			hlParams.MaxFee,
			math.ZeroInt(),
			nil,
			"",
		)

		if common.Source == SourceEIBC {
			eibcFeeStr, _ := cmd.Flags().GetString(FlagEIBCFee)
			memo, err = types.MakeRolForwardToHLMemoString(eibcFeeStr, hook)
		} else {
			memo, err = types.MakeIBCForwardToHLMemoString(hook)
		}

		if err != nil {
			return fmt.Errorf("create memo: %w", err)
		}

	} else if common.Dest == DestIBC {
		// IBC/EIBC -> IBC
		ibcParams, err := parseIBCFlags(cmd)
		if err != nil {
			return err
		}

		hook := types.NewHookForwardToIBC(
			ibcParams.Channel,
			ibcParams.Recipient,
			uint64(time.Now().Add(ibcParams.Timeout).UnixNano()),
		)

		if common.Source == SourceEIBC {
			eibcFeeStr, _ := cmd.Flags().GetString(FlagEIBCFee)
			memo, err = types.MakeRolForwardToIBCMemoString(eibcFeeStr, hook)
		} else {
			memo, err = types.MakeIBCForwardToIBCMemoString(hook)
		}

		if err != nil {
			return fmt.Errorf("create memo: %w", err)
		}

	} else {
		return fmt.Errorf("unsupported destination for IBC/EIBC source: %s", common.Dest)
	}

	fmt.Println(memo)
	return nil
}

func runCreateMemoFromHL(cmd *cobra.Command, common *CommonParams) error {
	if common.Dest == DestIBC {
		// HL -> IBC
		ibcParams, err := parseIBCFlags(cmd)
		if err != nil {
			return err
		}

		hook := types.NewHookForwardToIBC(
			ibcParams.Channel,
			ibcParams.Recipient,
			uint64(time.Now().Add(ibcParams.Timeout).UnixNano()),
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

	} else if common.Dest == DestHL {
		// HL -> HL
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
			hlParams.Recipient,
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

	} else {
		return fmt.Errorf("unsupported destination for HL source: %s", common.Dest)
	}

	return nil
}

func runCreateHLMessage(cmd *cobra.Command, args []string) error {
	common, err := parseCommonFlags(cmd)
	if err != nil {
		return err
	}

	if common.Source == SourceKaspa {
		return runCreateHLMessageFromKaspa(cmd, common)
	} else if common.Source == SourceHL {
		return runCreateHLMessageFromHL(cmd, common)
	} else {
		return fmt.Errorf("unsupported source: %s", common.Source)
	}
}

func runCreateHLMessageFromKaspa(cmd *cobra.Command, common *CommonParams) error {
	// Parse Kaspa specific flags
	kaspaParams, err := parseKaspaFlags(cmd)
	if err != nil {
		return err
	}

	// Parse token params
	tokenParams, err := parseTokenFlags(cmd)
	if err != nil {
		return err
	}

	// Create base message parameters
	var memo []byte

	// Handle different destinations
	if common.Dest == DestHub {
		// No memo needed for direct to hub
		memo = nil
	} else if common.Dest == DestIBC {
		// Create IBC forward hook
		ibcParams, err := parseIBCFlags(cmd)
		if err != nil {
			return err
		}

		hook := types.NewHookForwardToIBC(
			ibcParams.Channel,
			ibcParams.Recipient,
			uint64(time.Now().Add(ibcParams.Timeout).UnixNano()),
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

	} else if common.Dest == DestHL {
		// Create HL forward hook
		hlParams, err := parseHyperlaneFlags(cmd)
		if err != nil {
			return err
		}

		// Parse destination amount
		destAmountStr, _ := cmd.Flags().GetString(FlagDestAmount)
		destAmount, ok := math.NewIntFromString(destAmountStr)
		if !ok {
			return fmt.Errorf("invalid destination amount")
		}

		hook := types.NewHookForwardToHL(
			hlParams.TokenID,
			hlParams.DstDomain,
			hlParams.Recipient,
			destAmount,
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

	} else {
		return fmt.Errorf("unsupported destination: %s", common.Dest)
	}

	// Create the hyperlane message
	m, err := createHyperlaneMessage(
		0, // Fixed nonce for Kaspa
		kaspaParams.KasDomain,
		kaspaParams.TokenPlaceholder,
		kaspaParams.HubDomain,
		tokenParams.TokenID,
		kaspaParams.HubRecipient,
		tokenParams.Amount,
		memo,
	)
	if err != nil {
		return fmt.Errorf("create hyperlane message: %w", err)
	}

	// Output
	if common.Readable {
		fmt.Printf("hyperlane message: %+v\n", m)
		if common.Dest == DestIBC {
			fmt.Printf("ibc forward details: channel=%s, recipient=%s\n",
				cmd.Flag(FlagChannel).Value, cmd.Flag(FlagDestRecipient).Value)
		} else if common.Dest == DestHL {
			fmt.Printf("hl forward details: domain=%s, recipient=%s\n",
				cmd.Flag(FlagDestDomain).Value, cmd.Flag(FlagDestRecipient).Value)
		}
	} else {
		fmt.Print(m)
	}

	return nil
}

func runCreateHLMessageFromHL(cmd *cobra.Command, common *CommonParams) error {
	// Parse HL flags
	hlParams, err := parseHyperlaneFlags(cmd)
	if err != nil {
		return err
	}

	// Parse token params
	tokenParams, err := parseTokenFlags(cmd)
	if err != nil {
		return err
	}

	// Only IBC destination is supported for HL source
	if common.Dest != DestIBC {
		return fmt.Errorf("only IBC destination is supported for HL source")
	}

	// Parse IBC params
	ibcParams, err := parseIBCFlags(cmd)
	if err != nil {
		return err
	}

	hook := types.NewHookForwardToIBC(
		ibcParams.Channel,
		ibcParams.Recipient,
		uint64(time.Now().Add(ibcParams.Timeout).UnixNano()),
	)

	m, err := MakeForwardToIBCHyperlaneMessage(
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
	// Clean up input
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

		// Check for IBC forward
		if len(hlMetadata.HookForwardToIbc) > 0 {
			m, err := types.UnpackForwardToIBC(hlMetadata.HookForwardToIbc)
			if err != nil {
				return fmt.Errorf("unpack ibc forward: %w", err)
			}
			fmt.Printf("ibc forward: %+v\n", m)
		}

		// Check for HL forward
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
	hlAmountStr, _ := cmd.Flags().GetString("hl-amount")
	hlGasStr, _ := cmd.Flags().GetString("hl-gas")
	eibcFeeStr, _ := cmd.Flags().GetString("eibc-fee")
	bridgeFeeMulStr, _ := cmd.Flags().GetString("bridge-fee-mul")

	hlReceiveAmt, ok := math.NewIntFromString(hlAmountStr)
	if !ok {
		return fmt.Errorf("invalid hl amount")
	}

	hlMaxGas, ok := math.NewIntFromString(hlGasStr)
	if !ok {
		return fmt.Errorf("invalid hl gas")
	}

	eibcFee, ok := math.NewIntFromString(eibcFeeStr)
	if !ok {
		return fmt.Errorf("invalid eibc fee")
	}

	bridgeFeeMul, err := math.LegacyNewDecFromStr(bridgeFeeMulStr)
	if err != nil {
		return fmt.Errorf("invalid bridge fee multiplier: %w", err)
	}

	needForHl := hlReceiveAmt.Add(hlMaxGas)
	transferAmt := eibctypes.CalcTargetPriceAmt(needForHl, eibcFee, bridgeFeeMul)

	fmt.Print(transferAmt)
	return nil
}

// Helper functions (kept from original implementation)
func MakeForwardToIBCHyperlaneMessage(
	hyperlaneNonce uint32,
	hyperlaneSrcDomain uint32,
	hyperlaneSrcContract util.HexAddress,
	hyperlaneDstDomain uint32,
	hyperlaneTokenID util.HexAddress,
	hyperlaneRecipient sdk.AccAddress,
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
		hyperlaneRecipient,
		hyperlaneTokenAmt,
		memoBz,
	)
	if err != nil {
		return util.HyperlaneMessage{}, err
	}

	// sanity check
	{
		s := hlM.String()
		_, err := decodeHyperlaneMessageEthHexToHyperlaneToEIBCMemo(s)
		if err != nil {
			return util.HyperlaneMessage{}, errorsmod.Wrap(err, "decode eth hex")
		}
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

	// an address for eth which will be abi encoded, and then put parsed in the message
	// the first 24 bytes are for module routing, which aren't needed here, as this is inside the warp payload
	// https://github.com/dymensionxyz/hyperlane-cosmos/blob/bed4e0313aeddb8d2c59ab98d5e805955e88b300/util/hex_address.go#L29-L30
	prefix := "0x000000000000000000000000" // TODO: explain
	ret = prefix + ret

	return ret, nil
}

func createHyperlaneMessage(
	nonce uint32,
	srcDomain uint32,
	srcContract util.HexAddress,
	dstDomain uint32,
	tokenID util.HexAddress,
	recipient sdk.AccAddress,
	amt math.Int,
	memo []byte,
) (util.HyperlaneMessage, error) {
	p := sdk.GetConfig().GetBech32AccountAddrPrefix()
	bech32, err := sdk.Bech32ifyAddressBytes(p, recipient)
	if err != nil {
		return util.HyperlaneMessage{}, errorsmod.Wrap(err, "bech32ify address bytes")
	}
	recip, err := sdk.GetFromBech32(bech32, p)
	if err != nil {
		return util.HyperlaneMessage{}, errorsmod.Wrap(err, "get from bech32")
	}

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
		Recipient:   tokenID,
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

// Backward compatibility - old command implementations with deprecation warnings
func CmdMemoIBCtoHL() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "memo-ibc-to-hl [token-id] [destination-domain] [hl-recipient] [hl-amount] [max-hl-fee]",
		Args:       cobra.ExactArgs(5),
		Short:      "DEPRECATED: Use 'create-memo --source=ibc --dest=hl' instead",
		Deprecated: "use 'create-memo --source=ibc --dest=hl' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Convert positional args to new command
			newCmd := CmdCreateMemo()
			newCmd.SetArgs([]string{})
			newCmd.Flags().Set(FlagSource, SourceIBC)
			newCmd.Flags().Set(FlagDest, DestHL)
			newCmd.Flags().Set(FlagTokenID, args[0])
			newCmd.Flags().Set(FlagDestDomain, args[1])
			newCmd.Flags().Set(FlagDestRecipient, args[2])
			newCmd.Flags().Set(FlagAmount, args[3])
			newCmd.Flags().Set(FlagMaxFee, args[4])

			return runCreateMemo(newCmd, []string{})
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdMemoEIBCtoHL() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "memo-eibc-to-hl [eibc-fee] [token-id] [destination-domain] [hl-recipient] [hl-amount] [max-hl-fee]",
		Args:       cobra.ExactArgs(6),
		Short:      "DEPRECATED: Use 'create-memo --source=eibc --dest=hl' instead",
		Deprecated: "use 'create-memo --source=eibc --dest=hl' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateMemo()
			newCmd.SetArgs([]string{})
			newCmd.Flags().Set(FlagSource, SourceEIBC)
			newCmd.Flags().Set(FlagDest, DestHL)
			newCmd.Flags().Set(FlagEIBCFee, args[0])
			newCmd.Flags().Set(FlagTokenID, args[1])
			newCmd.Flags().Set(FlagDestDomain, args[2])
			newCmd.Flags().Set(FlagDestRecipient, args[3])
			newCmd.Flags().Set(FlagAmount, args[4])
			newCmd.Flags().Set(FlagMaxFee, args[5])

			return runCreateMemo(newCmd, []string{})
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdMemoIBCtoIBC() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "memo-ibc-to-ibc [ibc-source-chan] [ibc-recipient] [ibc timeout duration]",
		Args:       cobra.ExactArgs(3),
		Short:      "DEPRECATED: Use 'create-memo --source=ibc --dest=ibc' instead",
		Deprecated: "use 'create-memo --source=ibc --dest=ibc' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateMemo()
			newCmd.SetArgs([]string{})
			newCmd.Flags().Set(FlagSource, SourceIBC)
			newCmd.Flags().Set(FlagDest, DestIBC)
			newCmd.Flags().Set(FlagChannel, args[0])
			newCmd.Flags().Set(FlagRecipient, args[1])
			newCmd.Flags().Set(FlagTimeout, args[2])

			return runCreateMemo(newCmd, []string{})
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdMemoEIBCtoIBC() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "memo-eibc-to-ibc [eibc-fee] [ibc-source-chan] [ibc-recipient] [ibc timeout duration]",
		Args:       cobra.ExactArgs(4),
		Short:      "DEPRECATED: Use 'create-memo --source=eibc --dest=ibc' instead",
		Deprecated: "use 'create-memo --source=eibc --dest=ibc' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateMemo()
			newCmd.SetArgs([]string{})
			newCmd.Flags().Set(FlagSource, SourceEIBC)
			newCmd.Flags().Set(FlagDest, DestIBC)
			newCmd.Flags().Set(FlagEIBCFee, args[0])
			newCmd.Flags().Set(FlagChannel, args[1])
			newCmd.Flags().Set(FlagRecipient, args[2])
			newCmd.Flags().Set(FlagTimeout, args[3])

			return runCreateMemo(newCmd, []string{})
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdMemoHLtoIBCRaw() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "memo-hl-to-ibc [ibc-source-chan] [ibc-recipient] [ibc timeout duration]",
		Args:       cobra.ExactArgs(3),
		Short:      "DEPRECATED: Use 'create-memo --source=hl --dest=ibc' instead",
		Deprecated: "use 'create-memo --source=hl --dest=ibc' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateMemo()
			newCmd.SetArgs([]string{})
			newCmd.Flags().Set(FlagSource, SourceHL)
			newCmd.Flags().Set(FlagDest, DestIBC)
			newCmd.Flags().Set(FlagChannel, args[0])
			newCmd.Flags().Set(FlagRecipient, args[1])
			newCmd.Flags().Set(FlagTimeout, args[2])
			newCmd.Flags().Set(FlagReadable, strconv.FormatBool(cmd.Flag(FlagReadable).Changed))

			return runCreateMemo(newCmd, []string{})
		},
	}
	cmd.Flags().Bool(FlagReadable, false, "Show the message in a human readable format")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdMemoHLtoHLRaw() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "memo-hl-to-hl [token-id] [destination-domain] [hl-recipient] [hl-amount] [max-hl-fee]",
		Args:       cobra.ExactArgs(5),
		Short:      "DEPRECATED: Use 'create-memo --source=hl --dest=hl' instead",
		Deprecated: "use 'create-memo --source=hl --dest=hl' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateMemo()
			newCmd.SetArgs([]string{})
			newCmd.Flags().Set(FlagSource, SourceHL)
			newCmd.Flags().Set(FlagDest, DestHL)
			newCmd.Flags().Set(FlagTokenID, args[0])
			newCmd.Flags().Set(FlagDestDomain, args[1])
			newCmd.Flags().Set(FlagDestRecipient, args[2])
			newCmd.Flags().Set(FlagAmount, args[3])
			newCmd.Flags().Set(FlagMaxFee, args[4])
			newCmd.Flags().Set(FlagReadable, strconv.FormatBool(cmd.Flag(FlagReadable).Changed))

			return runCreateMemo(newCmd, []string{})
		},
	}
	cmd.Flags().Bool(FlagReadable, false, "Show the message in a human readable format")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdTestHLtoIBCMessage() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "hl-message [nonce] [src-domain] [src-contract] [dst-domain] [token-id] [hyperlane recipient] [amount] [ibc-source-chan] [ibc-recipient] [ibc timeout duration] [recovery-address]",
		Args:       cobra.ExactArgs(11),
		Short:      "DEPRECATED: Use 'create-hl-message --source=hl --dest=ibc' instead",
		Deprecated: "use 'create-hl-message --source=hl --dest=ibc' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateHLMessage()
			newCmd.SetArgs([]string{})
			newCmd.Flags().Set(FlagSource, SourceHL)
			newCmd.Flags().Set(FlagDest, DestIBC)
			newCmd.Flags().Set(FlagNonce, args[0])
			newCmd.Flags().Set(FlagSrcDomain, args[1])
			newCmd.Flags().Set(FlagSrcContract, args[2])
			newCmd.Flags().Set(FlagDestDomain, args[3])
			newCmd.Flags().Set(FlagTokenID, args[4])
			newCmd.Flags().Set(FlagRecipient, args[5])
			newCmd.Flags().Set(FlagAmount, args[6])
			newCmd.Flags().Set(FlagChannel, args[7])
			newCmd.Flags().Set(FlagDestRecipient, args[8])
			newCmd.Flags().Set(FlagTimeout, args[9])
			newCmd.Flags().Set(FlagRecoveryAddr, args[10])
			newCmd.Flags().Set(FlagReadable, strconv.FormatBool(cmd.Flag(FlagReadable).Changed))

			return runCreateHLMessage(newCmd, []string{})
		},
	}
	cmd.Flags().Bool(FlagReadable, false, "Show the message in a readable format")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdHLMessageKaspaToHub() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "hl-message-kaspa [token-id] [hub recipient] [amount] [kas token placeholder] [kas-domain] [hub-domain]",
		Args:       cobra.ExactArgs(6),
		Short:      "DEPRECATED: Use 'create-hl-message --source=kaspa --dest=hub' instead",
		Deprecated: "use 'create-hl-message --source=kaspa --dest=hub' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateHLMessage()
			newCmd.SetArgs([]string{})
			newCmd.Flags().Set(FlagSource, SourceKaspa)
			newCmd.Flags().Set(FlagDest, DestHub)
			newCmd.Flags().Set(FlagTokenID, args[0])
			newCmd.Flags().Set(FlagHubRecipient, args[1])
			newCmd.Flags().Set(FlagAmount, args[2])
			newCmd.Flags().Set(FlagKasToken, args[3])
			newCmd.Flags().Set(FlagKasDomain, args[4])
			newCmd.Flags().Set(FlagHubDomain, args[5])
			newCmd.Flags().Set(FlagReadable, strconv.FormatBool(cmd.Flag(FlagReadable).Changed))

			return runCreateHLMessage(newCmd, []string{})
		},
	}
	cmd.Flags().Bool(FlagReadable, false, "Show the message in a readable format")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdHLMessageKaspaToIBC() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "hl-message-kaspa-to-ibc [token-id] [hub recipient] [amount] [kas token placeholder] [kas-domain] [hub-domain] [ibc-source-chan] [ibc-recipient] [ibc timeout duration]",
		Args:       cobra.ExactArgs(9),
		Short:      "DEPRECATED: Use 'create-hl-message --source=kaspa --dest=ibc' instead",
		Deprecated: "use 'create-hl-message --source=kaspa --dest=ibc' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateHLMessage()
			newCmd.SetArgs([]string{})
			newCmd.Flags().Set(FlagSource, SourceKaspa)
			newCmd.Flags().Set(FlagDest, DestIBC)
			newCmd.Flags().Set(FlagTokenID, args[0])
			newCmd.Flags().Set(FlagHubRecipient, args[1])
			newCmd.Flags().Set(FlagAmount, args[2])
			newCmd.Flags().Set(FlagKasToken, args[3])
			newCmd.Flags().Set(FlagKasDomain, args[4])
			newCmd.Flags().Set(FlagHubDomain, args[5])
			newCmd.Flags().Set(FlagChannel, args[6])
			newCmd.Flags().Set(FlagDestRecipient, args[7])
			newCmd.Flags().Set(FlagTimeout, args[8])
			newCmd.Flags().Set(FlagReadable, strconv.FormatBool(cmd.Flag(FlagReadable).Changed))

			return runCreateHLMessage(newCmd, []string{})
		},
	}
	cmd.Flags().Bool(FlagReadable, false, "Show the message in a readable format")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdHLMessageKaspaToHL() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "hl-message-kaspa-to-hl [token-id] [hub recipient] [amount] [kas token placeholder] [kas-domain] [hub-domain] [hl-token-id] [hl-destination-domain] [hl-recipient] [hl-amount] [max-hl-fee]",
		Args:       cobra.ExactArgs(11),
		Short:      "DEPRECATED: Use 'create-hl-message --source=kaspa --dest=hl' instead",
		Deprecated: "use 'create-hl-message --source=kaspa --dest=hl' instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			newCmd := CmdCreateHLMessage()
			newCmd.SetArgs([]string{})
			newCmd.Flags().Set(FlagSource, SourceKaspa)
			newCmd.Flags().Set(FlagDest, DestHL)
			newCmd.Flags().Set(FlagTokenID, args[0])
			newCmd.Flags().Set(FlagHubRecipient, args[1])
			newCmd.Flags().Set(FlagAmount, args[2])
			newCmd.Flags().Set(FlagKasToken, args[3])
			newCmd.Flags().Set(FlagKasDomain, args[4])
			newCmd.Flags().Set(FlagHubDomain, args[5])
			newCmd.Flags().Set(FlagDestTokenID, args[6])
			newCmd.Flags().Set(FlagDestDomain, args[7])
			newCmd.Flags().Set(FlagDestRecipient, args[8])
			newCmd.Flags().Set(FlagDestAmount, args[9])
			newCmd.Flags().Set(FlagMaxFee, args[10])
			newCmd.Flags().Set(FlagReadable, strconv.FormatBool(cmd.Flag(FlagReadable).Changed))

			return runCreateHLMessage(newCmd, []string{})
		},
	}
	cmd.Flags().Bool(FlagReadable, false, "Show the message in a readable format")
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
