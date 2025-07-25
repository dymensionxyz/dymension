// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: dymensionxyz/dymension/kas/query.proto

package types

import (
	context "context"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-sdk/types/query"
	_ "github.com/cosmos/gogoproto/gogoproto"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	proto "github.com/cosmos/gogoproto/proto"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type QueryWithdrawalStatusRequest struct {
	WithdrawalId []*WithdrawalID `protobuf:"bytes,1,rep,name=withdrawal_id,json=withdrawalId,proto3" json:"withdrawal_id,omitempty"`
}

func (m *QueryWithdrawalStatusRequest) Reset()         { *m = QueryWithdrawalStatusRequest{} }
func (m *QueryWithdrawalStatusRequest) String() string { return proto.CompactTextString(m) }
func (*QueryWithdrawalStatusRequest) ProtoMessage()    {}
func (*QueryWithdrawalStatusRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ef2f956c4bfabfd6, []int{0}
}
func (m *QueryWithdrawalStatusRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryWithdrawalStatusRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryWithdrawalStatusRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryWithdrawalStatusRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryWithdrawalStatusRequest.Merge(m, src)
}
func (m *QueryWithdrawalStatusRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryWithdrawalStatusRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryWithdrawalStatusRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryWithdrawalStatusRequest proto.InternalMessageInfo

func (m *QueryWithdrawalStatusRequest) GetWithdrawalId() []*WithdrawalID {
	if m != nil {
		return m.WithdrawalId
	}
	return nil
}

type QueryWithdrawalStatusResponse struct {
	Status   []WithdrawalStatus  `protobuf:"varint,1,rep,packed,name=status,proto3,enum=dymensionxyz.dymension.kas.WithdrawalStatus" json:"status,omitempty"`
	Outpoint TransactionOutpoint `protobuf:"bytes,2,opt,name=outpoint,proto3" json:"outpoint"`
}

func (m *QueryWithdrawalStatusResponse) Reset()         { *m = QueryWithdrawalStatusResponse{} }
func (m *QueryWithdrawalStatusResponse) String() string { return proto.CompactTextString(m) }
func (*QueryWithdrawalStatusResponse) ProtoMessage()    {}
func (*QueryWithdrawalStatusResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ef2f956c4bfabfd6, []int{1}
}
func (m *QueryWithdrawalStatusResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryWithdrawalStatusResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryWithdrawalStatusResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryWithdrawalStatusResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryWithdrawalStatusResponse.Merge(m, src)
}
func (m *QueryWithdrawalStatusResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryWithdrawalStatusResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryWithdrawalStatusResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryWithdrawalStatusResponse proto.InternalMessageInfo

func (m *QueryWithdrawalStatusResponse) GetStatus() []WithdrawalStatus {
	if m != nil {
		return m.Status
	}
	return nil
}

func (m *QueryWithdrawalStatusResponse) GetOutpoint() TransactionOutpoint {
	if m != nil {
		return m.Outpoint
	}
	return TransactionOutpoint{}
}

type QueryOutpointRequest struct {
}

func (m *QueryOutpointRequest) Reset()         { *m = QueryOutpointRequest{} }
func (m *QueryOutpointRequest) String() string { return proto.CompactTextString(m) }
func (*QueryOutpointRequest) ProtoMessage()    {}
func (*QueryOutpointRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_ef2f956c4bfabfd6, []int{2}
}
func (m *QueryOutpointRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryOutpointRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryOutpointRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryOutpointRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryOutpointRequest.Merge(m, src)
}
func (m *QueryOutpointRequest) XXX_Size() int {
	return m.Size()
}
func (m *QueryOutpointRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryOutpointRequest.DiscardUnknown(m)
}

var xxx_messageInfo_QueryOutpointRequest proto.InternalMessageInfo

type QueryOutpointResponse struct {
	Outpoint TransactionOutpoint `protobuf:"bytes,1,opt,name=outpoint,proto3" json:"outpoint"`
}

func (m *QueryOutpointResponse) Reset()         { *m = QueryOutpointResponse{} }
func (m *QueryOutpointResponse) String() string { return proto.CompactTextString(m) }
func (*QueryOutpointResponse) ProtoMessage()    {}
func (*QueryOutpointResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_ef2f956c4bfabfd6, []int{3}
}
func (m *QueryOutpointResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *QueryOutpointResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_QueryOutpointResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *QueryOutpointResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_QueryOutpointResponse.Merge(m, src)
}
func (m *QueryOutpointResponse) XXX_Size() int {
	return m.Size()
}
func (m *QueryOutpointResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_QueryOutpointResponse.DiscardUnknown(m)
}

var xxx_messageInfo_QueryOutpointResponse proto.InternalMessageInfo

func (m *QueryOutpointResponse) GetOutpoint() TransactionOutpoint {
	if m != nil {
		return m.Outpoint
	}
	return TransactionOutpoint{}
}

func init() {
	proto.RegisterType((*QueryWithdrawalStatusRequest)(nil), "dymensionxyz.dymension.kas.QueryWithdrawalStatusRequest")
	proto.RegisterType((*QueryWithdrawalStatusResponse)(nil), "dymensionxyz.dymension.kas.QueryWithdrawalStatusResponse")
	proto.RegisterType((*QueryOutpointRequest)(nil), "dymensionxyz.dymension.kas.QueryOutpointRequest")
	proto.RegisterType((*QueryOutpointResponse)(nil), "dymensionxyz.dymension.kas.QueryOutpointResponse")
}

func init() {
	proto.RegisterFile("dymensionxyz/dymension/kas/query.proto", fileDescriptor_ef2f956c4bfabfd6)
}

var fileDescriptor_ef2f956c4bfabfd6 = []byte{
	// 448 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x93, 0x4f, 0x8b, 0xd3, 0x40,
	0x18, 0xc6, 0x33, 0x55, 0x97, 0x65, 0x56, 0x45, 0x86, 0x55, 0x96, 0xb0, 0xc6, 0x25, 0xc8, 0x52,
	0x64, 0xcd, 0x6c, 0xb3, 0x08, 0x7a, 0x5d, 0xbc, 0xec, 0x82, 0x48, 0xab, 0x20, 0x78, 0x91, 0x49,
	0x33, 0xa4, 0x63, 0x9b, 0x99, 0x34, 0x33, 0x69, 0x1b, 0x8f, 0x7e, 0x02, 0xc1, 0xb3, 0x5f, 0xa5,
	0xe7, 0x1e, 0x0b, 0x5e, 0x3c, 0x89, 0xb4, 0x1e, 0xfc, 0x18, 0xd2, 0xc9, 0x34, 0x2d, 0x85, 0xc6,
	0x3f, 0xec, 0x2d, 0x33, 0xf3, 0xbc, 0xbf, 0xe7, 0x79, 0xdf, 0x99, 0xc0, 0xe3, 0x30, 0x8f, 0x29,
	0x97, 0x4c, 0xf0, 0x51, 0xfe, 0x01, 0x97, 0x0b, 0xdc, 0x25, 0x12, 0xf7, 0x33, 0x9a, 0xe6, 0x5e,
	0x92, 0x0a, 0x25, 0x90, 0xbd, 0xae, 0xf3, 0xca, 0x85, 0xd7, 0x25, 0xd2, 0xde, 0x8f, 0x44, 0x24,
	0xb4, 0x0c, 0x2f, 0xbe, 0x8a, 0x0a, 0xfb, 0x30, 0x12, 0x22, 0xea, 0x51, 0x4c, 0x12, 0x86, 0x09,
	0xe7, 0x42, 0x11, 0xc5, 0x04, 0x97, 0xe6, 0xf4, 0x51, 0x5b, 0xc8, 0x58, 0x48, 0x1c, 0x10, 0x49,
	0x0b, 0x23, 0x3c, 0x68, 0x04, 0x54, 0x91, 0x06, 0x4e, 0x48, 0xc4, 0xb8, 0x16, 0x1b, 0xad, 0x5b,
	0x91, 0x31, 0x2c, 0x34, 0x6e, 0x0c, 0x0f, 0x9b, 0x0b, 0xca, 0x1b, 0xa6, 0x3a, 0x61, 0x4a, 0x86,
	0xa4, 0xf7, 0x4a, 0x11, 0x95, 0xc9, 0x16, 0xed, 0x67, 0x54, 0x2a, 0xf4, 0x02, 0xde, 0x1a, 0x96,
	0x47, 0xef, 0x58, 0x78, 0x00, 0x8e, 0xae, 0xd5, 0xf7, 0xfc, 0xba, 0xb7, 0xbd, 0x2f, 0x6f, 0xc5,
	0xba, 0x78, 0xde, 0xba, 0xb9, 0x2a, 0xbf, 0x08, 0xdd, 0x31, 0x80, 0xf7, 0xb7, 0xf8, 0xc9, 0x44,
	0x70, 0x49, 0xd1, 0x25, 0xdc, 0x91, 0x7a, 0x47, 0x3b, 0xdd, 0xf6, 0x4f, 0xfe, 0xce, 0xa9, 0xa0,
	0x9c, 0x5f, 0x9f, 0x7c, 0x7f, 0x60, 0xb5, 0x0c, 0x01, 0x35, 0xe1, 0xae, 0xc8, 0x54, 0x22, 0x18,
	0x57, 0x07, 0xb5, 0x23, 0x50, 0xdf, 0xf3, 0x71, 0x15, 0xed, 0x75, 0x4a, 0xb8, 0x24, 0xed, 0xc5,
	0x04, 0x5f, 0x9a, 0x32, 0x03, 0x2c, 0x31, 0xee, 0x3d, 0xb8, 0xaf, 0xf3, 0x2f, 0x05, 0x66, 0x4e,
	0xee, 0x7b, 0x78, 0x77, 0x63, 0xdf, 0xf4, 0xb3, 0x9e, 0x01, 0x5c, 0x49, 0x06, 0xff, 0x57, 0x0d,
	0xde, 0xd0, 0x66, 0x68, 0x0c, 0xe0, 0x9d, 0xcd, 0x19, 0xa0, 0xa7, 0x55, 0xfc, 0xaa, 0xcb, 0xb6,
	0x9f, 0xfd, 0x47, 0x65, 0xd1, 0xa6, 0xfb, 0xe4, 0xe3, 0xd7, 0x9f, 0x9f, 0x6b, 0x18, 0x3d, 0xc6,
	0x15, 0x8f, 0x6e, 0xed, 0x25, 0x99, 0x1b, 0xfa, 0x02, 0xe0, 0xee, 0xb2, 0x4f, 0x74, 0xfa, 0x47,
	0xfb, 0x8d, 0xa9, 0xdb, 0x8d, 0x7f, 0xa8, 0x30, 0x41, 0x4f, 0x74, 0xd0, 0x63, 0xf4, 0xb0, 0x2a,
	0xe8, 0x72, 0xd4, 0xe7, 0x97, 0x93, 0x99, 0x03, 0xa6, 0x33, 0x07, 0xfc, 0x98, 0x39, 0xe0, 0xd3,
	0xdc, 0xb1, 0xa6, 0x73, 0xc7, 0xfa, 0x36, 0x77, 0xac, 0xb7, 0xa7, 0x11, 0x53, 0x9d, 0x2c, 0xf0,
	0xda, 0x22, 0xde, 0x46, 0x1a, 0x9c, 0xe1, 0x91, 0xc6, 0xa9, 0x3c, 0xa1, 0x32, 0xd8, 0xd1, 0x7f,
	0xdc, 0xd9, 0xef, 0x00, 0x00, 0x00, 0xff, 0xff, 0x5a, 0xcf, 0x3e, 0x63, 0x3b, 0x04, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// QueryClient is the client API for Query service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type QueryClient interface {
	// check if a withdrawal was processed yet or not
	WithdrawalStatus(ctx context.Context, in *QueryWithdrawalStatusRequest, opts ...grpc.CallOption) (*QueryWithdrawalStatusResponse, error)
	// get the current outpoint which must be spent in all newly signed
	// transactions
	Outpoint(ctx context.Context, in *QueryOutpointRequest, opts ...grpc.CallOption) (*QueryOutpointResponse, error)
}

type queryClient struct {
	cc grpc1.ClientConn
}

func NewQueryClient(cc grpc1.ClientConn) QueryClient {
	return &queryClient{cc}
}

func (c *queryClient) WithdrawalStatus(ctx context.Context, in *QueryWithdrawalStatusRequest, opts ...grpc.CallOption) (*QueryWithdrawalStatusResponse, error) {
	out := new(QueryWithdrawalStatusResponse)
	err := c.cc.Invoke(ctx, "/dymensionxyz.dymension.kas.Query/WithdrawalStatus", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *queryClient) Outpoint(ctx context.Context, in *QueryOutpointRequest, opts ...grpc.CallOption) (*QueryOutpointResponse, error) {
	out := new(QueryOutpointResponse)
	err := c.cc.Invoke(ctx, "/dymensionxyz.dymension.kas.Query/Outpoint", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// QueryServer is the server API for Query service.
type QueryServer interface {
	// check if a withdrawal was processed yet or not
	WithdrawalStatus(context.Context, *QueryWithdrawalStatusRequest) (*QueryWithdrawalStatusResponse, error)
	// get the current outpoint which must be spent in all newly signed
	// transactions
	Outpoint(context.Context, *QueryOutpointRequest) (*QueryOutpointResponse, error)
}

// UnimplementedQueryServer can be embedded to have forward compatible implementations.
type UnimplementedQueryServer struct {
}

func (*UnimplementedQueryServer) WithdrawalStatus(ctx context.Context, req *QueryWithdrawalStatusRequest) (*QueryWithdrawalStatusResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method WithdrawalStatus not implemented")
}
func (*UnimplementedQueryServer) Outpoint(ctx context.Context, req *QueryOutpointRequest) (*QueryOutpointResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Outpoint not implemented")
}

func RegisterQueryServer(s grpc1.Server, srv QueryServer) {
	s.RegisterService(&_Query_serviceDesc, srv)
}

func _Query_WithdrawalStatus_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryWithdrawalStatusRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).WithdrawalStatus(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/dymensionxyz.dymension.kas.Query/WithdrawalStatus",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).WithdrawalStatus(ctx, req.(*QueryWithdrawalStatusRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Query_Outpoint_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(QueryOutpointRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(QueryServer).Outpoint(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/dymensionxyz.dymension.kas.Query/Outpoint",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(QueryServer).Outpoint(ctx, req.(*QueryOutpointRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Query_serviceDesc = grpc.ServiceDesc{
	ServiceName: "dymensionxyz.dymension.kas.Query",
	HandlerType: (*QueryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "WithdrawalStatus",
			Handler:    _Query_WithdrawalStatus_Handler,
		},
		{
			MethodName: "Outpoint",
			Handler:    _Query_Outpoint_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "dymensionxyz/dymension/kas/query.proto",
}

func (m *QueryWithdrawalStatusRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryWithdrawalStatusRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryWithdrawalStatusRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.WithdrawalId) > 0 {
		for iNdEx := len(m.WithdrawalId) - 1; iNdEx >= 0; iNdEx-- {
			{
				size, err := m.WithdrawalId[iNdEx].MarshalToSizedBuffer(dAtA[:i])
				if err != nil {
					return 0, err
				}
				i -= size
				i = encodeVarintQuery(dAtA, i, uint64(size))
			}
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func (m *QueryWithdrawalStatusResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryWithdrawalStatusResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryWithdrawalStatusResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Outpoint.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Status) > 0 {
		dAtA3 := make([]byte, len(m.Status)*10)
		var j2 int
		for _, num := range m.Status {
			for num >= 1<<7 {
				dAtA3[j2] = uint8(uint64(num)&0x7f | 0x80)
				num >>= 7
				j2++
			}
			dAtA3[j2] = uint8(num)
			j2++
		}
		i -= j2
		copy(dAtA[i:], dAtA3[:j2])
		i = encodeVarintQuery(dAtA, i, uint64(j2))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *QueryOutpointRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryOutpointRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryOutpointRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	return len(dAtA) - i, nil
}

func (m *QueryOutpointResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *QueryOutpointResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *QueryOutpointResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Outpoint.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintQuery(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintQuery(dAtA []byte, offset int, v uint64) int {
	offset -= sovQuery(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *QueryWithdrawalStatusRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.WithdrawalId) > 0 {
		for _, e := range m.WithdrawalId {
			l = e.Size()
			n += 1 + l + sovQuery(uint64(l))
		}
	}
	return n
}

func (m *QueryWithdrawalStatusResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.Status) > 0 {
		l = 0
		for _, e := range m.Status {
			l += sovQuery(uint64(e))
		}
		n += 1 + sovQuery(uint64(l)) + l
	}
	l = m.Outpoint.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func (m *QueryOutpointRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	return n
}

func (m *QueryOutpointResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.Outpoint.Size()
	n += 1 + l + sovQuery(uint64(l))
	return n
}

func sovQuery(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozQuery(x uint64) (n int) {
	return sovQuery(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *QueryWithdrawalStatusRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryWithdrawalStatusRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryWithdrawalStatusRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field WithdrawalId", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.WithdrawalId = append(m.WithdrawalId, &WithdrawalID{})
			if err := m.WithdrawalId[len(m.WithdrawalId)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryWithdrawalStatusResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryWithdrawalStatusResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryWithdrawalStatusResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType == 0 {
				var v WithdrawalStatus
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowQuery
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					v |= WithdrawalStatus(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				m.Status = append(m.Status, v)
			} else if wireType == 2 {
				var packedLen int
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowQuery
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					packedLen |= int(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				if packedLen < 0 {
					return ErrInvalidLengthQuery
				}
				postIndex := iNdEx + packedLen
				if postIndex < 0 {
					return ErrInvalidLengthQuery
				}
				if postIndex > l {
					return io.ErrUnexpectedEOF
				}
				var elementCount int
				if elementCount != 0 && len(m.Status) == 0 {
					m.Status = make([]WithdrawalStatus, 0, elementCount)
				}
				for iNdEx < postIndex {
					var v WithdrawalStatus
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowQuery
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						v |= WithdrawalStatus(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					m.Status = append(m.Status, v)
				}
			} else {
				return fmt.Errorf("proto: wrong wireType = %d for field Status", wireType)
			}
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Outpoint", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Outpoint.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryOutpointRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryOutpointRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryOutpointRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *QueryOutpointResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: QueryOutpointResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: QueryOutpointResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Outpoint", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthQuery
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthQuery
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Outpoint.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipQuery(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthQuery
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipQuery(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowQuery
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowQuery
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthQuery
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupQuery
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthQuery
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthQuery        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowQuery          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupQuery = fmt.Errorf("proto: unexpected end of group")
)
