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

const (
	flagContinueOrderId = "continue-order-id"
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
				return fmt.Errorf("input is not a valid Dym-Name: %s", dymName)
			}

			amount, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil || amount < 1 {
				return fmt.Errorf("amount must be a positive number")
			}

			if amount > maxDymBuyValueInteractingCLI {
				return fmt.Errorf(
					"excess maximum offer value, you should go to dApp. To prevent mistakenly in input, the maximum amount allowed via CLI is: %d %s",
					maxDymBuyValueInteractingCLI, params.DisplayDenom,
				)
			}
			denom := args[2]
			if !strings.EqualFold(denom, params.DisplayDenom) {
				return fmt.Errorf("denom must be %s", strings.ToUpper(params.DisplayDenom))
			}

			buyer := clientCtx.GetFromAddress().String()
			if buyer == "" {
				return fmt.Errorf("flag --%s is required", flags.FlagFrom)
			}

			continueOrderId, err := cmd.Flags().GetString(flagContinueOrderId)
			if err != nil {
				return err
			}
			if continueOrderId != "" && !dymnstypes.IsValidBuyOrderId(continueOrderId) {
				return fmt.Errorf("invalid continue buy-order id: %s", continueOrderId)
			}

			queryClient := dymnstypes.NewQueryClient(clientCtx)

			resParams, err := queryClient.Params(cmd.Context(), &dymnstypes.QueryParamsRequest{})
			if err != nil {
				return err
			}

			msg := &dymnstypes.MsgPlaceBuyOrder{
				AssetId:         dymName,
				AssetType:       dymnstypes.TypeName,
				Buyer:           buyer,
				ContinueOrderId: continueOrderId,
				Offer: sdk.Coin{
					Denom:  resParams.Params.Price.PriceDenom,
					Amount: sdk.NewInt(int64(amount)).MulRaw(adymToDymMultiplier),
				},
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(flagContinueOrderId, "", "if provided, will raise offer value of an existing offer")

	return cmd
}
