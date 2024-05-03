package cli

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func CmdListDemandOrdersByStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-demand-orders [status] [recipient] [denom] [fulfilled] [ids] [count]",
		Short: "List all demand orders with a specific status",
		Long:  `Query demand orders filtered by status. Examples of status include "pending", "finalized", and "reverted".`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			status := args[0]

			queryClient := types.NewQueryClient(clientCtx)
			res, err := queryClient.DemandOrdersByStatus(context.Background(), &types.QueryDemandOrdersByStatusRequest{
				Status: status,
			})
			if err != nil {
				return err
			}

			recipient, _ := cmd.Flags().GetString("recipient")
			denom, _ := cmd.Flags().GetString("denom")
			fulfilledFlag, _ := cmd.Flags().GetString("fulfilled")
			showCountFlag, _ := cmd.Flags().GetString("count")
			idsStr, _ := cmd.Flags().GetString("ids")

			var (
				filterIDs bool
				ids       []string
			)
			if len(idsStr) > 0 {
				ids = strings.Split(idsStr, ",")
				filterIDs = true
			}

			var fulfilled *bool
			if fulfilledFlag != "" {
				f := fulfilledFlag == "true"
				fulfilled = &f
			}

			var showCount *bool
			if showCountFlag != "" {
				c := showCountFlag == "true"
				showCount = &c
			}

			count := 0

			for _, o := range res.DemandOrders {
				if !o.Price.AmountOf(denom).IsPositive() {
					continue
				}

				if filterIDs && !slices.Contains(ids, o.Id) {
					continue
				}

				if fulfilled != nil && *fulfilled != o.IsFullfilled {
					continue
				}
				if recipient != "" && o.Recipient != recipient {
					continue
				}
				count++

				if showCount != nil && *showCount {
					continue
				}

				fmt.Printf(`
-id:		%s
  recipient:	%s
  price:	%s
  status:	%s
  packet_key:	%s
  fee:		%s
  is_fulfilled:	%v
`,
					o.Id, o.Recipient, o.Price.String(), o.TrackingPacketStatus.String(),
					o.TrackingPacketKey, o.Fee.String(), o.IsFullfilled)
			}

			fmt.Printf("\ncount: %d; ts: %s\n", count, time.Now().String())

			return nil
		},
	}

	cmd.Flags().StringP("recipient", "r", "", "Recipient address")
	cmd.Flags().StringP("denom", "d", "", "Denom")
	cmd.Flags().StringP("fulfilled", "f", "", "Filter by fulfillment status")
	cmd.Flags().StringP("ids", "i", "", "Filter by order IDs")
	cmd.Flags().StringP("count", "c", "", "Only show the count")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
