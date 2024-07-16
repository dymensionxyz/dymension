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
		Use:     "vote [gauge-weights] --from <voter>",
		Short:   "Submit a vote for gauges",
		Example: "dymd tx sponsorship vote gauge1=30,gauge2=40,abstain=30 --from my_validator",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			weights, err := ParseGaugeWeights(args[0])
			if err != nil {
				return fmt.Errorf("invalid gauge weights: %w", err)
			}

			msg := types.MsgVote{
				Voter:   clientCtx.GetFromAddress().String(),
				Weights: weights,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func ParseGaugeWeights(inputWeights string) ([]types.GaugeWeight, error) {
	if inputWeights == "" {
		return nil, fmt.Errorf("input weights must not be empty")
	}

	var weights []types.GaugeWeight
	pairs := strings.Split(inputWeights, ",")

	for _, pair := range pairs {
		idValue := strings.Split(pair, ":")
		if len(idValue) != 2 {
			return nil, fmt.Errorf("invalid gauge weight format: %s", pair)
		}

		gaugeID, err := strconv.ParseUint(idValue[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid gauge ID '%s': %w", idValue[0], err)
		}

		weight, err := strconv.Atoi(idValue[1])
		if err != nil {
			return nil, fmt.Errorf("invalid gauge weight '%s': %w", idValue[1], err)
		}

		if weight < 0 || weight > 100 {
			return nil, fmt.Errorf("weight must be between 0 and 100, got %d", weight)
		}

		weights = append(weights, types.GaugeWeight{
			GaugeId: gaugeID,
			Weight:  math.NewInt(int64(weight)),
		})
	}

	err := types.ValidateGaugeWeights(weights)
	if err != nil {
		return nil, fmt.Errorf("invalid gauge weights: %w", err)
	}

	return weights, nil
}
