package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

func CmdListDemandOrdersByStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-demand-orders status [rollapp_id] [type] [fulfillment] [limit]",
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

			if len(args) > 1 {
				request.RollappId = args[1]
			}
			if len(args) > 2 {
				packetType := strings.ToUpper(args[2])
				if !strings.HasPrefix(packetType, "ON_") {
					packetType = "ON_" + packetType
				}
				ptype, ok := commontypes.RollappPacket_Type_value[packetType]
				if !ok {
					return fmt.Errorf("invalid packet type: %s", args[2])
				}
				request.Type = commontypes.RollappPacket_Type(ptype)
			}
			if len(args) > 3 {
				limit, err := strconv.ParseInt(args[3], 10, 32)
				if err != nil {
					return fmt.Errorf("limit must be an integer: %s", args[3])
				}
				request.Limit = int32(limit)
			}

			if len(args) > 4 {
				fulfillmentState, ok := types.FulfillmentState_value[strings.ToUpper(args[4])]
				if !ok {
					return fmt.Errorf("invalid fulfillment state: %s", args[4])
				}
				request.FulfillmentState = types.FulfillmentState(fulfillmentState)
			}

			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.DemandOrdersByStatus(cmd.Context(), request)
			if err != nil {
				return fmt.Errorf("failed to fetch demand orders: %w", err)
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
