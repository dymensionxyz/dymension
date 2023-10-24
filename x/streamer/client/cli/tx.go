package cli

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/x/streamer/types"
	"github.com/osmosis-labs/osmosis/v15/osmoutils/osmocli"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := osmocli.TxIndexCmd(types.ModuleName)
	cmd.AddCommand(
	// NewCreateStreamCmd(),
	)

	return cmd
}

// NewCreateStreamCmd broadcasts a CreateStream message.
// func NewCreateStreamCmd() *cobra.Command {
// 	cmd := &cobra.Command{
// 		Use:   "create-gauge [lockup_denom] [reward] [flags]",
// 		Short: "create a gauge to distribute rewards to users",
// 		Args:  cobra.ExactArgs(2),
// 		RunE: func(cmd *cobra.Command, args []string) error {
// 			clientCtx, err := client.GetClientTxContext(cmd)
// 			if err != nil {
// 				return err
// 			}

// 			denom := args[0]

// 			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)
// 			coins, err := sdk.ParseCoinsNormalized(args[1])
// 			if err != nil {
// 				return err
// 			}

// 			var startTime time.Time
// 			timeStr, err := cmd.Flags().GetString(FlagStartTime)
// 			if err != nil {
// 				return err
// 			}
// 			if timeStr == "" { // empty start time
// 				startTime = time.Unix(0, 0)
// 			} else if timeUnix, err := strconv.ParseInt(timeStr, 10, 64); err == nil { // unix time
// 				startTime = time.Unix(timeUnix, 0)
// 			} else if timeRFC, err := time.Parse(time.RFC3339, timeStr); err == nil { // RFC time
// 				startTime = timeRFC
// 			} else { // invalid input
// 				return errors.New("invalid start time format")
// 			}

// 			epochs, err := cmd.Flags().GetUint64(FlagEpochs)
// 			if err != nil {
// 				return err
// 			}

// 			perpetual, err := cmd.Flags().GetBool(FlagPerpetual)
// 			if err != nil {
// 				return err
// 			}

// 			if perpetual {
// 				epochs = 1
// 			}

// 			msg := types.NewMsgCreateStream(
// 				// epochs == 1,
// 				// clientCtx.GetFromAddress(),
// 				distributeTo,
// 				coins,
// 				startTime,
// 				//TODO: fix epoch identifier
// 				"day",
// 				epochs,
// 			)

// 			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
// 		},
// 	}

// 	cmd.Flags().AddFlagSet(FlagSetCreateStream())
// 	flags.AddTxFlagsToCmd(cmd)
// 	return cmd
// }
