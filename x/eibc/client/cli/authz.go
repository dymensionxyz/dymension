package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

const (
	FlagSpendLimit          = "spend-limit"
	FlagExpiration          = "expiration"
	FlagRollapps            = "rollapps"
	FlagDenoms              = "denoms"
	FlagMinLPFeePercentage  = "min-lp-fee-percentage"
	FlagMaxPrice            = "max-price"
	FlagOperatorFeeShare    = "operator-fee-share"
	FlagSettlementValidated = "settlement-validated"
)

// NewCmdGrantAuthorization returns a CLI command handler for creating a MsgGrant transaction.
func NewCmdGrantAuthorization() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant <grantee> --from <granter>",
		Short: "Grant authorization to an address",
		Long: strings.TrimSpace(
			fmt.Sprintf(`create a new grant authorization to an address to execute a transaction on your behalf:

Examples:
 $ %s tx %s grant dym1skjw.. --spend-limit=1000stake... --from=dym1skl..`, version.AppName, authz.ModuleName),
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return fmt.Errorf("failed to get client context: %w", err)
			}

			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return fmt.Errorf("failed to parse grantee address: %w", err)
			}

			rollapps, err := cmd.Flags().GetStringSlice(FlagRollapps)
			if err != nil {
				return fmt.Errorf("failed to get rollapps: %w", err)
			}
			denoms, err := cmd.Flags().GetStringSlice(FlagDenoms)
			if err != nil {
				return fmt.Errorf("failed to get denoms: %w", err)
			}

			minFeeStr, err := cmd.Flags().GetString(FlagMinLPFeePercentage)
			if err != nil {
				return fmt.Errorf("failed to get min fee: %w", err)
			}

			minFeePercDec, err := sdk.NewDecFromStr(minFeeStr)
			if err != nil {
				return fmt.Errorf("invalid min lp fee percentage: %w", err)
			}
			minLPFeePercent := sdk.DecProto{Dec: minFeePercDec}

			maxPriceStr, err := cmd.Flags().GetString(FlagMaxPrice)
			if err != nil {
				return fmt.Errorf("failed to get max price: %w", err)
			}

			maxPrice, err := sdk.ParseCoinsNormalized(maxPriceStr)
			if err != nil {
				return fmt.Errorf("failed to parse max price: %w", err)
			}

			fulfillerFeePartStr, err := cmd.Flags().GetString(FlagOperatorFeeShare)
			if err != nil {
				return fmt.Errorf("failed to get fulfiller fee part: %w", err)
			}

			fulfillerFeePartDec, err := sdk.NewDecFromStr(fulfillerFeePartStr)
			if err != nil {
				return fmt.Errorf("failed to parse fulfiller fee part: %w", err)
			}
			fulfillerFeePart := sdk.DecProto{Dec: fulfillerFeePartDec}

			settlementValidated, err := cmd.Flags().GetBool(FlagSettlementValidated)
			if err != nil {
				return fmt.Errorf("failed to get settlement validated: %w", err)
			}

			limit, err := cmd.Flags().GetString(FlagSpendLimit)
			if err != nil {
				return fmt.Errorf("failed to get spend limit: %w", err)
			}

			var spendLimit sdk.Coins
			if limit != "" {
				spendLimit, err = sdk.ParseCoinsNormalized(limit)
				if err != nil {
					return fmt.Errorf("failed to parse spend limit: %w", err)
				}
				spendLimit, err := sdk.ParseCoinsNormalized(limit)
				if err != nil {
					return fmt.Errorf("failed to parse spend limit: %w", err)
				}

				if !spendLimit.IsAllPositive() {
					return fmt.Errorf("spend-limit should be greater than zero")
				}
			}

			authorization := types.NewFulfillOrderAuthorization(
				rollapps,
				denoms,
				minLPFeePercent,
				maxPrice,
				fulfillerFeePart,
				settlementValidated,
				spendLimit,
			)

			expire, err := getExpireTime(cmd)
			if err != nil {
				return fmt.Errorf("failed to get expiration time: %w", err)
			}

			msg, err := authz.NewMsgGrant(clientCtx.GetFromAddress(), grantee, authorization, expire)
			if err != nil {
				return fmt.Errorf("failed to create MsgGrant: %w", err)
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().StringSlice(FlagRollapps, []string{}, "An array of Rollapp IDs allowed")
	cmd.Flags().StringSlice(FlagDenoms, []string{}, "An array of denoms allowed to use")
	cmd.Flags().String(FlagSpendLimit, "", "An array of Coins allowed to spend")
	cmd.Flags().Bool(FlagSettlementValidated, false, "Settlement validated flag")
	cmd.Flags().String(FlagMinLPFeePercentage, "", "Minimum fee")
	cmd.Flags().String(FlagMaxPrice, "", "Maximum price")
	cmd.Flags().String(FlagOperatorFeeShare, "", "Fulfiller fee part")
	cmd.Flags().Int64(FlagExpiration, 0, "Expire time as Unix timestamp. Set zero (0) for no expiry. Default is 0.")
	return cmd
}

func getExpireTime(cmd *cobra.Command) (*time.Time, error) {
	exp, err := cmd.Flags().GetInt64(FlagExpiration)
	if err != nil {
		return nil, err
	}
	if exp == 0 {
		return nil, nil
	}
	e := time.Unix(exp, 0)
	return &e, nil
}
