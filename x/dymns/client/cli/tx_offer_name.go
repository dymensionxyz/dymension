package cli

import (
	"context"
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

const (
	flagContinueOfferId = "continue-offer-id"
)

// NewOfferBuyDymNameTxCmd is the CLI command for creating an offer to buy a Dym-Name.
func NewOfferBuyDymNameTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("offer-name [Dym-Name] [amount] %s", params.DisplayDenom),
		Aliases: []string{"offer"},
		Short:   "Create an offer to buy a Dym-Name",
		Example: fmt.Sprintf(
			"$ %s tx %s offer myname 50 %s --%s hub-user",
			version.AppName, dymnstypes.ModuleName,
			params.DisplayDenom,
			flags.FlagFrom,
		),
		Args: cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			dymName := args[0]
			if !dymnsutils.IsValidDymName(dymName) {
				return fmt.Errorf("input Dym-Name '%s' is not a valid Dym-Name", dymName)
			}

			amount, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil || amount < 1 {
				return fmt.Errorf("amount must be a positive number")
			}

			const maximumOfferValue = 10_000
			if amount > maximumOfferValue {
				return fmt.Errorf("maximum offer value allowed via CLI is %d %s", maximumOfferValue, params.DisplayDenom)
			}
			denom := args[2]
			if !strings.EqualFold(denom, params.DisplayDenom) {
				return fmt.Errorf("denom must be %s", strings.ToUpper(params.DisplayDenom))
			}

			buyer := clientCtx.GetFromAddress().String()
			if buyer == "" {
				return fmt.Errorf("flag --%s is required", flags.FlagFrom)
			}

			continueOfferId, _ := cmd.Flags().GetString(flagContinueOfferId)
			if continueOfferId != "" && !dymnstypes.IsValidBuyOfferId(continueOfferId) {
				return fmt.Errorf("invalid continue offer id")
			}

			queryClient := dymnstypes.NewQueryClient(clientCtx)

			resParams, err := queryClient.Params(context.Background(), &dymnstypes.QueryParamsRequest{})
			if err != nil {
				return err
			}

			msg := &dymnstypes.MsgPlaceBuyOrder{
				GoodsId:         dymName,
				OrderType:       dymnstypes.MarketOrderType_MOT_DYM_NAME,
				Buyer:           buyer,
				ContinueOfferId: continueOfferId,
				Offer: sdk.Coin{
					Denom:  resParams.Params.Price.PriceDenom,
					Amount: sdk.NewInt(int64(amount)).MulRaw(1e18),
				},
			}

			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(flagContinueOfferId, "", "if provided, will raise offer value of an existing offer")

	return cmd
}
