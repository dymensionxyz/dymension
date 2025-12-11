package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
)

// CmdDecodeHL returns the decode-hl command
func CmdDecodeHL() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decode-hl [hexstring]",
		Args:  cobra.ExactArgs(1),
		Short: "Decode a Hyperlane message or warp payload from hex string",
		Long: `Decode a Hyperlane message or warp payload from hex string.

The command auto-detects the format:
- If the input is a full Hyperlane message, it decodes the message envelope and the warp payload body
- If the input is just a warp payload body, it decodes that directly

The forwarding memo (HLMetadata) is decoded by default if present.`,
		Example: `# Decode a full Hyperlane message
dymd q forward decode-hl 0x0300023889...

# Decode just a warp payload body
dymd q forward decode-hl 0x000000000000...`,
		RunE: runDecodeHL,
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func runDecodeHL(cmd *cobra.Command, args []string) error {
	input := args[0]
	input = strings.TrimSpace(input)
	input = strings.ReplaceAll(input, `\`, "")
	input = strings.ReplaceAll(input, " ", "")

	bz, err := util.DecodeEthHex(input)
	if err != nil {
		return fmt.Errorf("decode hex: %w", err)
	}

	// Try to parse as a full Hyperlane message first (returns bz unchanged if not)
	body := tryParseAsHyperlaneMessage(bz)

	warpPL, err := warptypes.ParseWarpPayload(body)
	if err != nil {
		return fmt.Errorf("parse warp payload: %w", err)
	}

	// Decode the forwarding memo
	decodeForwardingMemo(warpPL)

	// Print warp payload
	printWarpPayload(warpPL)

	return nil
}

// tryParseAsHyperlaneMessage attempts to parse bytes as a Hyperlane message.
// Returns the message body if successful, or the original bytes if not.
func tryParseAsHyperlaneMessage(bz []byte) []byte {
	msg, err := util.ParseHyperlaneMessage(bz)
	if err != nil {
		return bz
	}

	// Sanity check: version should be reasonable (currently 3)
	if msg.Version > 10 {
		return bz
	}

	// Sanity check: body should not be empty for a warp message
	if len(msg.Body) == 0 {
		return bz
	}

	printHyperlaneMessage(msg)
	return msg.Body
}

func decodeForwardingMemo(warpPL warptypes.WarpPayload) {
	metadata := warpPL.Metadata()
	if len(metadata) == 0 {
		fmt.Println("\n=== Forwarding Memo ===")
		fmt.Println("  (none)")
		return
	}

	hlMetadata, err := types.UnpackHLMetadata(metadata)
	if err != nil {
		fmt.Println("\n=== Forwarding Memo ===")
		fmt.Println("  (not a valid HLMetadata protobuf)")
		fmt.Printf("  Raw hex:    %s\n", util.EncodeEthHex(metadata))
		if isASCIIPrintable(metadata) {
			fmt.Printf("  Raw string: %s\n", string(metadata))
		}
		return
	}

	printHLMetadata(hlMetadata)

	if hlMetadata != nil {
		if len(hlMetadata.HookForwardToIbc) > 0 {
			m, err := types.UnpackForwardToIBC(hlMetadata.HookForwardToIbc)
			if err != nil {
				fmt.Printf("\n=== Forward to IBC ===\n  (decode error: %v)\n", err)
			} else {
				printForwardToIBC(m)
			}
		}

		if len(hlMetadata.HookForwardToHl) > 0 {
			m, err := types.UnpackForwardToHL(hlMetadata.HookForwardToHl)
			if err != nil {
				fmt.Printf("\n=== Forward to Hyperlane ===\n  (decode error: %v)\n", err)
			} else {
				printForwardToHL(m)
			}
		}
	}
}

func printHyperlaneMessage(msg util.HyperlaneMessage) {
	fmt.Println("=== Hyperlane Message ===")
	fmt.Printf("  Version:     %d\n", msg.Version)
	fmt.Printf("  Nonce:       %d\n", msg.Nonce)
	fmt.Printf("  Origin:      %d\n", msg.Origin)
	fmt.Printf("  Sender:      %s\n", msg.Sender)
	fmt.Printf("  Destination: %d\n", msg.Destination)
	fmt.Printf("  Recipient:   %s\n", msg.Recipient)
	fmt.Printf("  Body:        %s\n", util.EncodeEthHex(msg.Body))
}

func printHLMetadata(m *types.HLMetadata) {
	fmt.Println("\n=== Forwarding Memo ===")
	if m == nil {
		fmt.Println("  (nil)")
		return
	}

	hasIBC := len(m.HookForwardToIbc) > 0
	hasHL := len(m.HookForwardToHl) > 0
	hasKaspa := len(m.Kaspa) > 0

	if !hasIBC && !hasHL && !hasKaspa {
		fmt.Println("  (empty - no forwarding)")
		return
	}

	if hasIBC {
		fmt.Printf("  Forward to IBC: yes (%d bytes)\n", len(m.HookForwardToIbc))
	}
	if hasHL {
		fmt.Printf("  Forward to HL:  yes (%d bytes)\n", len(m.HookForwardToHl))
	}
	if hasKaspa {
		fmt.Printf("  Kaspa metadata: yes (%d bytes)\n", len(m.Kaspa))
	}
}

func printForwardToIBC(m *types.HookForwardToIBC) {
	fmt.Println("\n=== Forward to IBC ===")
	if m == nil || m.Transfer == nil {
		fmt.Println("  (nil)")
		return
	}

	t := m.Transfer
	fmt.Printf("  Source Port:       %s\n", t.SourcePort)
	fmt.Printf("  Source Channel:    %s\n", t.SourceChannel)
	fmt.Printf("  Receiver:          %s\n", t.Receiver)
	fmt.Printf("  Timeout Timestamp: %d", t.TimeoutTimestamp)
	if t.TimeoutTimestamp > 0 {
		ts := time.Unix(0, int64(t.TimeoutTimestamp)) // #nosec G115 - timestamp is always positive
		fmt.Printf(" (%s)", ts.UTC().Format(time.RFC3339))
	}
	fmt.Println()
	if t.Memo != "" {
		fmt.Printf("  Memo:              %s\n", t.Memo)
	}
}

func printForwardToHL(m *types.HookForwardToHL) {
	fmt.Println("\n=== Forward to Hyperlane ===")
	if m == nil || m.HyperlaneTransfer == nil {
		fmt.Println("  (nil)")
		return
	}

	t := m.HyperlaneTransfer
	fmt.Printf("  Token ID:           %s\n", util.EncodeEthHex(t.TokenId[:]))
	fmt.Printf("  Destination Domain: %d\n", t.DestinationDomain)
	fmt.Printf("  Recipient:          %s\n", util.EncodeEthHex(t.Recipient[:]))
	fmt.Printf("  Amount:             %s\n", t.Amount.String())
	fmt.Printf("  Max Fee:            %s\n", t.MaxFee.String())
	if !t.GasLimit.IsNil() && !t.GasLimit.IsZero() {
		fmt.Printf("  Gas Limit:          %s\n", t.GasLimit.String())
	}
	if t.CustomHookId != nil {
		fmt.Printf("  Custom Hook ID:     %s\n", util.EncodeEthHex(t.CustomHookId[:]))
	}
	if t.CustomHookMetadata != "" {
		fmt.Printf("  Custom Hook Meta:   %s\n", t.CustomHookMetadata)
	}
}

func printWarpPayload(warpPL warptypes.WarpPayload) {
	fmt.Println("\n=== Warp Payload ===")
	fmt.Printf("  Cosmos Account: %s\n", warpPL.GetCosmosAccount().String())
	fmt.Printf("  Amount:         %s\n", warpPL.Amount().String())
}

// isASCIIPrintable checks if all bytes are printable ASCII characters
func isASCIIPrintable(data []byte) bool {
	for _, b := range data {
		if b < 32 || b > 126 {
			return false
		}
	}
	return len(data) > 0
}
