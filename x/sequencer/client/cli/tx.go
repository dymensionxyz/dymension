package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/utils"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdCreateSequencer())
	cmd.AddCommand(CmdUpdateSequencer())
	cmd.AddCommand(CmdUpdateRewardAddress())
	cmd.AddCommand(CmdUpdateWhitelistedRelayers())
	cmd.AddCommand(CmdUnbond())
	cmd.AddCommand(CmdIncreaseBond())
	cmd.AddCommand(CmdDecreaseBond())

	return cmd
}

const (
	FlagRewardAddress       = "reward-address"
	FlagWhitelistedRelayers = "whitelisted-relayers"
)

func CmdCreateSequencer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-sequencer [pubkey] [rollapp-id] [bond] [metadata] --reward-address [reward_addr] --whitelisted-relayers [addr1,addr2,addr3]",
		Short: "Create a new sequencer for a rollapp",
		Long: `Create a new sequencer for a rollapp. 
Metadata is an optional arg.
Reward address is an optional flag-arg. It expects a bech32-encoded address which will be used as a sequencer's reward address.
Whitelisted relayers is an optional flag-arg. It expects a comma-separated list of bech32-encoded addresses.`,
		Args: cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argPubkey := args[0]
			argRollappId := args[1]
			bond := args[2]

			var metadata types.SequencerMetadata
			if len(args) == 4 {
				if err = utils.ParseJsonFromFile(args[3], &metadata); err != nil {
					return
				}
			}

			rewardAddr, _ := cmd.Flags().GetString(FlagRewardAddress)

			var whitelistedRelayers []string
			if relayers, _ := cmd.Flags().GetString(FlagWhitelistedRelayers); relayers != "" {
				whitelistedRelayers = strings.Split(relayers, ",")
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var pk cryptotypes.PubKey

			if err = clientCtx.Codec.UnmarshalInterfaceJSON([]byte(argPubkey), &pk); err != nil {
				return err
			}

			bondCoin, err := sdk.ParseCoinNormalized(bond)
			if err != nil {
				return err
			}

			msg, err := types.NewMsgCreateSequencer(
				clientCtx.GetFromAddress().String(),
				pk,
				argRollappId,
				&metadata,
				bondCoin,
				rewardAddr,
				whitelistedRelayers,
			)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(FlagRewardAddress, "", "Sequencer reward address")
	cmd.Flags().String(FlagWhitelistedRelayers, "", "Sequencer whitelisted relayers")

	return cmd
}

func CmdUpdateSequencer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-sequencer [metadata]",
		Short: "Update a sequencer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			metadata := new(types.SequencerMetadata)
			if err = utils.ParseJsonFromFile(args[0], metadata); err != nil {
				return
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg, err := types.NewMsgUpdateSequencerInformation(
				clientCtx.GetFromAddress().String(),
				metadata,
			)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdUpdateRewardAddress() *cobra.Command {
	short := "Update a sequencer reward address"
	cmd := &cobra.Command{
		Use:     "update-reward-address [addr]",
		Example: "update-reward-address ethm1lhk5cnfrhgh26w5r6qft36qerg4dclfev9nprc --from foouser",
		Args:    cobra.ExactArgs(1),
		Short:   short,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgUpdateRewardAddress{
				Creator:    sdk.ValAddress(ctx.GetFromAddress()).String(),
				RewardAddr: args[0],
			}

			return tx.GenerateOrBroadcastTxCLI(ctx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdUpdateWhitelistedRelayers() *cobra.Command {
	short := "Update a sequencer whitelisted relayer list"
	cmd := &cobra.Command{
		Use:     "update-whitelisted-relayers [relayers]",
		Example: "update-whitelisted-relayers ethm1lhk5cnfrhgh26w5r6qft36qerg4dclfev9nprc,ethm1lhasdf8969asdfgj2g3j4,ethmasdfkjhgjkhg123j4hgasv7ghi4v --from foouser",
		Args:    cobra.ExactArgs(1),
		Short:   short,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgUpdateWhitelistedRelayers{
				Creator:  sdk.ValAddress(ctx.GetFromAddress()).String(),
				Relayers: strings.Split(args[0], ","),
			}

			return tx.GenerateOrBroadcastTxCLI(ctx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdUnbond() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbond",
		Short: "Unbond the sequencer",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUnbond(
				clientCtx.GetFromAddress().String(),
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdIncreaseBond() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "increase-bond [amount]",
		Short: "Increase the bond of a sequencer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			amount := args[0]
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amountCoin, err := sdk.ParseCoinNormalized(amount)
			if err != nil {
				return err
			}

			msg := types.NewMsgIncreaseBond(
				clientCtx.GetFromAddress().String(),
				amountCoin,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdDecreaseBond() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decrease-bond [amount]",
		Short: "Decrease the bond of a sequencer",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			amount := args[0]
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amountCoin, err := sdk.ParseCoinNormalized(amount)
			if err != nil {
				return err
			}

			msg := types.NewMsgDecreaseBond(
				clientCtx.GetFromAddress().String(),
				amountCoin,
			)

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
