package cli

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"

	"github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
)

type decodedHL struct {
	hlMsg        *util.HyperlaneMessage
	warpPL       warptypes.WarpPayload
	rawMetadata  []byte            // always set when metadata exists
	hlMetadata   *types.HLMetadata // set only when rawMetadata is valid HLMetadata protobuf
	forwardToIBC *types.HookForwardToIBC
	forwardToHL  *types.HookForwardToHL
}

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

	decoded, err := parseHL(bz)
	if err != nil {
		return err
	}

	printDecoded(decoded)
	return nil
}

func parseHL(bz []byte) (*decodedHL, error) {
	decoded := &decodedHL{}

	// Input is either a full HL message (envelope + body) or just the body (warp payload).
	// Try HL message first; version 0 means it's not a real HL message.
	body := bz
	if msg, err := util.ParseHyperlaneMessage(bz); err == nil && msg.Version != 0 && len(msg.Body) > 0 {
		decoded.hlMsg = &msg
		body = msg.Body
	}

	warpPL, err := warptypes.ParseWarpPayload(body)
	if err != nil {
		return nil, fmt.Errorf("parse warp payload: %w", err)
	}
	decoded.warpPL = warpPL

	metadata := warpPL.Metadata()
	if len(metadata) > 0 {
		decoded.rawMetadata = metadata
		if hlMetadata, err := types.UnpackHLMetadata(metadata); err == nil {
			decoded.hlMetadata = hlMetadata
			if len(hlMetadata.HookForwardToIbc) > 0 {
				decoded.forwardToIBC, _ = types.UnpackForwardToIBC(hlMetadata.HookForwardToIbc)
			}
			if len(hlMetadata.HookForwardToHl) > 0 {
				decoded.forwardToHL, _ = types.UnpackForwardToHL(hlMetadata.HookForwardToHl)
			}
		}
	}

	return decoded, nil
}

func printDecoded(d *decodedHL) {
	if d.hlMsg != nil {
		printHyperlaneMessage(*d.hlMsg)
	}

	printForwardingMemo(d)
	printWarpPayload(d.warpPL)
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

func printForwardingMemo(d *decodedHL) {
	fmt.Println("\n=== Forwarding Memo ===")

	if d.rawMetadata == nil {
		fmt.Println("  (none)")
		return
	}

	fmt.Printf("  Raw hex:    %s\n", util.EncodeEthHex(d.rawMetadata))
	if isASCIIPrintable(d.rawMetadata) {
		fmt.Printf("  Raw string: %s\n", string(d.rawMetadata))
	}

	if d.hlMetadata == nil {
		fmt.Println("  (not a valid HLMetadata protobuf)")
		return
	}

	m := d.hlMetadata
	hasIBC := len(m.HookForwardToIbc) > 0
	hasHL := len(m.HookForwardToHl) > 0
	hasKaspa := len(m.Kaspa) > 0

	if !hasIBC && !hasHL && !hasKaspa {
		fmt.Println("  (empty HLMetadata - no forwarding)")
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

	if d.forwardToIBC != nil {
		printForwardToIBC(d.forwardToIBC)
	}
	if d.forwardToHL != nil {
		printForwardToHL(d.forwardToHL)
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

	recipient := warpPL.Recipient()
	// EVM addresses are 20 bytes zero-padded to 32 bytes (12 leading zeros + 20 bytes)
	if len(recipient) == 32 && bytes.Equal(recipient[:12], make([]byte, 12)) {
		fmt.Printf("  EVM Address:    0x%s\n", hex.EncodeToString(recipient[12:]))
	}
	fmt.Printf("  Cosmos Account: %s\n", warpPL.GetCosmosAccount().String())

	fmt.Printf("  Amount:         %s\n", warpPL.Amount().String())
}

func isASCIIPrintable(data []byte) bool {
	for _, b := range data {
		if b < 32 || b > 126 {
			return false
		}
	}
	return len(data) > 0
}
