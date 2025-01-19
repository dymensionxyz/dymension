package uevent

import (
	"testing"

	ctypes "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
)

const eventType = "test.EventMock"

func TestTypedEventToEvent(t *testing.T) {
	type args struct {
		tev proto.Message
	}
	tests := []struct {
		name    string
		args    args
		wantEv  types.Event
		wantErr error
	}{
		{
			name: "success",
			args: args{
				tev: &EventMock{
					Attribute1: "value1",
					Attribute2: 2,
					Attribute3: &TestAttribute{
						Key: "key",
					},
				},
			},
			wantEv: types.Event{
				Type: eventType,
				Attributes: []ctypes.EventAttribute{
					{
						Key:   "attribute1",
						Value: "value1",
					}, {
						Key:   "attribute2",
						Value: "2",
					}, {
						Key:   "attribute3",
						Value: `{"key":"key"}`,
					},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEv, err := TypedEventToEvent(tt.args.tev)
			require.ErrorIs(t, err, tt.wantErr)
			if err == nil {
				require.Equal(t, tt.wantEv, gotEv)
			}
		})
	}
}

type EventMock struct {
	Attribute1 string         `protobuf:"bytes,1,opt,name=attribute1,proto3" json:"attribute1,omitempty"`
	Attribute2 int32          `protobuf:"varint,2,opt,name=attribute2,proto3" json:"attribute2,omitempty"`
	Attribute3 *TestAttribute `protobuf:"bytes,3,opt,name=attribute3,proto3" json:"attribute3,omitempty"`
}

func (m *EventMock) Reset()         { *m = EventMock{} }
func (m *EventMock) String() string { return proto.CompactTextString(m) }
func (*EventMock) ProtoMessage()    {}

type TestAttribute struct {
	Key string `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
}

func (m *TestAttribute) Reset()         { *m = TestAttribute{} }
func (m *TestAttribute) String() string { return proto.CompactTextString(m) }
func (*TestAttribute) ProtoMessage()    {}

func init() {
	proto.RegisterType((*EventMock)(nil), "test.EventMock")
	proto.RegisterType((*TestAttribute)(nil), "test.TestAttribute")
}
