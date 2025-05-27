package cli

import (
	"errors"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils/osmocli"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/incentives/types"
	lockuptypes "github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := osmocli.TxIndexCmd(types.ModuleName)
	cmd.AddCommand(
		NewCreateAssetGaugeCmd(),
		NewCreateEndorsementGaugeCmd(),
		NewAddToGaugeCmd(),
	)

	return cmd
}

// TODO: add flag to use lock age instead/with duration
// FIXME: change to create-asset-gauge

// NewCreateAssetGaugeCmd broadcasts a CreateGauge message.
func NewCreateAssetGaugeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-gauge [lockup_denom] [reward] [flags]",
		Short: "create a gauge to distribute rewards to users",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			denom := args[0]

			txfCli, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			txf := txfCli.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)
			coins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			var startTime time.Time
			timeStr, err := cmd.Flags().GetString(FlagStartTime)
			if err != nil {
				return err
			}
			if timeStr == "" { // empty start time
				startTime = time.Unix(0, 0)
			} else if timeUnix, err := strconv.ParseInt(timeStr, 10, 64); err == nil { // unix time
				startTime = time.Unix(timeUnix, 0)
			} else if timeRFC, err := time.Parse(time.RFC3339, timeStr); err == nil { // RFC time
				startTime = timeRFC
			} else { // invalid input
				return errors.New("invalid start time format")
			}

			epochs, err := cmd.Flags().GetUint64(FlagEpochs)
			if err != nil {
				return err
			}

			perpetual, err := cmd.Flags().GetBool(FlagPerpetual)
			if err != nil {
				return err
			}

			if perpetual {
				epochs = 1
			}

			duration, err := cmd.Flags().GetDuration(FlagDuration)
			if err != nil {
				return err
			}

			distributeTo := lockuptypes.QueryCondition{
				Denom:    denom,
				Duration: duration,
			}

			msg := types.MsgCreateGauge{
				IsPerpetual:       epochs == 1,
				GaugeType:         types.GaugeType_GAUGE_TYPE_ASSET,
				Asset:             &distributeTo,
				Coins:             coins,
				StartTime:         startTime,
				NumEpochsPaidOver: epochs,
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, &msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetCreateGauge())
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

// NewCreateEndorsementGaugeCmd broadcasts a CreateGauge message for endorsement gauges.
func NewCreateEndorsementGaugeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-endorsement-gauge [rollapp_id] [reward] [flags]",
		Short: "create an endorsement gauge to distribute rewards to rollapp endorsers",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			rollappID := args[0]

			txfCli, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			txf := txfCli.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)
			coins, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			var startTime time.Time
			timeStr, err := cmd.Flags().GetString(FlagStartTime)
			if err != nil {
				return err
			}
			if timeStr == "" { // empty start time
				startTime = time.Unix(0, 0)
			} else if timeUnix, err := strconv.ParseInt(timeStr, 10, 64); err == nil { // unix time
				startTime = time.Unix(timeUnix, 0)
			} else if timeRFC, err := time.Parse(time.RFC3339, timeStr); err == nil { // RFC time
				startTime = timeRFC
			} else { // invalid input
				return errors.New("invalid start time format")
			}

			epochs, err := cmd.Flags().GetUint64(FlagEpochs)
			if err != nil {
				return err
			}

			perpetual, err := cmd.Flags().GetBool(FlagPerpetual)
			if err != nil {
				return err
			}

			if perpetual {
				epochs = 1
			}

			msg := types.MsgCreateGauge{
				IsPerpetual:       epochs == 1,
				GaugeType:         types.GaugeType_GAUGE_TYPE_ENDORSEMENT,
				Endorsement:       &types.EndorsementGauge{RollappId: rollappID},
				Coins:             coins,
				StartTime:         startTime,
				NumEpochsPaidOver: epochs,
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, &msg)
		},
	}

	cmd.Flags().AddFlagSet(FlagSetCreateGauge())
	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewAddToGaugeCmd() *cobra.Command {
	return osmocli.BuildTxCli[*types.MsgAddToGauge](&osmocli.TxCliDesc{
		Use:   "add-to-gauge [gauge_id] [rewards] [flags]",
		Short: "add coins to gauge to distribute more rewards to users",
	})
}
