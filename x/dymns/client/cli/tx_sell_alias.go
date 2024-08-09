package cli

import (
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

// NewPlaceAliasSellOrderTxCmd is the CLI command for creating a Sell-Order to sell an Alias/Handle.
func NewPlaceAliasSellOrderTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sell-alias [Alias/Handle]",
		Aliases: []string{"sell-handle"},
		Short:   "Create a sell-order to sell Alias/Handle of a RollApp you owned",
		Long:    fmt.Sprintf(`Create a sell-order to sell Alias/Handle of a RollApp you owned. Flag --%s indicate the starting price of the Alias/Handle, and flag --%s indicate the immediately sell price of the Alias/Handle. If immediately sell price is not supplied or the highest bid does not reaching this amount, auction can only be ended when the sell-order expired.`, flagMinPrice, flagImmediatelySellPrice),
		Example: fmt.Sprintf(
			"$ %s tx %s sell-alias dym --%s 50 [--%s 100] --%s sequencer",
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

			alias := args[0]
			if !dymnsutils.IsValidAlias(alias) {
				return fmt.Errorf("input is not a valid Dym-Name: %s", alias)
			}

			minPriceDym, err := cmd.Flags().GetUint64(flagMinPrice)
			if err != nil {
				return fmt.Errorf("error reading flag --%s: %w", flagMinPrice, err)
			}
			sellPriceDym, err := cmd.Flags().GetUint64(flagImmediatelySellPrice)
			if err != nil {
				return fmt.Errorf("error reading flag --%s: %w", flagImmediatelySellPrice, err)
			}

			if minPriceDym < 1 {
				return fmt.Errorf("--%s must be a positive number", flagMinPrice)
			}

			if sellPriceDym > 0 && sellPriceDym < minPriceDym {
				return fmt.Errorf("--%s must be greater than or equal to --%s", flagImmediatelySellPrice, flagMinPrice)
			}

			if minPriceDym > maxDymSellValueInteractingCLI || sellPriceDym > maxDymSellValueInteractingCLI {
				return fmt.Errorf("price is too high, over %d %s", maxDymSellValueInteractingCLI, params.DisplayDenom)
			}

			seller := clientCtx.GetFromAddress().String()
			if seller == "" {
				return fmt.Errorf("flag --%s is required", flags.FlagFrom)
			}

			queryClient := dymnstypes.NewQueryClient(clientCtx)

			resParams, err := queryClient.Params(cmd.Context(), &dymnstypes.QueryParamsRequest{})
			if err != nil {
				return err
			}

			var sellPrice *sdk.Coin
			if sellPriceDym > 0 {
				sellPrice = &sdk.Coin{
					Denom:  resParams.Params.Price.PriceDenom,
					Amount: sdk.NewInt(int64(sellPriceDym)).MulRaw(adymToDymMultiplier),
				}
			}

			msg := &dymnstypes.MsgPlaceSellOrder{
				GoodsId:   alias,
				OrderType: dymnstypes.AliasOrder,
				MinPrice:  sdk.NewCoin(resParams.Params.Price.PriceDenom, sdk.NewInt(int64(minPriceDym)).MulRaw(adymToDymMultiplier)),
				SellPrice: sellPrice,
				Owner:     seller,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().Uint64(flagMinPrice, 0, "minimum price to sell the Alias/Handle")
	cmd.Flags().Uint64(flagImmediatelySellPrice, 0, "immediately sell price of the Alias/Handle, when someone placed a bid on it that matching the immediately sell price, auction stopped and the Alias/Handle will be sold immediately, otherwise the Alias/Handle will be sold to the highest bidder when the sell-order expired")

	return cmd
}
