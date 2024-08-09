package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/dymensionxyz/dymension/v3/app/params"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/spf13/cobra"
)

// NewPlaceBidOnDymNameOrderTxCmd is the CLI command for placing a bid on a Dym-Name Sell-Order.
func NewPlaceBidOnDymNameOrderTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("bid-name [Dym-Name] [amount] %s", params.DisplayDenom),
		Aliases: []string{"bid"},
		Short:   "place a bid on a Dym-Name Sell-Order",
		Example: fmt.Sprintf(
			"$ %s tx %s bid-name myname 100 %s --%s hub-user",
			version.AppName, dymnstypes.ModuleName, params.DisplayDenom, flags.FlagFrom,
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			dymName := args[0]
			if !dymnsutils.IsValidDymName(dymName) {
				return fmt.Errorf("invalid Dym-Name: %s", dymName)
			}
			amount, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil || amount < 1 {
				return fmt.Errorf("amount must be a positive number")
			}
			if amount > maxDymBuyValueInteractingCLI {
				return fmt.Errorf(
					"excess maximum bid value, you should go to dApp. To prevent mistakenly in input, the maximum amount allowed via CLI is: %d %s",
					maxDymBuyValueInteractingCLI, params.DisplayDenom,
				)
			}
			denom := args[2]
			if !strings.EqualFold(denom, params.DisplayDenom) {
				return fmt.Errorf("denom must be: %s", strings.ToUpper(params.DisplayDenom))
			}

			buyer := clientCtx.GetFromAddress().String()
			if buyer == "" {
				return fmt.Errorf("flag --%s is required", flags.FlagFrom)
			}

			msg := &dymnstypes.MsgPurchaseOrder{
				GoodsId:   dymName,
				OrderType: dymnstypes.NameOrder,
				Offer:     sdk.NewCoin(params.BaseDenom, sdk.NewInt(int64(amount)).MulRaw(adymToDymMultiplier)),
				Buyer:     buyer,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
