package cli

import (
	"fmt"
	"strings"

	sdkmath "cosmossdk.io/math"

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
	flagYears          = "years"
	flagConfirmPayment = "confirm-payment"
	flagContact        = "contact"
)

// NewRegisterDymNameTxCmd is the CLI command for registering a new Dym-Name or extending the duration of an owned Dym-Name.
func NewRegisterDymNameTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "register-name [Dym-Name]",
		Aliases: []string{"register"},
		Short:   "Register a new Dym-Name or Extends the duration of an owned Dym-Name.",
		Example: fmt.Sprintf(
			"$ %s tx %s register myname --years 3 --confirm-payment 15000000000000000000%s --%s hub-user",
			version.AppName, dymnstypes.ModuleName,
			params.BaseDenom,
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
				return fmt.Errorf("input is not a valid Dym-Name: %s", dymName)
			}

			years, err := cmd.Flags().GetInt64(flagYears)
			if err != nil {
				return err
			}
			if years < 1 {
				return fmt.Errorf("years must be greater than 0, specify by flag --%s", flagYears)
			}

			buyer := clientCtx.GetFromAddress().String()

			if buyer == "" {
				return fmt.Errorf("flag --%s is required", flags.FlagFrom)
			}

			confirmPaymentStr, err := cmd.Flags().GetString(flagConfirmPayment)
			if err != nil {
				return err
			}
			if confirmPaymentStr == "" {
				// mode query to get the estimated payment amount
				queryClient := dymnstypes.NewQueryClient(clientCtx)

				resEst, err := queryClient.EstimateRegisterName(cmd.Context(), &dymnstypes.EstimateRegisterNameRequest{
					Name:     dymName,
					Duration: years,
					Owner:    buyer,
				})
				if err != nil {
					return fmt.Errorf("failed to estimate registration/renew fee for '%s': %w", dymName, err)
				}

				fmt.Println("Estimated payment amount:")
				if resEst.FirstYearPrice.IsNil() || resEst.FirstYearPrice.IsZero() {
					fmt.Println("- Registration fee: None")
				} else {
					fmt.Println("- Registration fee + first year fee: ", resEst.FirstYearPrice)
					if estAmt, ok := toEstimatedCoinAmount(resEst.FirstYearPrice); ok {
						fmt.Printf("  (~ %s)\n", estAmt)
					}
				}
				fmt.Print("- Extends duration fee: ")
				if resEst.ExtendPrice.IsNil() || resEst.ExtendPrice.IsZero() {
					fmt.Println("None")
				} else {
					fmt.Println(resEst.ExtendPrice)
					if estAmt, ok := toEstimatedCoinAmount(resEst.ExtendPrice); ok {
						fmt.Printf("  (~ %s)\n", estAmt)
					}
				}
				fmt.Println("- Total fee: ", resEst.TotalPrice)
				if estAmt, ok := toEstimatedCoinAmount(resEst.TotalPrice); ok {
					fmt.Printf("  (~ %s)\n", estAmt)
				}

				fmt.Printf("Supplying flag '--%s=%s' to submit the registration\n", flagConfirmPayment, resEst.TotalPrice.String())

				return nil
			}

			confirmPayment, err := sdk.ParseCoinNormalized(confirmPaymentStr)
			if err != nil {
				return fmt.Errorf("invalid confirm payment: %w", err)
			}

			contact, err := cmd.Flags().GetString(flagContact)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &dymnstypes.MsgRegisterName{
				Name:           dymName,
				Duration:       years,
				Owner:          buyer,
				ConfirmPayment: confirmPayment,
				Contact:        contact,
			})
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().Int64(flagYears, 0, "number of years to register the Dym-Name for")
	cmd.Flags().String(flagConfirmPayment, "", "confirm payment for the Dym-Name registration, without this flag, the command will query the estimated payment amount")
	cmd.Flags().String(flagContact, "", "contact information for the Dym-Name")

	return cmd
}

func toEstimatedAmount(amount sdkmath.Int) string {
	return fmt.Sprintf("%s %s", amount.QuoRaw(adymToDymMultiplier), strings.ToUpper(params.DisplayDenom))
}

func toEstimatedCoinAmount(amount sdk.Coin) (estimatedAmount string, success bool) {
	if amount.Denom == params.BaseDenom {
		return toEstimatedAmount(amount.Amount), true
	} else {
		return amount.String(), false
	}
}
