package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func CmdQueryTeeConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tee-config",
		Short: "shows TEE configuration parameters in a readable format",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			teeConfig := res.Params.TeeConfig

			fmt.Println("TEE Configuration")
			fmt.Println("=================")
			fmt.Println()

			fmt.Printf("Enabled: %v\n", teeConfig.Enabled)
			fmt.Printf("Verify:  %v\n", teeConfig.Verify)
			fmt.Println()

			if teeConfig.PolicyValues != "" {
				fmt.Println("Policy Values:")
				fmt.Println("-------------")
				var policyValues map[string]interface{}
				if err := json.Unmarshal([]byte(teeConfig.PolicyValues), &policyValues); err == nil {
					prettyJSON, _ := json.MarshalIndent(policyValues, "", "  ")
					fmt.Println(string(prettyJSON))
				} else {
					fmt.Println(teeConfig.PolicyValues)
				}
				fmt.Println()
			}

			if teeConfig.PolicyStructure != "" {
				fmt.Println("Policy Structure:")
				fmt.Println("----------------")
				fmt.Println(teeConfig.PolicyStructure)
				fmt.Println()
			}

			if teeConfig.PolicyQuery != "" {
				fmt.Println("Policy Query:")
				fmt.Println("------------")
				fmt.Println(teeConfig.PolicyQuery)
				fmt.Println()
			}

			if teeConfig.GcpRootCertPem != "" {
				fmt.Println("GCP Root Certificate:")
				fmt.Println("--------------------")
				certLines := strings.Split(teeConfig.GcpRootCertPem, "\n")
				if len(certLines) > 0 {
					fmt.Println(certLines[0])
					if len(certLines) > 3 {
						fmt.Printf("... (%d lines)\n", len(certLines)-2)
					}
					if len(certLines) > 1 {
						fmt.Println(certLines[len(certLines)-1])
					}
				}
				fmt.Println()
			}

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
