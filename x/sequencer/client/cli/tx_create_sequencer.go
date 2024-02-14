package cli

import (
	"strconv"

	"encoding/json"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
	"github.com/spf13/cobra"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ = strconv.Itoa(0)

func CmdCreateSequencer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-sequencer [pubkey] [rollapp-id] [description] [bond]",
		Short: "Create a new sequencer for a rollapp",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argPubkey := args[0]
			argRollappId := args[1]
			bond := args[3]
			argDescription := new(types.Description)
			err = json.Unmarshal([]byte(args[2]), argDescription)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var pk cryptotypes.PubKey
			if err := clientCtx.Codec.UnmarshalInterfaceJSON([]byte(argPubkey), &pk); err != nil {
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
				argDescription,
				bondCoin,
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
