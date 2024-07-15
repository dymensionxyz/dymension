package cli

import (
	"fmt"
	"strconv"
	"strings"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdVote())

	return cmd
}

func CmdVote() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "vote [gauges] --from <voter>",
		Short:   "Submit a vote for gauges",
		Example: "dymd tx sponsorship vote gauge1=30,gauge2=40,abstain=30",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			msg := types.MsgVote{
				Voter:   clientCtx.GetFromAddress().String(),
				Weights: nil,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func ParseGaugeWeights(weights string) ([]types.GaugeWeight, error) {
	var result []types.GaugeWeight
	pairs := strings.Split(weights, ",")

	var totalWeight int

	for _, pair := range pairs {
		idValue := strings.Split(pair, ":")
		if len(idValue) != 2 {
			return nil, fmt.Errorf("invalid input")
		}

		gaugeID, err := strconv.ParseUint(idValue[0], 10, 64)
		if err != nil {
			return nil, err
		}

		weight, err := strconv.Atoi(idValue[1])
		if err != nil {
			return nil, err
		}

		if weight < 0 || weight > 100 {
			return nil, fmt.Errorf("weight must be between 0 and 100")
		}

		totalWeight += weight

		result = append(result, types.GaugeWeight{
			GaugeId: gaugeID,
			Weight:  math.NewInt(int64(weight)),
		})
	}

	if totalWeight != 100 {
		return nil, fmt.Errorf("sum of all weights must be 100, got %d", totalWeight)
	}

	return result, nil
}
