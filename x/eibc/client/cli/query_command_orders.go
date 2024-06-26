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

func CmdListDemandOrdersByStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-demand-orders status [rollapp] [recipient] [type] [denom] [fulfilled] [fulfiller] [limit]",
		Short: "List all demand orders with a specific status",
		Long: `Query demand orders filtered by status. Examples of status include "pending", "finalized", and "reverted".
				Optional arguments include rollapp_id, type (recv, timeout, ack), and limit.`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			status, ok := commontypes.Status_value[strings.ToUpper(args[0])]
			if !ok {
				return fmt.Errorf("invalid status: %s", args[0])
			}
			request := &types.QueryDemandOrdersByStatusRequest{
				Status: commontypes.Status(status),
				Type:   commontypes.RollappPacket_UNDEFINED, // default to undefined, as '0' is a valid type
			}

			var err error
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

			clientCtx := client.GetClientContextFromCmd(cmd)
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
  fulfiller:	%s
`,
					o.Id, o.Recipient, parseAndFormat(o.Price), parseAndFormat(o.Fee), o.RollappId,
					o.TrackingPacketStatus, strings.TrimSpace(o.TrackingPacketKey), o.FulfillerAddress)
			}

			fmt.Printf("\ncount: %d; ts: %s\n", len(res.DemandOrders), time.Now().Format(time.RFC3339))

			return nil
		},
	}

	cmd.Flags().StringP("rollapp", "r", "", "Rollapp ID")
	cmd.Flags().StringP("recipient", "c", "", "Recipient address")
	cmd.Flags().StringP("type", "t", "", "Packet type")
	cmd.Flags().StringP("denom", "d", "", "Denom")
	cmd.Flags().StringP("fulfilled", "f", "", "Filter by fulfillment status")
	cmd.Flags().StringP("fulfiller", "a", "", "Filter by fulfiller address")
	cmd.Flags().Int32P("limit", "l", 0, "Limit orders to display")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func parseAndFormat(amount sdk.Coins) string {
	if len(amount) == 0 {
		return "0"
	}
	return fmt.Sprintf("%s %s", amount[0].Amount, amount[0].Denom)
}
