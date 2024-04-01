package keeper

import (
	"testing"

	"github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/stretchr/testify/require"
)

func TestByRollappID(t *testing.T) {
	type args struct {
		rollappID string
	}
	tests := []struct {
		name string
		args args
		want rollappPacketListFilter
	}{
		{
			name: "Test with rollappID 1",
			args: args{
				rollappID: "testRollappID1",
			},
			want: rollappPacketListFilter{
				prefixes: []prefix{
					{
						start: []uint8{0x00, 0x01, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f},
					}, {
						start: []uint8{0x00, 0x02, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f},
					}, {
						start: []uint8{0x00, 0x03, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f},
					},
				},
			},
		}, {
			name: "Test with empty rollappID",
			args: args{
				rollappID: "",
			},
			want: rollappPacketListFilter{
				prefixes: []prefix{
					{
						start: []uint8{0x00, 0x01, 0x2f, 0x2f},
					}, {
						start: []uint8{0x00, 0x02, 0x2f, 0x2f},
					}, {
						start: []uint8{0x00, 0x03, 0x2f, 0x2f},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := ByRollappID(tt.args.rollappID)
			require.Equal(t, tt.want, filter)
		})
	}
}

func TestByRollappIDByStatus(t *testing.T) {
	type args struct {
		rollappID string
		status    []types.Status
	}
	tests := []struct {
		name string
		args args
		want rollappPacketListFilter
	}{
		{
			name: "Test with rollappID 1 and status PENDING",
			args: args{
				rollappID: "testRollappID1",
				status:    []types.Status{types.Status_PENDING},
			},
			want: rollappPacketListFilter{
				prefixes: []prefix{
					{
						start: []uint8{0x00, 0x01, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f},
					},
				},
			},
		}, {
			name: "Test with rollappID 1 and status FINALIZED",
			args: args{
				rollappID: "testRollappID1",
				status:    []types.Status{types.Status_FINALIZED},
			},
			want: rollappPacketListFilter{
				prefixes: []prefix{
					{
						start: []uint8{0x00, 0x02, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f},
					},
				},
			},
		}, {
			name: "Test with rollappID 1 and status REVERTED",
			args: args{
				rollappID: "testRollappID1",
				status:    []types.Status{types.Status_REVERTED},
			},
			want: rollappPacketListFilter{
				prefixes: []prefix{
					{
						start: []uint8{0x00, 0x03, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f},
					},
				},
			},
		}, {
			name: "Test with rollappID 1 and status PENDING, FINALIZED",
			args: args{
				rollappID: "testRollappID1",
				status:    []types.Status{types.Status_PENDING, types.Status_FINALIZED},
			},
			want: rollappPacketListFilter{
				prefixes: []prefix{
					{
						start: []uint8{0x00, 0x01, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f},
					}, {
						start: []uint8{0x00, 0x02, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f},
					},
				},
			},
		}, {
			name: "Test with empty rollappID and status PENDING",
			args: args{
				rollappID: "",
				status:    []types.Status{types.Status_PENDING},
			},
			want: rollappPacketListFilter{
				prefixes: []prefix{
					{
						start: []uint8{0x00, 0x01, 0x2f, 0x2f},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := ByRollappIDByStatus(tt.args.rollappID, tt.args.status...)
			require.Equal(t, tt.want, filter)
		})
	}
}

func TestByStatus(t *testing.T) {
	type args struct {
		status []types.Status
	}
	tests := []struct {
		name string
		args args
		want rollappPacketListFilter
	}{
		{
			name: "Test with status PENDING",
			args: args{
				status: []types.Status{types.Status_PENDING},
			},
			want: rollappPacketListFilter{
				prefixes: []prefix{
					{
						start: []uint8{0x00, 0x01, 0x2f},
					},
				},
			},
		}, {
			name: "Test with status FINALIZED",
			args: args{
				status: []types.Status{types.Status_FINALIZED},
			},
			want: rollappPacketListFilter{
				prefixes: []prefix{
					{
						start: []uint8{0x00, 0x02, 0x2f},
					},
				},
			},
		}, {
			name: "Test with status REVERTED",
			args: args{
				status: []types.Status{types.Status_REVERTED},
			},
			want: rollappPacketListFilter{
				prefixes: []prefix{
					{
						start: []uint8{0x00, 0x03, 0x2f},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := ByStatus(tt.args.status...)
			require.Equal(t, tt.want, filter)
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
		want rollappPacketListFilter
	}{
		{
			name: "Test with rollappID 1 and maxProofHeight 100",
			args: args{
				rollappID:      "testRollappID1",
				maxProofHeight: 100,
			},
			want: rollappPacketListFilter{
				prefixes: []prefix{
					{
						start: []uint8{0x00, 0x01, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
						end:   []uint8{0x00, 0x01, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x65},
					},
				},
			},
		}, {
			name: "Test with empty rollappID and maxProofHeight 100",
			args: args{
				rollappID:      "",
				maxProofHeight: 100,
			},
			want: rollappPacketListFilter{
				prefixes: []prefix{
					{
						start: []uint8{0x0, 0x1, 0x2f, 0x2f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
						end:   []uint8{0x0, 0x1, 0x2f, 0x2f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x65},
					},
				},
			},
		}, {
			name: "Test with rollappID 1 and maxProofHeight 0",
			args: args{
				rollappID:      "testRollappID1",
				maxProofHeight: 0,
			},
			want: rollappPacketListFilter{
				prefixes: []prefix{
					{
						start: []uint8{0x0, 0x1, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0},
						end:   []uint8{0x0, 0x1, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x52, 0x6f, 0x6c, 0x6c, 0x61, 0x70, 0x70, 0x49, 0x44, 0x31, 0x2f, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := PendingByRollappIDByMaxHeight(tt.args.rollappID, tt.args.maxProofHeight)
			require.Equal(t, tt.want, filter)
		})
	}
}
