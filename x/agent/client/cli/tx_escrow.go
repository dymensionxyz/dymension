package cli

import (
	"strconv"
	"strings"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func CmdFundAgentEscrow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fund-escrow [agent-id] [amount]",
		Short: "Deposit funds into an agent's escrow (anyone may fund)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgFundAgentEscrow(clientCtx.GetFromAddress().String(), args[0], amount)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdWithdrawAgentEscrow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-escrow [agent-id] [amount]",
		Short: "Withdraw funds from an agent's escrow (owner only)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			amount, err := sdk.ParseCoinsNormalized(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdrawAgentEscrow(clientCtx.GetFromAddress().String(), args[0], amount)
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdUpdateAgentSpendPolicy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update-spend-policy [agent-id] [spend-denom] [spend-limit-per-window] [spend-window-blocks] [recipient-allowlist]",
		Short: "Set an agent's spend policy (owner only, effective immediately)",
		Long:  "Set an agent's spend policy. Pass recipient-allowlist as comma-separated addresses, or '' for unrestricted spending. Pass an empty spend-denom ('') with limit 0 and window 0 to disable spending.",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg, err := newMsgUpdateAgentSpendPolicy(clientCtx.GetFromAddress().String(), args)
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newMsgUpdateAgentSpendPolicy(owner string, args []string) (*types.MsgUpdateAgentSpendPolicy, error) {
	limit, ok := math.NewIntFromString(args[2])
	if !ok {
		return nil, gerrc.ErrInvalidArgument.Wrap("spend limit per window")
	}
	windowBlocks, err := strconv.ParseUint(args[3], 10, 64)
	if err != nil {
		return nil, err
	}
	msg := types.NewMsgUpdateAgentSpendPolicy(owner, args[0], args[1], limit, windowBlocks)
	if args[4] != "" {
		msg.RecipientAllowlist = strings.Split(args[4], ",")
	}
	return msg, nil
}

func CmdSubmitAttestedTransfer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit-attested-transfer [agent-id] [recipient] [amount] [memo] [token]",
		Short: "Submit an enclave-attested transfer from an agent's escrow",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg, err := newMsgSubmitAttestedTransfer(clientCtx.GetFromAddress().String(), args)
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func newMsgSubmitAttestedTransfer(submitter string, args []string) (*types.MsgSubmitAttestedTransfer, error) {
	amount, ok := math.NewIntFromString(args[2])
	if !ok {
		return nil, gerrc.ErrInvalidArgument.Wrap("amount")
	}
	return &types.MsgSubmitAttestedTransfer{
		Submitter: submitter,
		AgentId:   args[0],
		Recipient: args[1],
		Amount:    amount,
		Memo:      []byte(args[3]),
		Token:     args[4],
	}, nil
}
