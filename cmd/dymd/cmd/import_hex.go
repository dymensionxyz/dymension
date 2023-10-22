package cmd

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"
)

const flagKeyType = "key-type"

func ImportHexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-hex <name> <hex>",
		Short: "Import private keys into the local keybase",
		Long:  fmt.Sprintf("Import hex encoded private key into the local keybase.\nSupported key-types can be obtained with:\n%s list-key-types", version.AppName),
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			keyType, _ := cmd.Flags().GetString(flagKeyType)
			return clientCtx.Keyring.ImportPrivKeyHex(args[0], args[1], keyType)
		},
	}
	return cmd
}
