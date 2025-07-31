package cli

import (
	"github.com/bcp-innovations/hyperlane-cosmos/util"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/utils"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
)

func NewCmdRegisterHLTokenDenomMetadata() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "register-hl-token-denom-metadata [hl_token_id] [denom_metadata.json]",
		Short:   "Register denom metadata as owner of a HL token",
		Example: `dymd tx denommetadata register-hl-token-denom-metadata 0x934b867052ca9c65e33362112f35fb548f8732c2fe45f07b9c591958e865def0 denom_metadata.json`,
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			hlTokenId, err := util.DecodeHexAddress(args[0])
			if err != nil {
				return err
			}

			path := args[1]

			var metadata banktypes.Metadata
			err = utils.ParseJsonFromFile(path, &metadata)
			if err != nil {
				return err
			}
			msg := types.MsgRegisterHLTokenDenomMetadata{
				HlTokenId:     hlTokenId,
				HlTokenOwner:  clientCtx.GetFromAddress().String(),
				TokenMetadata: metadata,
			}

			txfCli, err := tx.NewFactoryCLI(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			txf := txfCli.WithTxConfig(clientCtx.TxConfig).WithAccountRetriever(clientCtx.AccountRetriever)
			return tx.GenerateOrBroadcastTxWithFactory(clientCtx, txf, &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
