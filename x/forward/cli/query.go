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
	"github.com/dymensionxyz/dymension/v3/utils/utransfer"
	"github.com/dymensionxyz/dymension/v3/x/forward/keeper"
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

	return cmd
}

func CmdForwardMemo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forward [eibc-fee] [token-id] [destination-domain] [recipient] [amount] [max-fee]",
		Args:  cobra.ExactArgs(6),
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

			// TODO: fix
			recovery := "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96"

			memo, err := NewForwardMemo(
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

			fmt.Println(memo)
			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func NewForwardMemo(
	eibcFee string,
	tokenId hyperutil.HexAddress,
	destinationDomain uint32,
	recipient hyperutil.HexAddress,
	amount math.Int,
	maxFee sdk.Coin,

	recoveryAddr string,

	gasLimit math.Int,
	customHookId *hyperutil.HexAddress,
	customHookMetadata string) (string, error) {

	hook, err := keeper.NewEIBCFulfillHook(
		types.NewHookEIBCtoHL(
			types.NewRecovery(recoveryAddr),
			tokenId,
			destinationDomain,
			recipient,
			amount,
			maxFee,
			gasLimit,
			customHookId,
			customHookMetadata,
		),
	)
	if err != nil {
		return "", err
	}

	return utransfer.CreateMemo(eibcFee, hook.HookData), nil
}
