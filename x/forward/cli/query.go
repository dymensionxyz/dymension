package cli

import (
	"fmt"
	"strconv"

	"cosmossdk.io/math"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"

	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
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

			/*
			   What is the memo actually supposed to be
			   {
			   eibc:..,
			   fulfill_hook:

			   	BYTES(
			   		type FulfillHook struct {
			   			HookName string = 'forward'
			   			HookData BYTES(HookEIBCtoHL)
			   		}
			   	)

			   )
			   }
			*/

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

			// TODO: fix
			// recovery := "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96"
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
		Use:   "hyperlane-message [counterparty-domain] [counterparty-contract] [local-domain] [token-id] [recipient] [amount]",
		Args:  cobra.ExactArgs(6),
		Short: "Create a forward memo for IBC transfer",

		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE: func(cmd *cobra.Command, args []string) error {

			counterpartyDomain, err := strconv.ParseUint(args[0], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid counterparty domain: %w", err)
			}

			// need to decode hex addresses
			counterpartyContract, err := hyperutil.DecodeHexAddress(args[1])
			if err != nil {
				return fmt.Errorf("invalid counterparty contract: %w", err)
			}

			localDomain, err := strconv.ParseUint(args[2], 10, 32)
			if err != nil {
				return fmt.Errorf("invalid local domain: %w", err)
			}

			tokenId, err := hyperutil.DecodeHexAddress(args[3])
			if err != nil {
				return fmt.Errorf("invalid token id: %w", err)
			}

			recipient, err := sdk.AccAddressFromBech32(args[4])
			if err != nil {
				return fmt.Errorf("invalid recipient address: %w", err)
			}

			amount, ok := math.NewIntFromString(args[5])
			if !ok {
				return fmt.Errorf("invalid amount")
			}

			m, err := types.NewHyperlaneMessage(
				uint32(counterpartyDomain),
				counterpartyContract,
				uint32(localDomain),
				tokenId,
				recipient,
				amount,
			)
			if err != nil {
				return fmt.Errorf("invalid memo: %w", err)
			}

			bz := m.Bytes()
			fmt.Println(bz)
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
