package cli

import (
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/spf13/cobra"
)

// NewUpdateResolveDymNameAddressTxCmd returns the CLI command for
// updating the address resolution configuration for a Dym-Name.
func NewUpdateResolveDymNameAddressTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve [Dym-Name address] [?resolve to]",
		Short: "Configure resolve Dym-Name address. 2nd arg if empty means to remove the configuration.",
		Example: fmt.Sprintf(
			"$ %s tx %s resolve bonded.staking@dym dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue --%s hub-user",
			version.AppName, dymnstypes.ModuleName, flags.FlagFrom,
		),
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			var dymNameAddress, resolveTo string
			dymNameAddress = args[0]
			if len(args) > 1 {
				resolveTo = args[1]
			}

			queryClient := dymnstypes.NewQueryClient(clientCtx)

			subName, dymName, chainIdOrAlias, err := dymnskeeper.ParseDymNameAddress(dymNameAddress)
			if err != nil {
				return errorsmod.Wrap(err, "failed to parse input Dym-Name-Address")
			}

			respTranslateChainId, err := queryClient.TranslateAliasOrChainIdToChainId(cmd.Context(), &dymnstypes.QueryTranslateAliasOrChainIdToChainIdRequest{
				AliasOrChainId: chainIdOrAlias,
			})
			if err != nil || respTranslateChainId.ChainId == "" {
				return errorsmod.Wrapf(err, "failed to translate alias to chain-id: %s", chainIdOrAlias)
			}

			chainId := respTranslateChainId.ChainId
			fmt.Printf("Translated '%s' => '%s'\n", chainIdOrAlias, chainId)
			time.Sleep(5 * time.Second)

			if !dymnsutils.IsValidChainIdFormat(chainId) {
				return fmt.Errorf("input chain-id '%s' is not a valid chain-id", chainId)
			}

			controller := clientCtx.GetFromAddress().String()

			if controller == "" {
				return fmt.Errorf("flag --%s is required", flags.FlagFrom)
			}

			msg := &dymnstypes.MsgUpdateResolveAddress{
				Name:       dymName,
				ChainId:    chainId,
				SubName:    subName,
				ResolveTo:  resolveTo,
				Controller: controller,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
