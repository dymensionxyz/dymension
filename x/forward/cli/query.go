package cli

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"cosmossdk.io/math"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
)

func GetQueryCmd() *cobra.Command {
	// Group eibc queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdForwardMemo())
	cmd.AddCommand(CmdHyperlaneMessage())
	cmd.AddCommand(CmdDecodeHyperlane())

	return cmd
}

func CmdForwardMemo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forward [eibc-fee] [token-id] [destination-domain] [recipient] [amount] [max-fee] [recovery-address]",
		Args:  cobra.ExactArgs(7),
		Short: "Create a forward memo for IBC transfer",

		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {

			eibcFee := args[0]
			_, err := strconv.Atoi(eibcFee)
			if err != nil {
				return fmt.Errorf("invalid eibc fee: %w", err)
			}

			tokenId, err := hyperutil.DecodeHexAddress(args[1])
			if err != nil {
				return fmt.Errorf("invalid token id: %w", err)
			}

			destinationDomain, err := strconv.ParseUint(args[2], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid destination domain: %w", err)
			}

			recipient, err := hyperutil.DecodeHexAddress(args[3])
			if err != nil {
				return fmt.Errorf("invalid recipient: %w", err)
			}

			amount, ok := math.NewIntFromString(args[4])
			if !ok {
				return fmt.Errorf("invalid amount")
			}

			maxFee, err := sdk.ParseCoinNormalized(args[5])
			if err != nil {
				return fmt.Errorf("invalid max fee: %w", err)
			}

			recovery := args[6]

			memo, err := types.NewForwardMemo(
				eibcFee,
				tokenId,
				uint32(destinationDomain),
				recipient,
				amount,
				maxFee,
				recovery,
				math.ZeroInt(),
				nil,
				"",
			)
			if err != nil {
				return fmt.Errorf("invalid memo: %w", err)
			}

			fmt.Println(memo)
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdHyperlaneMessage() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "hyperlane-message [nonce] [src-domain] [src-contract] [dst-domain] [token-id] [hyperlane recipient] [amount] [ibc-source-chan] [ibc-recipient] [hub-token] [ibc timeout duration] [recovery-address]",
		Args:  cobra.ExactArgs(12),
		Short: "Create a hyperlane message for testing Hl -> IBC",
		Example: `
		dymd q forward hyperlane-message 1 1
0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0 1 0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0
dym139mq752delxv78jvtmwxhasyrycufsvrw4aka9 50 channel-0 ethm1wqg8227q0p7pgp7lj7z6cu036l6eg34d9cp6lk 100ibc/9A1EACD53A6A197ADC81DF9A49F0C4A26F7FF685ACF415EE726D7D59796E71A7 5m
dym1yecvrgz7yp26keaxa4r00554uugatxfegk76hz`,

		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {

			hlNonce, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid nonce: %w", err)
			}

			hlSrcDomain, err := strconv.ParseUint(args[1], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid counterparty domain: %w", err)
			}

			hlSrcContract, err := hyperutil.DecodeHexAddress(args[2])
			if err != nil {
				return fmt.Errorf("invalid counterparty contract: %w", err)
			}

			hlDstDomain, err := strconv.ParseUint(args[3], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid local domain: %w", err)
			}

			hlTokenID, err := hyperutil.DecodeHexAddress(args[4])
			if err != nil {
				return fmt.Errorf("invalid token id: %w", err)
			}

			hlRecipient, err := sdk.AccAddressFromBech32(args[5])
			if err != nil {
				return fmt.Errorf("invalid recipient address: %w", err)
			}

			hlAmt, ok := math.NewIntFromString(args[6])
			if !ok {
				return fmt.Errorf("invalid amount")
			}

			ibcSourceChan := args[7]

			ibcRecipient := args[8]

			hubToken, err := sdk.ParseCoinNormalized(args[9])
			if err != nil {
				return fmt.Errorf("invalid hub token: %w", err)
			}

			ibcTimeoutDuration, err := time.ParseDuration(args[10])
			if err != nil {
				return fmt.Errorf("invalid ibc timeout duration: %w", err)
			}

			ibcTimeoutTimestamp := uint64(time.Now().Add(ibcTimeoutDuration).UnixNano())

			recovery := args[11]

			m, err := types.NewHyperlaneMessage(
				uint32(hlNonce),
				uint32(hlSrcDomain),
				hlSrcContract,
				uint32(hlDstDomain),
				hlTokenID,
				hlRecipient,
				hlAmt,
				ibcSourceChan,
				ibcRecipient,
				hubToken,
				ibcTimeoutTimestamp,
				recovery,
			)
			if err != nil {
				return fmt.Errorf("new hl message: %w", err)
			}

			s := m.String()
			fmt.Println(s)
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// A quick util to debug hyperlane messages (including show the memo if there is one). Expects Ethereum Hex bytes
func CmdDecodeHyperlane() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "hyperlane-decode (body | message) [hexstring]",
		Args:                       cobra.ExactArgs(2),
		Short:                      "Decode a message or message body from an ethereum hex string",
		Example:                    `dymd q forward hyperlane-message-decode message 0x00000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000000`,
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {

			kind := args[0]
			if kind != "body" && kind != "message" {
				return fmt.Errorf("invalid message type: %s", kind)
			}
			d := args[1]

			d = strings.TrimSpace(d)
			d = strings.ReplaceAll(d, "\"", "")

			bz, err := hyperutil.DecodeEthHex(d)
			if err != nil {
				return fmt.Errorf("invalid hexstring: %w", err)
			}

			var message *hyperutil.HyperlaneMessage
			var body []byte
			if kind == "message" {
				m, err := hyperutil.ParseHyperlaneMessage(bz)
				if err != nil {
					return fmt.Errorf("invalid: %w", err)
				}
				body = m.Body
				message = &m
			}
			payload, err := warptypes.ParseWarpMemoPayload(body)
			if err != nil {
				return fmt.Errorf("invalid hexstring: %w", err)
			}
			fmt.Printf("hyperlane message: %+v\n", message)
			fmt.Printf("token message: %+v\n", payload)

			memo, _, err := types.UnpackMemoFromHyperlane(payload.Memo)
			if err != nil {
				return fmt.Errorf("invalid memo: %w", err)
			}
			fmt.Printf("eibc memo: %+v\n", memo)
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
