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

// NewPlaceBidOnOrderTxCmd is the CLI command for placing a bid on a Dym-Name Sell-Order.
func NewPlaceBidOnOrderTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("bid-name [Dym-Name] [amount] %s", params.DisplayDenom),
		Aliases: []string{"bid"},
		Short:   "place a bid on a Dym-Name Sell-Order",
		Example: fmt.Sprintf(
			"$ %s tx %s bid-name myname 100 %s",
			version.AppName, dymnstypes.ModuleName, params.DisplayDenom,
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
			const maximumBidValue = 10_000
			if amount > maximumBidValue {
				return fmt.Errorf("maximum bid value allowed via CLI is: %d %s", maximumBidValue, params.DisplayDenom)
			}
			denom := args[2]
			if !strings.EqualFold(denom, params.DisplayDenom) {
				return fmt.Errorf("denom must be: %s", strings.ToUpper(params.DisplayDenom))
			}

			msg := &dymnstypes.MsgPurchaseOrder{
				GoodsId:   dymName,
				OrderType: dymnstypes.MarketOrderType_MOT_DYM_NAME,
				Offer:     sdk.NewCoin(params.BaseDenom, sdk.NewInt(int64(amount)).MulRaw(1e18)),
				Buyer:     clientCtx.GetFromAddress().String(),
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
