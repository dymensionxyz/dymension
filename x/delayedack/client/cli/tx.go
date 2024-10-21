package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdFinalizePacket())

	return cmd
}

func CmdFinalizePacket() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "finalize-packet [rollapp-id] [proof-height] [packet-type] [packet-src-channel] [packet-sequence] --from <sender>",
		Short: "Finalize a specified packet",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			proofHeight, err := strconv.Atoi(args[1])
			if err != nil {
				return err
			}

			packetType, err := parsePacketType(args[2])
			if err != nil {
				return err
			}

			packetSequence, err := strconv.Atoi(args[4])
			if err != nil {
				return err
			}

			msg := types.MsgFinalizePacket{
				Sender:            clientCtx.GetFromAddress().String(),
				RollappId:         args[0],
				PacketProofHeight: uint64(proofHeight),
				PacketType:        packetType,
				PacketSrcChannel:  args[3],
				PacketSequence:    uint64(packetSequence),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), &msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func parsePacketType(packetType string) (commontypes.RollappPacket_Type, error) {
	switch packetType {
	case commontypes.RollappPacket_ON_RECV.String():
		return commontypes.RollappPacket_ON_RECV, nil
	case commontypes.RollappPacket_ON_ACK.String():
		return commontypes.RollappPacket_ON_ACK, nil
	case commontypes.RollappPacket_ON_TIMEOUT.String():
		return commontypes.RollappPacket_ON_TIMEOUT, nil
	default:
		return 0, fmt.Errorf("invalid packet type: %s; must be one of: %s, %s, %s",
			packetType,
			commontypes.RollappPacket_ON_RECV.String(),
			commontypes.RollappPacket_ON_ACK.String(),
			commontypes.RollappPacket_ON_TIMEOUT.String(),
		)
	}
}
