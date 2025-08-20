package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

// TODO: rename file

func CmdListDemandOrdersByStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-demand-orders [status]",
		Short: "Query eIBC demand orders by status with optional filters",
		Long: `Query eIBC demand orders filtered by status and various optional criteria.

Required argument:
  status - Order status: "pending" (awaiting fulfillment), "finalized" (completed), or "reverted" (rolled back)

Optional filters (use flags):
  --rollapp    Filter by RollApp ID
  --recipient  Filter by recipient address
  --type       Filter by packet type (recv, timeout, or ack)
  --denom      Filter by token denomination
  --fulfilled  Filter by fulfillment state (fulfilled or unfulfilled)
  --fulfiller  Filter by fulfiller address
  --limit      Maximum number of orders to return`,
		Example: `# List all pending eIBC orders
dymd query eibc list-demand-orders pending

# List pending orders for a specific RollApp
dymd query eibc list-demand-orders pending --rollapp rollapp_1234-1

# List finalized orders that were fulfilled
dymd query eibc list-demand-orders finalized --fulfilled fulfilled

# List pending orders for a specific recipient with limit
dymd query eibc list-demand-orders pending --recipient dym1abc... --limit 10

# List pending orders for specific packet type and denom
dymd query eibc list-demand-orders pending --type recv --denom adym

# List orders fulfilled by a specific address
dymd query eibc list-demand-orders finalized --fulfiller dym1xyz...`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			status, ok := commontypes.Status_value[strings.ToUpper(args[0])]
			if !ok {
				return fmt.Errorf("invalid status: %s", args[0])
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			request := &types.QueryDemandOrdersByStatusRequest{
				Status:     commontypes.Status(status),
				Type:       commontypes.RollappPacket_UNDEFINED, // default to undefined, as '0' is a valid type
				Pagination: pageReq,
			}

			request.RollappId, err = cmd.Flags().GetString("rollapp")
			if err != nil {
				return err
			}

			request.Recipient, err = cmd.Flags().GetString("recipient")
			if err != nil {
				return err
			}

			packetType, err := cmd.Flags().GetString("type")
			if err != nil {
				return err
			}
			if packetType != "" {
				packetType = strings.ToUpper(packetType)
				if !strings.HasPrefix(packetType, "ON_") {
					packetType = "ON_" + packetType
				}
				ptype, ok := commontypes.RollappPacket_Type_value[packetType]
				if !ok {
					return fmt.Errorf("invalid packet type: %s", packetType)
				}
				request.Type = commontypes.RollappPacket_Type(ptype)
			}

			request.Denom, err = cmd.Flags().GetString("denom")
			if err != nil {
				return err
			}

			fulfilled, err := cmd.Flags().GetString("fulfilled")
			if err != nil {
				return err
			}
			if fulfilled != "" {
				fulfillmentState, ok := types.FulfillmentState_value[strings.ToUpper(fulfilled)]
				if !ok {
					return fmt.Errorf("invalid fulfillment state: %s", fulfilled)
				}
				request.FulfillmentState = types.FulfillmentState(fulfillmentState)
			}

			request.Fulfiller, err = cmd.Flags().GetString("fulfiller")
			if err != nil {
				return err
			}

			request.Limit, err = cmd.Flags().GetInt32("limit")
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.DemandOrdersByStatus(cmd.Context(), request)
			if err != nil {
				return fmt.Errorf("failed to fetch demand orders: %w", err)
			}

			out, err := cmd.Flags().GetString("output")
			if err == nil && out == "json" {
				return clientCtx.PrintProto(res)
			}

			for _, o := range res.DemandOrders {
				fmt.Printf(`
-id:		%s
  recipient:	%s
  price:	%s
  fee:		%s
  rollapp_id:	%s
  status:	%s
  packet_key:	%s
  packet_type:  %s
  fulfiller:	%s
`,
					o.Id, o.Recipient, parseAndFormat(o.Price), parseAndFormat(o.Fee), o.RollappId,
					o.TrackingPacketStatus, strings.TrimSpace(o.TrackingPacketKey), o.Type, o.FulfillerAddress)
			}

			fmt.Printf("\ncount: %d; ts: %s\n", len(res.DemandOrders), time.Now().Format(time.RFC3339))

			return nil
		},
	}

	cmd.Flags().StringP("rollapp", "r", "", "Filter by RollApp ID (e.g., rollapp_1234-1)")
	cmd.Flags().StringP("recipient", "c", "", "Filter by recipient address (e.g., dym1abc...)")
	cmd.Flags().StringP("type", "t", "", "Filter by packet type: recv, timeout, or ack (can omit 'ON_' prefix)")
	cmd.Flags().StringP("denom", "d", "", "Filter by token denomination (e.g., adym, ibc/...)")
	cmd.Flags().StringP("fulfilled", "f", "", "Filter by fulfillment state: 'fulfilled' or 'unfulfilled'")
	cmd.Flags().StringP("fulfiller", "a", "", "Filter by fulfiller address (e.g., dym1xyz...)")
	cmd.Flags().Int32P("limit", "l", 0, "Maximum number of orders to return (0 for no limit)")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func parseAndFormat(amount sdk.Coins) string {
	if len(amount) == 0 {
		return "0"
	}
	return fmt.Sprintf("%s %s", amount[0].Amount, amount[0].Denom)
}
