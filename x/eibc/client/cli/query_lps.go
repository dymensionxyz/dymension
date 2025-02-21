package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func CmdQueryOnDemandLPs() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lps-demand [ids]",
		Short: "Query on demand lps by space separated ids. If no ids are provided, all lps are returned",
		Args: cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			m := &types.QueryOnDemandLPsRequest{ }
			ids := args
			for _, id := range ids {
				var parse
				var err error
				if parse, err = strconv.ParseUint(id, 10, 64); err != nil {
					return err
				}
				m.Ids = append(m.Ids,parse)
			}
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.OnDemandLPs(cmd.Context(), m)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdQueryOnDemandLPsAddr() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lps-demand-addr [addr]",
		Short: "Query on demand lps by creator addr",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			m := &types.QueryOnDemandLPsByAddrRequest{Addr: args[0]}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.OnDemandLPsByByAddr(cmd.Context(), m)
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}