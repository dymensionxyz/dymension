package cli

import (
	"context"
	"fmt"

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

const (
	flagMinPrice             = "min-price"
	flagImmediatelySellPrice = "immediately-sell-price"
)

// NewPlaceDymNameSellOrderTxCmd is the CLI command for creating a Sell-Order to sell a Dym-Name.
func NewPlaceDymNameSellOrderTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sell-name [Dym-Name]",
		Aliases: []string{"sell"},
		Short:   "Create a sell-order to sell your Dym-Name",
		Long:    fmt.Sprintf(`Create a sell-order to sell your Dym-Name. Flag --%s indicate the starting price of the Dym-Name, and flag --%s indicate the immediately sell price of the Dym-Name. If immediately sell price is not supplied or the highest bid does not reaching this amount, auction can only be ended when the sell-order expired.`, flagMinPrice, flagImmediatelySellPrice),
		Example: fmt.Sprintf(
			"$ %s tx %s sell-name myname --%s 50 [--%s 100] --%s hub-user",
			version.AppName, dymnstypes.ModuleName,
			flagMinPrice, flagImmediatelySellPrice,
			flags.FlagFrom,
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			dymName := args[0]
			if !dymnsutils.IsValidDymName(dymName) {
				return fmt.Errorf("input Dym-Name '%s' is not a valid Dym-Name", dymName)
			}

			minPriceDym, err := cmd.Flags().GetUint64(flagMinPrice)
			if err != nil {
				return fmt.Errorf("error reading flag --%s: %v", flagMinPrice, err)
			}
			sellPriceDym, err := cmd.Flags().GetUint64(flagImmediatelySellPrice)
			if err != nil {
				return fmt.Errorf("error reading flag --%s: %v", flagImmediatelySellPrice, err)
			}

			if minPriceDym < 1 {
				return fmt.Errorf("--%s must be a positive number", flagMinPrice)
			}

			if sellPriceDym > 0 && sellPriceDym < minPriceDym {
				return fmt.Errorf("--%s must be greater than or equal to --%s", flagImmediatelySellPrice, flagMinPrice)
			}

			const maximumPriceDym = 10_000_000
			if minPriceDym > maximumPriceDym || sellPriceDym > maximumPriceDym {
				return fmt.Errorf("price is too high, over %d %s", maximumPriceDym, params.DisplayDenom)
			}

			seller := clientCtx.GetFromAddress().String()
			if seller == "" {
				return fmt.Errorf("flag --%s is required", flags.FlagFrom)
			}

			queryClient := dymnstypes.NewQueryClient(clientCtx)

			resParams, err := queryClient.Params(context.Background(), &dymnstypes.QueryParamsRequest{})
			if err != nil {
				return err
			}

			var sellPrice *sdk.Coin
			if sellPriceDym > 0 {
				sellPrice = &sdk.Coin{
					Denom:  resParams.Params.Price.PriceDenom,
					Amount: sdk.NewInt(int64(sellPriceDym)).MulRaw(1e18),
				}
			}

			msg := &dymnstypes.MsgPlaceSellOrder{
				GoodsId:   dymName,
				OrderType: dymnstypes.NameOrder,
				MinPrice:  sdk.NewCoin(resParams.Params.Price.PriceDenom, sdk.NewInt(int64(minPriceDym)).MulRaw(1e18)),
				SellPrice: sellPrice,
				Owner:     seller,
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().Uint64(flagMinPrice, 0, "minimum price to sell the Dym-Name")
	cmd.Flags().Uint64(flagImmediatelySellPrice, 0, "immediately sell price of the Dym-Name, when someone placed a bid on it that matching the immediately sell price, auction stopped and the Dym-Name will be sold immediately, otherwise the Dym-Name will be sold to the highest bidder when the sell-order expired")

	return cmd
}
