package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	math "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
	hyperutil "github.com/dymensionxyz/hyperlane-cosmos/util"
	warptypes "github.com/dymensionxyz/hyperlane-cosmos/x/warp/types"
)

func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Utility commands for %s logic", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdMemoEIBCtoHL())
	cmd.AddCommand(CmdMemoEIBCtoIBC())
	cmd.AddCommand(CmdMemoHLtoEIBCRaw())
	cmd.AddCommand(CmdTestHLtoIBCMessage())
	cmd.AddCommand(CmdDecodeHyperlaneMessage())
	cmd.AddCommand(EstimateEIBCtoHLTransferAmt())

	return cmd
}

const (
	MessageReadableFlag = "readable"
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

			tokenId, err := hyperutil.DecodeHexAddress(args[1])
			if err != nil {
				return fmt.Errorf("token id: %w", err)
			}

			destinationDomain, err := strconv.ParseUint(args[2], 10, 32)
			if err != nil {
				return fmt.Errorf("destination domain: %w", err)
			}

			recipient, err := hyperutil.DecodeHexAddress(args[3])
			if err != nil {
				return fmt.Errorf("recipient: %w", err)
			}

			amount, ok := math.NewIntFromString(args[4])
			if !ok {
				return fmt.Errorf("amount")
			}

			maxFee, err := sdk.ParseCoinNormalized(args[5])
			if err != nil {
				return fmt.Errorf("max fee: %w", err)
			}

			memo, err := types.NewRollToHLMemoString(
				eibcFee,
				tokenId,
				uint32(destinationDomain),
				recipient,
				amount,
				maxFee,
				math.ZeroInt(), // ignored
				nil,            // ignored
				"",             // ignored
			)
			if err != nil {
				return fmt.Errorf("new: %w", err)
			}

			fmt.Println(memo)
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// get a memo for the direction (E)IBC -> IBC
func CmdMemoEIBCtoIBC() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "memo-eibc-to-ibc [eibc-fee] [ibc-source-chan] [ibc-recipient] [hub-token] [ibc timeout duration]",
		Args:    cobra.ExactArgs(5),
		Short:   "Create a memo for the direction (E)IBC -> IBC",
		Example: `dymd q forward memo-eibc-to-ibc 100 "channel-0" ethm1a30y0h95a7p38plnv5s02lzrgcy0m0xumq0ymn 100ibc/9A1EACD53A6A197ADC81DF9A49F0C4A26F7FF685ACF415EE726D7D59796E71A7 5m`,

		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {

			data, err := memoForwardToIBC(args[1:])
			if err != nil {
				return fmt.Errorf("memo hl to ibc: %w", err)
			}

			memo, err := types.NewRollToIBCMemoString(args[0], data)
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

// Get a memo for the direction HL -> (E)IBC
func CmdMemoHLtoEIBCRaw() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "memo-hl-to-ibc [ibc-source-chan] [ibc-recipient] [hub-token] [ibc timeout duration]",
		Args:                       cobra.ExactArgs(4),
		Short:                      "Get the memo for the direction HL -> IBC or EIBC",
		Example:                    `dymd q forward memo-hl-to-ibc "channel-0" ethm1a30y0h95a7p38plnv5s02lzrgcy0m0xumq0ymn 100ibc/9A1EACD53A6A197ADC81DF9A49F0C4A26F7FF685ACF415EE726D7D59796E71A7 5m`,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {

			hook, err := memoForwardToIBC(args)
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
				fmt.Printf("hyperlane message: %s\n", hyperutil.EncodeEthHex(bz))
			}

			return nil
		},
	}

	cmd.Flags().Bool(MessageReadableFlag, false, "Show the message in a human readable format (for debug)")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// Get a message for the direction HL -> (E)IBC. Intended for testing (check that the hub handles inbound messages correctly.)
func CmdTestHLtoIBCMessage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hl-message [nonce] [src-domain] [src-contract] [dst-domain] [token-id] [hyperlane recipient] [amount] [ibc-source-chan] [ibc-recipient] [hub-token] [ibc timeout duration] [recovery-address]",
		Args:  cobra.ExactArgs(12),
		Short: "Create a hyperlane message for testing Hl -> IBC",
		Example: `
		dymd q forward hl-message 1 1
0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0 1 0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0
dym139mq752delxv78jvtmwxhasyrycufsvrw4aka9 50 channel-0 ethm1wqg8227q0p7pgp7lj7z6cu036l6eg34d9cp6lk 100ibc/9A1EACD53A6A197ADC81DF9A49F0C4A26F7FF685ACF415EE726D7D59796E71A7 5m
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

			hlSrcContract, err := hyperutil.DecodeHexAddress(args[2])
			if err != nil {
				return fmt.Errorf("counterparty contract: %w", err)
			}

			hlDstDomain, err := strconv.ParseUint(args[3], 10, 32)
			if err != nil {
				return fmt.Errorf("local domain: %w", err)
			}

			hlTokenID, err := hyperutil.DecodeHexAddress(args[4])
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

			hook, err := memoForwardToIBC(args[7:])
			if err != nil {
				return fmt.Errorf("memo hl to ibc: %w", err)
			}

			readable, err := cmd.Flags().GetBool(MessageReadableFlag)
			if err != nil {
				return fmt.Errorf("encode flag: %w", err)
			}

			m, err := types.NewForwardToIBCHyperlaneMessage(
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

			bz, err := hyperutil.DecodeEthHex(d)
			if err != nil {
				return fmt.Errorf("decode eth hex: %w", err)
			}

			var message *hyperutil.HyperlaneMessage
			var body []byte
			if kind == "message" {
				m, err := hyperutil.ParseHyperlaneMessage(bz)
				if err != nil {
					return fmt.Errorf(" %w", err)
				}
				body = m.Body
				message = &m
			}
			payload, err := warptypes.ParseWarpMemoPayload(body)
			if err != nil {
				return fmt.Errorf("parse warp memo payload: %w", err)
			}
			fmt.Printf("hyperlane message: %+v\n", message)
			fmt.Printf("token message: %+v\n", payload)

			memo, err := types.UnpackForwardToIBC(payload.Memo)
			if err != nil {
				return fmt.Errorf("unpack memo from warp message: %w", err)
			}
			fmt.Printf("eibc memo: %+v\n", memo)
			return nil
		},
	}

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

func memoForwardToIBC(args []string) (*types.HookForwardToIBC, error) {

	ibcSourceChan := args[0]

	ibcRecipient := args[1]

	hubToken, err := sdk.ParseCoinNormalized(args[2])
	if err != nil {
		return nil, fmt.Errorf("hub token: %w", err)
	}

	ibcTimeoutDuration, err := time.ParseDuration(args[3])
	if err != nil {
		return nil, fmt.Errorf("ibc timeout duration: %w", err)
	}

	ibcTimeoutTimestamp := uint64(time.Now().Add(ibcTimeoutDuration).UnixNano())

	hook := types.MakeHookForwardToIBC(
		ibcSourceChan,
		hubToken,
		ibcRecipient,
		ibcTimeoutTimestamp,
	)
	err = hook.ValidateBasic()
	if err != nil {
		return nil, fmt.Errorf("new hl message: %w", err)
	}
	return hook, nil
}
