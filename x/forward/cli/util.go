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

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Utility commands for %s logic", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdMemoIBCtoHL())
	cmd.AddCommand(CmdMemoIBCtoIBC())
	cmd.AddCommand(CmdMemoEIBCtoHL())
	cmd.AddCommand(CmdMemoEIBCtoIBC())
	cmd.AddCommand(CmdMemoHLtoIBCRaw())
	cmd.AddCommand(CmdHLEthTransferRecipientHubAccount())
	cmd.AddCommand(CmdTestHLtoIBCMessage())
	cmd.AddCommand(CmdDecodeHyperlaneMessage())
	cmd.AddCommand(EstimateEIBCtoHLTransferAmt())

	return cmd
}

const (
	MessageReadableFlag = "readable"
	DecodeMemoFlag      = "memo"
)

// get a memo for the direction (E)IBC -> HL. This should be directly included in the memo of the ibc transfer.
func CmdMemoEIBCtoHL() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "memo-eibc-to-hl [eibc-fee] [token-id] [destination-domain] [hl-recipient] [hl-amount] [max-hl-fee]",
		Args:    cobra.ExactArgs(6),
		Short:   "Create a memo for the direction (E)IBC -> HL",
		Example: `dymd q forward memo-eibc-to-hl 100 0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0 1 0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0 10000 20foo`,

		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {
			eibcFee := args[0]
			_, err := strconv.Atoi(eibcFee)
			if err != nil {
				return fmt.Errorf("eibc fee: %w", err)
			}

			hook, err := hookForwardToHL(args[1:])
			if err != nil {
				return fmt.Errorf("hook forward to hl: %w", err)
			}

			memo, err := types.MakeRolForwardToHLMemoString(eibcFee, hook)
			if err != nil {
				return fmt.Errorf("new memo: %w", err)
			}

			fmt.Println(memo)
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// get a memo for the direction IBC -> HL. This should be directly included in the memo of the ibc transfer.
func CmdMemoIBCtoHL() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "memo-ibc-to-hl [token-id] [destination-domain] [hl-recipient] [hl-amount] [max-hl-fee]",
		Args:    cobra.ExactArgs(5),
		Short:   "Create a memo for the direction IBC -> HL",
		Example: `dymd q forward memo-ibc-to-hl 100 0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0 1 0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0 10000 20foo`,

		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {
			hook, err := hookForwardToHL(args)
			if err != nil {
				return fmt.Errorf("hook forward to hl: %w", err)
			}

			memo, err := types.MakeIBCForwardToHLMemoString(hook)
			if err != nil {
				return fmt.Errorf("new memo: %w", err)
			}
			fmt.Println(memo)
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func hookForwardToHL(args []string) (*types.HookForwardToHL, error) {
	tokenId, err := util.DecodeHexAddress(args[0])
	if err != nil {
		return nil, fmt.Errorf("token id: %w", err)
	}

	destinationDomain, err := strconv.ParseUint(args[1], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("destination domain: %w", err)
	}

	recipient, err := util.DecodeHexAddress(args[2])
	if err != nil {
		return nil, fmt.Errorf("recipient: %w", err)
	}

	amount, ok := math.NewIntFromString(args[3])
	if !ok {
		return nil, fmt.Errorf("amount")
	}

	maxFee, err := sdk.ParseCoinNormalized(args[4])
	if err != nil {
		return nil, fmt.Errorf("max fee: %w", err)
	}

	return types.NewHookForwardToHL(
		tokenId,
		uint32(destinationDomain),
		recipient,
		amount,
		maxFee,
		math.ZeroInt(), // ignored
		nil,            // ignored
		"",             // ignored
	), nil
}

// get a memo for the direction (E)IBC -> IBC
func CmdMemoEIBCtoIBC() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "memo-eibc-to-ibc [eibc-fee] [ibc-source-chan] [ibc-recipient] [ibc timeout duration]",
		Args:    cobra.ExactArgs(4),
		Short:   "Create a memo for the direction (E)IBC -> IBC",
		Example: `dymd q forward memo-eibc-to-ibc 100 "channel-0" ethm1a30y0h95a7p38plnv5s02lzrgcy0m0xumq0ymn 5m`,

		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {
			hook, err := hookForwardToIBC(args[1:])
			if err != nil {
				return fmt.Errorf("memo hl to ibc: %w", err)
			}

			memo, err := types.MakeRolForwardToIBCMemoString(args[0], hook)
			if err != nil {
				return fmt.Errorf("new memo: %w", err)
			}

			fmt.Println(memo)
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// get a memo for the direction (E)IBC -> IBC
func CmdMemoIBCtoIBC() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "memo-ibc-to-ibc [ibc-source-chan] [ibc-recipient] [ibc timeout duration]",
		Args:    cobra.ExactArgs(3),
		Short:   "Create a memo for the direction IBC -> IBC",
		Example: `dymd q forward memo-ibc-to-ibc "channel-0" ethm1a30y0h95a7p38plnv5s02lzrgcy0m0xumq0ymn 5m`,

		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {
			hook, err := hookForwardToIBC(args[1:])
			if err != nil {
				return fmt.Errorf("memo hl to ibc: %w", err)
			}

			s, err := types.MakeIBCForwardToIBCMemoString(hook)
			if err != nil {
				return fmt.Errorf("new memo: %w", err)
			}

			fmt.Println(s)
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// Get a memo for the direction HL -> (E)IBC
func CmdMemoHLtoIBCRaw() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "memo-hl-to-ibc [ibc-source-chan] [ibc-recipient] [ibc timeout duration]",
		Args:                       cobra.ExactArgs(3),
		Short:                      "Get the memo for the direction HL -> IBC or EIBC",
		Example:                    `dymd q forward memo-hl-to-ibc "channel-0" ethm1a30y0h95a7p38plnv5s02lzrgcy0m0xumq0ymn 5m`,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {
			hook, err := hookForwardToIBC(args)
			if err != nil {
				return fmt.Errorf("memo hl to ibc: %w", err)
			}

			readable, err := cmd.Flags().GetBool(MessageReadableFlag)
			if err != nil {
				return fmt.Errorf("encode flag: %w", err)
			}

			if readable {
				fmt.Printf("hyperlane message: %+v\n", hook)
			} else {
				bz, err := proto.Marshal(hook)
				if err != nil {
					return fmt.Errorf("marshal: %w", err)
				}
				fmt.Printf("%s\n", util.EncodeEthHex(bz))
			}

			return nil
		},
	}

	cmd.Flags().Bool(MessageReadableFlag, false, "Show the message in a human readable format (for debug)")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func hookForwardToIBC(args []string) (*types.HookForwardToIBC, error) {
	ibcSourceChan := args[0]

	ibcRecipient := args[1]

	ibcTimeoutDuration, err := time.ParseDuration(args[2])
	if err != nil {
		return nil, fmt.Errorf("ibc timeout duration: %w", err)
	}

	ibcTimeoutTimestamp := uint64(time.Now().Add(ibcTimeoutDuration).UnixNano())

	hook := types.NewHookForwardToIBC(
		ibcSourceChan,
		ibcRecipient,
		ibcTimeoutTimestamp,
	)
	err = hook.ValidateBasic()
	if err != nil {
		return nil, fmt.Errorf("new hl message: %w", err)
	}
	return hook, nil
}

// Get a message for the direction HL -> (E)IBC. Intended for testing (check that the hub handles inbound messages correctly.)
func CmdTestHLtoIBCMessage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hl-message [nonce] [src-domain] [src-contract] [dst-domain] [token-id] [hyperlane recipient] [amount] [ibc-source-chan] [ibc-recipient] [ibc timeout duration] [recovery-address]",
		Args:  cobra.ExactArgs(11),
		Short: "Create a hyperlane message for testing Hl -> IBC",
		Example: `
		dymd q forward hl-message 1 1
0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0 1 0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0
dym139mq752delxv78jvtmwxhasyrycufsvrw4aka9 50 channel-0 ethm1wqg8227q0p7pgp7lj7z6cu036l6eg34d9cp6lk 5m
dym1yecvrgz7yp26keaxa4r00554uugatxfegk76hz`,

		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {
			hlNonce, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return fmt.Errorf("nonce: %w", err)
			}

			hlSrcDomain, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("counterparty domain: %w", err)
			}

			hlSrcContract, err := util.DecodeHexAddress(args[2])
			if err != nil {
				return fmt.Errorf("counterparty contract: %w", err)
			}

			hlDstDomain, err := strconv.ParseUint(args[3], 10, 32)
			if err != nil {
				return fmt.Errorf("local domain: %w", err)
			}

			hlTokenID, err := util.DecodeHexAddress(args[4])
			if err != nil {
				return fmt.Errorf("token id: %w", err)
			}

			hlRecipient, err := sdk.AccAddressFromBech32(args[5])
			if err != nil {
				return fmt.Errorf("recipient address: %w", err)
			}

			hlAmt, ok := math.NewIntFromString(args[6])
			if !ok {
				return fmt.Errorf("amount")
			}

			hook, err := hookForwardToIBC(args[7:])
			if err != nil {
				return fmt.Errorf("memo hl to ibc: %w", err)
			}

			readable, err := cmd.Flags().GetBool(MessageReadableFlag)
			if err != nil {
				return fmt.Errorf("encode flag: %w", err)
			}

			m, err := MakeForwardToIBCHyperlaneMessage(
				uint32(hlNonce),
				uint32(hlSrcDomain),
				hlSrcContract,
				uint32(hlDstDomain),
				hlTokenID,
				hlRecipient,
				hlAmt,
				hook,
			)
			if err != nil {
				return fmt.Errorf("new hl message: %w", err)
			}

			if readable {
				fmt.Printf("hyperlane message: %+v\n", m)
			} else {
				fmt.Print(m) // encodes with .String()
			}

			return nil
		},
	}

	cmd.Flags().Bool(MessageReadableFlag, false, "Show the message in a readable format")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// Get a value to pass to Ethereum as the recipient of a token transfer on the hub
func CmdHLEthTransferRecipientHubAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "hl-eth-recipient [addr]",
		Args:                       cobra.ExactArgs(1),
		Short:                      "",
		Example:                    `dymd q forward eth-hub-recipient dym139mq752delxv78jvtmwxhasyrycufsvrw4aka9`,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, err := EthRecipient(args[0])
			if err != nil {
				return fmt.Errorf("hl eth addr: %w", err)
			}

			fmt.Println(addr)

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// addr like dym166kyzqc2e0ewmwmv4vj68pzqp57tgts5lyawlc
// returns a value which can be passed to ethereum as recipient, e.g 0x000000000000000000000000d6ac41030acbf2edbb6cab25a384400d3cb42e14
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

// Util to debug a hyperlane message or hyperlane body (including show the memo if there is one). Expects Ethereum Hex bytes
func CmdDecodeHyperlaneMessage() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "hl-decode (body | message) [hexstring]",
		Args:                       cobra.ExactArgs(2),
		Short:                      "Decode a message or message body from an hex string",
		Long:                       "Provide a HL message or message body string in hex form and see what it decodes to",
		Example:                    `dymd q forward hl-decode message 0x00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000`,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {
			kind := args[0]
			if kind != "body" && kind != "message" {
				return fmt.Errorf("unsupported message type: %s", kind)
			}
			d := args[1]

			// be a bit kinder to what can be passed into the terminal
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
				var message util.HyperlaneMessage
				m, err := util.ParseHyperlaneMessage(bz)
				if err != nil {
					return fmt.Errorf(" %w", err)
				}
				body = m.Body
				message = m
				fmt.Printf("hyperlane message: %+v\n", message)
			}

			memo, err := cmd.Flags().GetBool(DecodeMemoFlag)
			if err != nil {
				return fmt.Errorf("encode flag: %w", err)
			}

			warpPL, err := warptypes.ParseWarpPayload(body)
			if err != nil {
				return fmt.Errorf("parse warp payload: %w", err)
			}
			if memo {
				m, err := types.UnpackForwardToIBC(warpPL.Metadata())
				if err != nil {
					return fmt.Errorf("unpack memo from warp message: %w", err)
				}
				fmt.Printf("ibc memo: %+v\n", m)
			}
			fmt.Printf("warp payload message: %+v\n", warpPL)
			fmt.Printf("cosmos account: %s\n", warpPL.GetCosmosAccount().String())
			fmt.Printf("amount: %s\n", warpPL.Amount().String())
			return nil
		},
	}

	cmd.Flags().Bool(DecodeMemoFlag, false, "Decode the memo from the payload")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func EstimateEIBCtoHLTransferAmt() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "amt-eibc-to-hl [hl receive amt] [hl max gas] [eibc fee] [bridge fee mul]",
		Args:                       cobra.ExactArgs(4),
		Short:                      "Get amount of tokens for ibc transfer to ensure enough arrive on final destination after fees and HL",
		Long:                       "Estimate the amount of tokens to send over EIBC to be forwarded to HL and to make sure to receive the specified amount on the final destination",
		Example:                    `dymd q forward amt-eibc-to-hl 125000000000000 200000 2000 0.01`,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {
			hlReceiveAmt, ok := math.NewIntFromString(args[0])
			if !ok {
				return fmt.Errorf("hl receive amt")
			}

			hlMaxGas, ok := math.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("max gas")
			}

			eibcFee, ok := math.NewIntFromString(args[2])
			if !ok {
				return fmt.Errorf("eibc fee")
			}

			bridgeFeeMul, err := math.LegacyNewDecFromStr(args[3])
			if err != nil {
				return fmt.Errorf("fee mul: %w", err)
			}

			// price calculation always includes the bridge fee, so we don't need to do another calculation for the finalize case

			needForHl := hlReceiveAmt.Add(hlMaxGas)

			transferAmt := eibctypes.CalcTargetPriceAmt(needForHl, eibcFee, bridgeFeeMul)

			fmt.Print(transferAmt)
			return nil
		},
	}
	return cmd
}

// get a message for sending directly to hyperlane module on hub
// for testing
// potentially computationally expensive
func MakeForwardToIBCHyperlaneMessage(
	hyperlaneNonce uint32,
	hyperlaneSrcDomain uint32, // e.g. 1 for Ethereum
	hyperlaneSrcContract util.HexAddress, // e.g. Ethereum token contract as defined in token remote router
	hyperlaneDstDomain uint32, // e.g. 0 for Dymension
	hyperlaneTokenID util.HexAddress,
	hyperlaneRecipient sdk.AccAddress, // hub account to get the tokens
	hyperlaneTokenAmt math.Int, // must be at least hub token amount
	hook *types.HookForwardToIBC,
) (util.HyperlaneMessage, error) {
	if err := hook.ValidateBasic(); err != nil {
		return util.HyperlaneMessage{}, errorsmod.Wrap(err, "validate basic")
	}

	memoBz, err := proto.Marshal(hook)
	if err != nil {
		return util.HyperlaneMessage{}, err
	}

	hlM, err := createTestHyperlaneMessage(
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
		return util.HyperlaneMessage{}, err
	}

	// sanity
	{
		s := hlM.String()
		_, err := decodeHyperlaneMessageEthHexToHyperlaneToEIBCMemo(s)
		if err != nil {
			return util.HyperlaneMessage{}, errorsmod.Wrap(err, "decode eth hex")
		}
	}

	return hlM, nil
}

// A message which can be sent to the mailbox in TX to trigger a transfer
func createTestHyperlaneMessage(
	version uint8, // e.g. 1
	nonce uint32, // e.g. 1
	srcDomain uint32, // e.g. 1 (Ethereum)
	srcContract util.HexAddress, // e.g Ethereum token contract
	dstDomain uint32, // e.g. 0 (Dymension)
	tokenID util.HexAddress,
	recipient sdk.AccAddress,
	amt math.Int,
	memo []byte,
) (util.HyperlaneMessage, error) {
	p := sdk.GetConfig().GetBech32AccountAddrPrefix()
	bech32, err := sdk.Bech32ifyAddressBytes(p, recipient) // TODO: fix
	if err != nil {
		return util.HyperlaneMessage{}, err
	}
	recip, err := sdk.GetFromBech32(bech32, p) // TODO: fix
	if err != nil {
		return util.HyperlaneMessage{}, err
	}

	wmpl, err := warptypes.NewWarpPayload(recip, *big.NewInt(amt.Int64()), memo)
	if err != nil {
		return util.HyperlaneMessage{}, err
	}

	body := wmpl.Bytes()
	return util.HyperlaneMessage{
		Version:     version,
		Nonce:       nonce,
		Origin:      srcDomain,
		Sender:      srcContract,
		Destination: dstDomain,
		Recipient:   tokenID,
		Body:        body,
	}, nil
}

// intended for tests/clients, expensive
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
	d, err := types.UnpackForwardToIBC(pl.Metadata())
	if err != nil {
		return nil, errorsmod.Wrap(err, "unpack memo from hl message")
	}
	return d, nil
}
