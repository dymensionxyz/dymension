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

// NewRegisterAliasTxCmd is the CLI command for registering a new Alias for an owned RollApp.
func NewRegisterAliasTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "register-alias [Alias/Handle] [RollApp ID]",
		Aliases: []string{"register-handle"},
		Short:   "Register a new Alias/Handle for the owned Roll App",
		Example: fmt.Sprintf(
			"$ %s tx %s register-alias rolx rollappx_1-1 --confirm-payment 15000000000000000000%s --%s hub-user",
			version.AppName, dymnstypes.ModuleName,
			params.BaseDenom,
			flags.FlagFrom,
		),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			alias := args[0]
			if !dymnsutils.IsValidAlias(alias) {
				return fmt.Errorf("input is not a valid alias: %s", alias)
			}

			rollAppId := args[1]
			if !dymnsutils.IsValidChainIdFormat(rollAppId) {
				return fmt.Errorf("input is not a valid RollApp ID: %s", rollAppId)
			}

			rollAppOwnerAsBuyer := clientCtx.GetFromAddress().String()

			if rollAppOwnerAsBuyer == "" {
				return fmt.Errorf("flag --%s is required", flags.FlagFrom)
			}

			confirmPaymentStr, err := cmd.Flags().GetString(flagConfirmPayment)
			if err != nil {
				return err
			}
			if confirmPaymentStr == "" {
				// mode query to get the estimated payment amount
				queryClient := dymnstypes.NewQueryClient(clientCtx)

				resEst, err := queryClient.EstimateRegisterAlias(cmd.Context(), &dymnstypes.EstimateRegisterAliasRequest{
					Alias:     alias,
					RollappId: rollAppId,
					Owner:     rollAppOwnerAsBuyer,
				})
				if err != nil {
					return fmt.Errorf("failed to estimate registration fee for '%s': %w", alias, err)
				}

				fmt.Println("Estimated registration fee: ", resEst.Price)
				if estAmt, ok := toEstimatedCoinAmount(resEst.Price); ok {
					fmt.Printf("  (~ %s)\n", estAmt)
				}

				fmt.Printf("Supplying flag '--%s=%s' to submit the registration\n", flagConfirmPayment, resEst.Price.String())

				return nil
			}

			confirmPayment, err := sdk.ParseCoinNormalized(confirmPaymentStr)
			if err != nil {
				return fmt.Errorf("invalid confirm payment: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &dymnstypes.MsgRegisterAlias{
				Alias:          alias,
				RollappId:      rollAppId,
				Owner:          rollAppOwnerAsBuyer,
				ConfirmPayment: confirmPayment,
			})
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	cmd.Flags().String(flagConfirmPayment, "", "confirm payment for the Alias/Handle registration, without this flag, the command will query the estimated payment amount")

	return cmd
}
