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
	"github.com/spf13/cobra"
)

// NewAcceptBuyOrderTxCmd is the CLI command for accepting a Buy-Order of a Dym-Name/Alias/Handle.
// or offer-back to raise the price.
func NewAcceptBuyOrderTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("accept-offer [offer-id] [amount] %s", params.DisplayDenom),
		Short: "Accept a Buy-Order for your Dym-Name/Alias/Handle",
		Example: fmt.Sprintf(
			"$ %s tx %s accept-offer 1 50 %s --%s owner",
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

			orderId := args[0]
			if !dymnstypes.IsValidBuyOrderId(orderId) {
				return fmt.Errorf("input is not a valid Buy-Order ID: %s", orderId)
			}

			amount, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil || amount < 1 {
				return fmt.Errorf("amount must be a positive number")
			}

			denom := args[2]
			if !strings.EqualFold(denom, params.DisplayDenom) {
				return fmt.Errorf("denom must be %s", strings.ToUpper(params.DisplayDenom))
			}

			owner := clientCtx.GetFromAddress().String()
			if owner == "" {
				return fmt.Errorf("flag --%s is required", flags.FlagFrom)
			}

			queryClient := dymnstypes.NewQueryClient(clientCtx)

			resParams, err := queryClient.Params(cmd.Context(), &dymnstypes.QueryParamsRequest{})
			if err != nil {
				return err
			}

			msg := &dymnstypes.MsgAcceptBuyOrder{
				OrderId: orderId,
				Owner:   owner,
				MinAccept: sdk.Coin{
					Denom:  resParams.Params.Price.PriceDenom,
					Amount: sdk.NewInt(amount).MulRaw(adymToDymMultiplier),
				},
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
