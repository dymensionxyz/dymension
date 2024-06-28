package types_test

import (
	"testing"

	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/stretchr/testify/require"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"
)

func TestByRollappID(t *testing.T) {
	type args struct {
		rollappID string
	}
	tests := []struct {
		name string
		args args
		want []types.Prefix
	}{
		{
			name: "Test with rollappID 1",
			args: args{
				rollappID: "testRollappID1",
			},
			want: []types.Prefix{
				{
					Start: []uint8{0x00, 0x01, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f},
				}, {
					Start: []uint8{0x00, 0x02, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f},
				}, {
					Start: []uint8{0x00, 0x03, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f},
				},
			},
		}, {
			name: "Test with empty rollappID",
			args: args{
				rollappID: "",
			},
			want: []types.Prefix{
				{
					Start: []uint8{0x00, 0x01, 0x2f, 0x2f},
				}, {
					Start: []uint8{0x00, 0x02, 0x2f, 0x2f},
				}, {
					Start: []uint8{0x00, 0x03, 0x2f, 0x2f},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := types.ByRollappID(tt.args.rollappID)
			require.Equal(t, tt.want, filter.Prefixes)
		})
	}
}

func TestByRollappIDByStatus(t *testing.T) {
	type args struct {
		rollappID string
		status    []commontypes.Status
	}
	tests := []struct {
		name string
		args args
		want []types.Prefix
	}{
		{
			name: "Test with rollappID 1 and status PENDING",
			args: args{
				rollappID: "testRollappID1",
				status:    []commontypes.Status{commontypes.Status_PENDING},
			},
			want: []types.Prefix{
				{
					Start: []uint8{0x00, 0x01, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f},
				},
			},
		}, {
			name: "Test with rollappID 1 and status FINALIZED",
			args: args{
				rollappID: "testRollappID1",
				status:    []commontypes.Status{commontypes.Status_FINALIZED},
			},
			want: []types.Prefix{
				{
					Start: []uint8{0x00, 0x02, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f},
				},
			},
		}, {
			name: "Test with rollappID 1 and status REVERTED",
			args: args{
				rollappID: "testRollappID1",
				status:    []commontypes.Status{commontypes.Status_REVERTED},
			},
			want: []types.Prefix{
				{
					Start: []uint8{0x00, 0x03, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f},
				},
			},
		}, {
			name: "Test with rollappID 1 and status PENDING, FINALIZED",
			args: args{
				rollappID: "testRollappID1",
				status:    []commontypes.Status{commontypes.Status_PENDING, commontypes.Status_FINALIZED},
			},
			want: []types.Prefix{
				{
					Start: []uint8{0x00, 0x01, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f},
				}, {
					Start: []uint8{0x00, 0x02, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f},
				},
			},
		}, {
			name: "Test with empty rollappID and status PENDING",
			args: args{
				rollappID: "",
				status:    []commontypes.Status{commontypes.Status_PENDING},
			},
			want: []types.Prefix{
				{
					Start: []uint8{0x00, 0x01, 0x2f, 0x2f},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := types.ByRollappIDByStatus(tt.args.rollappID, tt.args.status...)
			require.Equal(t, tt.want, filter.Prefixes)
		})
	}
}

func TestByStatus(t *testing.T) {
	type args struct {
		status []commontypes.Status
	}
	tests := []struct {
		name string
		args args
		want []types.Prefix
	}{
		{
			name: "Test with status PENDING",
			args: args{
				status: []commontypes.Status{commontypes.Status_PENDING},
			},
			want: []types.Prefix{
				{
					Start: []uint8{0x00, 0x01, 0x2f},
				},
			},
		}, {
			name: "Test with status FINALIZED",
			args: args{
				status: []commontypes.Status{commontypes.Status_FINALIZED},
			},
			want: []types.Prefix{
				{
					Start: []uint8{0x00, 0x02, 0x2f},
				},
			},
		}, {
			name: "Test with status REVERTED",
			args: args{
				status: []commontypes.Status{commontypes.Status_REVERTED},
			},
			want: []types.Prefix{
				{
					Start: []uint8{0x00, 0x03, 0x2f},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := types.ByStatus(tt.args.status...)
			require.Equal(t, tt.want, filter.Prefixes)
		})
	}
}

func TestPendingByRollappIDByMaxHeight(t *testing.T) {
	type args struct {
		rollappID      string
		maxProofHeight uint64
	}
	tests := []struct {
		name string
		args args
		want []types.Prefix
	}{
		{
			name: "Test with rollappID 1 and maxProofHeight 100",
			args: args{
				rollappID:      "testRollappID1",
				maxProofHeight: 100,
			},
			want: []types.Prefix{
				{
					Start: []uint8{0x00, 0x01, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
					End:   []uint8{0x00, 0x01, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x65},
				},
			},
		}, {
			name: "Test with empty rollappID and maxProofHeight 100",
			args: args{
				rollappID:      "",
				maxProofHeight: 100,
			},
			want: []types.Prefix{
				{
					Start: []uint8{0x0, 0x1, 0x2f, 0x2f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
					End:   []uint8{0x0, 0x1, 0x2f, 0x2f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x65},
				},
			},
		}, {
			name: "Test with rollappID 1 and maxProofHeight 0",
			args: args{
				rollappID:      "testRollappID1",
				maxProofHeight: 0,
			},
			want: []types.Prefix{
				{
					Start: []uint8{0x0, 0x1, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
					End:   []uint8{0x0, 0x1, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := types.PendingByRollappIDByMaxHeight(tt.args.rollappID, tt.args.maxProofHeight)
			require.Equal(t, tt.want, filter.Prefixes)
		})
	}
}

func TestByType(t *testing.T) {
	type args struct {
		packetType commontypes.RollappPacket_Type
	}
	tests := []struct {
		name string
		args args
		want []commontypes.RollappPacket
	}{
		{
			name: "Test with Type ON_RECV",
			args: args{
				packetType: commontypes.RollappPacket_ON_RECV,
			},
			want: testRollappPackets[1:2],
		}, {
			name: "Test with Type ON_ACK",
			args: args{
				packetType: commontypes.RollappPacket_ON_ACK,
			},
			want: testRollappPackets[2:3],
		}, {
			name: "Test with Type ON_TIMEOUT",
			args: args{
				packetType: commontypes.RollappPacket_ON_TIMEOUT,
			},
			want: testRollappPackets[3:4],
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := types.ByTypeByStatus(tt.args.packetType)
			var filtered []commontypes.RollappPacket
			for _, packet := range testRollappPackets {
				if filter.FilterFunc(packet) {
					filtered = append(filtered, packet)
				}
			}
			require.Equal(t, tt.want, filtered)
		})
	}
}

var testRollappPackets = []commontypes.RollappPacket{
	{
		RollappId: "rollapp-id-1",
		Packet:    &channeltypes.Packet{},
		Type:      commontypes.RollappPacket_UNDEFINED,
	}, {
		RollappId: "rollapp-id-2",
		Packet:    &channeltypes.Packet{},
		Type:      commontypes.RollappPacket_ON_RECV,
	}, {
		RollappId: "rollapp-id-3",
		Packet:    &channeltypes.Packet{},
		Type:      commontypes.RollappPacket_ON_ACK,
	}, {
		RollappId: "rollapp-id-4",
		Packet:    &channeltypes.Packet{},
		Type:      commontypes.RollappPacket_ON_TIMEOUT,
	},
}
