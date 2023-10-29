package cli

import (
	"errors"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/x/streamer/types"

	govcli "github.com/cosmos/cosmos-sdk/x/gov/client/cli"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	// cmd := osmocli.TxIndexCmd(types.ModuleName)
	// cmd.AddCommand(
	// 	NewCmdSubmitCreateStreamProposal(),
	// )

	// return cmd
	return nil
}

// NewCreateStreamCmd broadcasts a CreateStream message.
func NewCmdSubmitCreateStreamProposal() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-stream-proposal [dest addr] [reward] [flags]",
		Short: "proposal to create a stream to distribute rewards to a recipient over a period of time",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			txf := tx.NewFactoryCLI(clientCtx, cmd.Flags()).WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)

			distributeTo, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

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

			epochIdentifier, err := cmd.Flags().GetString(FlagEpochIdentifier)
			if err != nil {
				return err
			}

			epochs, err := cmd.Flags().GetUint64(FlagEpochs)
			if err != nil {
				return err
			}

			title, _ := cmd.Flags().GetString(govcli.FlagTitle)
			description, _ := cmd.Flags().GetString(govcli.FlagDescription)
			deposit, _ := cmd.Flags().GetString(govcli.FlagDeposit)

			depositAmt, err := sdk.ParseCoinsNormalized(deposit)
			if err != nil {
				return err
			}

			reqStream := types.Stream{
				DistributeTo:         distributeTo.String(),
				Coins:                coins,
				StartTime:            startTime,
				DistrEpochIdentifier: epochIdentifier,
				NumEpochsPaidOver:    epochs,
			}
			content := types.NewCreateStreamProposal(title, description, reqStream)

			msg, err := govtypes.NewMsgSubmitProposal(content, depositAmt, clientCtx.GetFromAddress())
			if err != nil {
				return err
			}

			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, msg)
		},
	}

	cmd.Flags().String(govcli.FlagTitle, "", "The proposal title")
	cmd.Flags().String(govcli.FlagDescription, "", "The proposal description")
	cmd.Flags().String(govcli.FlagDeposit, "", "The proposal deposit")

	cmd.Flags().AddFlagSet(FlagSetCreateStream())
	return cmd
}
