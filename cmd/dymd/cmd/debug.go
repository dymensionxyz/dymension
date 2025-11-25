package cmd

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/evmos/ethermint/ethereum/eip712"

	"github.com/cosmos/cosmos-sdk/client"
	cosmosclientdebug "github.com/cosmos/cosmos-sdk/client/debug"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
)

// NOTE: flagPrefix is defined as a local constant in AddrCmd to prevent global namespace pollution.

// DebugCmds creates the main CLI command for debugging tools.
func DebugCmds() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "debug",
		Short: "Tool for helping with debugging your application",
		RunE:  client.ValidateCmd,
	}

	cmd.AddCommand(
		// Default Cosmos SDK debug commands
		cosmosclientdebug.CodecCmd(),
		cosmosclientdebug.PrefixesCmd(),
		// PubkeyRawCmd might need internal SDK adjustments for eth_secp256k1 support.
		cosmosclientdebug.PubkeyRawCmd(),

		// Cosmos EVM adjusted debug commands
		PubkeyCmd(),
		AddrCmd(),
		RawBytesCmd(),
		LegacyEIP712Cmd(),
	)

	return cmd
}

// getPubKeyFromString decodes an SDK PubKey from proto JSON format.
func getPubKeyFromString(ctx client.Context, pkstr string) (cryptotypes.PubKey, error) {
	var pk cryptotypes.PubKey
	err := ctx.Codec.UnmarshalInterfaceJSON([]byte(pkstr), &pk)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal PubKey from JSON")
	}
	return pk, nil
}

// PubkeyCmd decodes a proto JSON pubkey and displays its associated addresses (EIP-55 and Bech32).
func PubkeyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "pubkey [pubkey]",
		Short: "Decode a pubkey from proto JSON and display its addresses",
		Example: fmt.Sprintf(
			`$ %s debug pubkey '{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AurroA7jvfPd1AadmmOvWM2rJSwipXfRf8yD6pLbA2DJ"}'`, //gitleaks:allow
			version.AppName,
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			pk, err := getPubKeyFromString(clientCtx, args[0])
			if err != nil {
				return err
			}

			addr := pk.Address()
			cmd.Printf("Address (EIP-55): %s\n", common.BytesToAddress(addr).String()) // Added .String() for explicit string conversion
			cmd.Printf("Bech32 Acc: %s\n", sdk.AccAddress(addr).String())
			cmd.Println("PubKey Hex:", hex.EncodeToString(pk.Bytes()))
			return nil
		},
	}
}

// AddrCmd converts an address between hex (EIP-55) and bech32 formats.
func AddrCmd() *cobra.Command {
	const flagPrefix = "prefix"
	cmd := &cobra.Command{
		Use:   "addr [address]",
		Short: "Convert an address between hex and bech32",
		Long:  "Convert an address between hex encoding (EIP-55) and bech32.",
		Example: fmt.Sprintf(
			`$ %s debug addr cosmos1qqqqhe5pnaq5qq39wqkn957aydnrm45sdn8583
$ %s debug addr 0x00000Be6819f41400225702D32d3dd23663Dd690 --prefix cosmosevmtypes`, version.AppName, version.AppName),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			addrString := args[0]
			prefix, err := cmd.Flags().GetString(flagPrefix)
			if err != nil {
				return err
			}

			switch {
			case common.IsHexAddress(addrString):
				// Hex to Bech32 conversion
				addr := common.HexToAddress(addrString).Bytes()
				cmd.Println("Address bytes:", addr)

				if prefix == "" {
					cmd.Printf("Bech32 Acc: %s\n", sdk.AccAddress(addr))
					cmd.Printf("Bech32 Val: %s\n", sdk.ValAddress(addr))
				} else {
					// Use specific custom prefix
					bech32Address, err := sdk.Bech32ifyAddressBytes(prefix, addr)
					if err != nil {
						return errors.Wrapf(err, "failed to convert hex address to bech32 with prefix %s", prefix)
					}
					cmd.Printf("Bech32 %s\n", bech32Address)
				}
			default:
				// Bech32 to Hex conversion
				// Rely on the SDK's GetFromBech32 for robustness, avoiding manual string manipulation.
				prefixFromAddr := strings.SplitN(addrString, "1", 2)[0]
				hexAddrBytes, err := sdk.GetFromBech32(addrString, prefixFromAddr)
				if err != nil {
					return errors.Wrapf(err, "failed to decode bech32 address %s", addrString)
				}

				hexAddrString := common.BytesToAddress(hexAddrBytes).String()

				cmd.Println("Address bytes:", hexAddrBytes)
				cmd.Printf("Address hex (EIP-55): %s\n", hexAddrString)
			}
			return nil
		},
	}

	cmd.Flags().String(flagPrefix, "", "Bech32 encoded address prefix for conversion (e.g., cosmosevmtypes, cosmosvaloper)")
	return cmd
}

// RawBytesCmd converts raw byte representation (e.g., [10 21 13]) into a hex string.
func RawBytesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "raw-bytes [raw-bytes]",
		Short: "Convert raw bytes output (eg. [10 21 13 255]) to hex",
		Example: fmt.Sprintf(`$ %s debug raw-bytes [72 101 108 108 111 44 32 112 108 97 121 103 114 111 117 110 100]`, version.AppName),
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			stringBytes := args[0]
			stringBytes = strings.Trim(stringBytes, "[")
			stringBytes = strings.Trim(stringBytes, "]")
			
			// OPTIMIZATION: Use strings.Fields to safely handle multiple spaces and trim remaining whitespace.
			spl := strings.Fields(stringBytes)

			byteArray := make([]byte, 0, len(spl))
			for _, s := range spl {
				// Parse the string as a decimal integer (base 10) into an 8-bit integer.
				b, err := strconv.ParseInt(s, 10, 8)
				if err != nil {
					return errors.Wrapf(err, "failed to parse byte value %s", s)
				}
				byteArray = append(byteArray, byte(b))
			}
			fmt.Printf("%X\n", byteArray)
			return nil
		},
	}
}

// LegacyEIP712Cmd outputs types of legacy EIP712 typed data for a given transaction.
func LegacyEIP712Cmd() *cobra.Command {
	return &cobra.Command{
		Use:   "legacy-eip712 [file] [evm-chain-id]",
		Short: "Output types of legacy eip712 typed data according to the given transaction",
		Example: fmt.Sprintf(`$ %s debug legacy-eip712 tx.json 4221 --chain-id evmd-1`, version.AppName),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// Read transaction from file
			stdTx, err := authclient.ReadTxFromFile(clientCtx, args[0])
			if err != nil {
				return errors.Wrap(err, "read tx from file")
			}

			// Encode transaction to bytes
			txBytes, err := clientCtx.TxConfig.TxJSONEncoder()(stdTx)
			if err != nil {
				return errors.Wrap(err, "encode tx")
			}

			// Parse EVM Chain ID (use ParseUint for better type consistency with uint64)
			evmChainID, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return errors.Wrap(err, "parse evm-chain-id")
			}

			// Wrap the transaction into EIP712 Typed Data structure.
			// Gosec Warning G115 is mitigated by using ParseUint and explicit casting.
			td, err := eip712.LegacyWrapTxToTypedData(clientCtx.Codec, evmChainID, stdTx.GetMsgs()[0], txBytes, nil)
			if err != nil {
				return errors.Wrap(err, "wrap tx to typed data")
			}

			// Extract and marshal the "types" map from the TypedData object.
			bz, err := json.MarshalIndent(td.Map()["types"], "", "  ")
			if err != nil {
				return err
			}

			fmt.Println(string(bz))
			return nil
		},
	}
}
