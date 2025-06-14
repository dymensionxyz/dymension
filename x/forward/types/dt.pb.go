// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: dymensionxyz/dymension/forward/dt.proto

package types

import (
	fmt "fmt"
	types "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	proto "github.com/cosmos/gogoproto/proto"
	types1 "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
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

type HookForwardToHL struct {
	HyperlaneTransfer *types.MsgRemoteTransfer `protobuf:"bytes,1,opt,name=hyperlane_transfer,json=hyperlaneTransfer,proto3" json:"hyperlane_transfer,omitempty"`
}

func (m *HookForwardToHL) Reset()         { *m = HookForwardToHL{} }
func (m *HookForwardToHL) String() string { return proto.CompactTextString(m) }
func (*HookForwardToHL) ProtoMessage()    {}
func (*HookForwardToHL) Descriptor() ([]byte, []int) {
	return fileDescriptor_abdb3fdf27098576, []int{0}
}
func (m *HookForwardToHL) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *HookForwardToHL) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_HookForwardToHL.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *HookForwardToHL) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HookForwardToHL.Merge(m, src)
}
func (m *HookForwardToHL) XXX_Size() int {
	return m.Size()
}
func (m *HookForwardToHL) XXX_DiscardUnknown() {
	xxx_messageInfo_HookForwardToHL.DiscardUnknown(m)
}

var xxx_messageInfo_HookForwardToHL proto.InternalMessageInfo

func (m *HookForwardToHL) GetHyperlaneTransfer() *types.MsgRemoteTransfer {
	if m != nil {
		return m.HyperlaneTransfer
	}
	return nil
}

type HookForwardToIBC struct {
	Transfer *types1.MsgTransfer `protobuf:"bytes,1,opt,name=transfer,proto3" json:"transfer,omitempty"`
}

func (m *HookForwardToIBC) Reset()         { *m = HookForwardToIBC{} }
func (m *HookForwardToIBC) String() string { return proto.CompactTextString(m) }
func (*HookForwardToIBC) ProtoMessage()    {}
func (*HookForwardToIBC) Descriptor() ([]byte, []int) {
	return fileDescriptor_abdb3fdf27098576, []int{1}
}
func (m *HookForwardToIBC) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *HookForwardToIBC) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_HookForwardToIBC.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *HookForwardToIBC) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HookForwardToIBC.Merge(m, src)
}
func (m *HookForwardToIBC) XXX_Size() int {
	return m.Size()
}
func (m *HookForwardToIBC) XXX_DiscardUnknown() {
	xxx_messageInfo_HookForwardToIBC.DiscardUnknown(m)
}

var xxx_messageInfo_HookForwardToIBC proto.InternalMessageInfo

func (m *HookForwardToIBC) GetTransfer() *types1.MsgTransfer {
	if m != nil {
		return m.Transfer
	}
	return nil
}

// Expected format of metadata received in HL warp route messages
// There is only one metadata, so we need to share it amongst our applications,
// so that they can compose and not conflict
type HLMetadata struct {
	// optional, can be empty
	HookForwardToIbc []byte `protobuf:"bytes,1,opt,name=hook_forward_to_ibc,json=hookForwardToIbc,proto3" json:"hook_forward_to_ibc,omitempty"`
	// optional, can be empty
	// see
	// https://www.notion.so/dymension/ADR-Kaspa-Bridge-Implementation-206a4a51f86a803980aec7099c826fb4?source=copy_link#208a4a51f86a8093a843cf4b5e903588
	Kaspa []byte `protobuf:"bytes,2,opt,name=kaspa,proto3" json:"kaspa,omitempty"`
}

func (m *HLMetadata) Reset()         { *m = HLMetadata{} }
func (m *HLMetadata) String() string { return proto.CompactTextString(m) }
func (*HLMetadata) ProtoMessage()    {}
func (*HLMetadata) Descriptor() ([]byte, []int) {
	return fileDescriptor_abdb3fdf27098576, []int{2}
}
func (m *HLMetadata) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *HLMetadata) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_HLMetadata.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *HLMetadata) XXX_Merge(src proto.Message) {
	xxx_messageInfo_HLMetadata.Merge(m, src)
}
func (m *HLMetadata) XXX_Size() int {
	return m.Size()
}
func (m *HLMetadata) XXX_DiscardUnknown() {
	xxx_messageInfo_HLMetadata.DiscardUnknown(m)
}

var xxx_messageInfo_HLMetadata proto.InternalMessageInfo

func (m *HLMetadata) GetHookForwardToIbc() []byte {
	if m != nil {
		return m.HookForwardToIbc
	}
	return nil
}

func (m *HLMetadata) GetKaspa() []byte {
	if m != nil {
		return m.Kaspa
	}
	return nil
}

func init() {
	proto.RegisterType((*HookForwardToHL)(nil), "dymensionxyz.dymension.forward.HookForwardToHL")
	proto.RegisterType((*HookForwardToIBC)(nil), "dymensionxyz.dymension.forward.HookForwardToIBC")
	proto.RegisterType((*HLMetadata)(nil), "dymensionxyz.dymension.forward.HLMetadata")
}

func init() {
	proto.RegisterFile("dymensionxyz/dymension/forward/dt.proto", fileDescriptor_abdb3fdf27098576)
}

var fileDescriptor_abdb3fdf27098576 = []byte{
	// 324 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x74, 0x91, 0xc1, 0x4e, 0x02, 0x31,
	0x14, 0x45, 0x19, 0x13, 0x8d, 0xa9, 0x26, 0xe2, 0xe8, 0x82, 0xb0, 0x68, 0x0c, 0xd1, 0xa8, 0x0b,
	0xdb, 0x20, 0x7e, 0x01, 0x46, 0x83, 0x09, 0x98, 0x88, 0x6c, 0x74, 0x33, 0x69, 0x3b, 0x85, 0x69,
	0x80, 0x79, 0x4d, 0x5b, 0x81, 0xf1, 0x2b, 0xfc, 0x2c, 0x97, 0x2c, 0x5d, 0x1a, 0xf8, 0x11, 0x23,
	0x0c, 0x0d, 0x98, 0xb8, 0xbc, 0xed, 0xb9, 0xf7, 0xbe, 0xbc, 0x87, 0xce, 0xe3, 0x6c, 0x28, 0x53,
	0xab, 0x20, 0x9d, 0x64, 0xef, 0xd4, 0x0b, 0xda, 0x05, 0x33, 0x66, 0x26, 0xa6, 0xb1, 0x23, 0xda,
	0x80, 0x83, 0x10, 0xaf, 0x83, 0xc4, 0x0b, 0x92, 0x83, 0xe5, 0x72, 0x92, 0x69, 0x69, 0x06, 0x2c,
	0x95, 0x74, 0xcc, 0x8c, 0xa6, 0xa3, 0x2a, 0x75, 0x93, 0xa5, 0xb7, 0x7c, 0xa6, 0xb8, 0xa0, 0x4c,
	0xeb, 0x81, 0x12, 0xcc, 0x29, 0x48, 0x2d, 0x75, 0x86, 0xa5, 0xb6, 0x2b, 0xcd, 0x3a, 0x56, 0xe9,
	0xa2, 0x83, 0x06, 0x40, 0xff, 0x7e, 0x99, 0xd8, 0x81, 0x46, 0x33, 0x7c, 0x46, 0xa1, 0xcf, 0x8d,
	0x56, 0xa6, 0x52, 0x70, 0x12, 0x5c, 0xec, 0x5d, 0x9f, 0x12, 0xff, 0x45, 0x7e, 0x2b, 0xc9, 0xa8,
	0x4a, 0x5a, 0xb6, 0xd7, 0x96, 0x43, 0x70, 0xb2, 0x93, 0xb3, 0xed, 0x43, 0x0f, 0xad, 0x9e, 0x2a,
	0x2f, 0xa8, 0xb8, 0xd1, 0xf3, 0x50, 0xbf, 0x0d, 0xef, 0xd0, 0xee, 0x9f, 0xf8, 0x4b, 0xa2, 0xb8,
	0x20, 0xeb, 0x53, 0x93, 0x15, 0x91, 0x37, 0xf9, 0x0e, 0x6f, 0xad, 0x3c, 0x21, 0xd4, 0x68, 0xb6,
	0xa4, 0x63, 0x31, 0x73, 0x2c, 0xbc, 0x42, 0x47, 0x09, 0x40, 0x3f, 0xca, 0x77, 0x14, 0x39, 0x88,
	0x14, 0x17, 0x8b, 0xfc, 0xfd, 0x76, 0x31, 0xd9, 0x98, 0x81, 0x8b, 0xf0, 0x18, 0x6d, 0xf7, 0x99,
	0xd5, 0xac, 0xb4, 0xb5, 0x00, 0x96, 0xa2, 0xfe, 0xf8, 0x39, 0xc3, 0xc1, 0x74, 0x86, 0x83, 0xef,
	0x19, 0x0e, 0x3e, 0xe6, 0xb8, 0x30, 0x9d, 0xe3, 0xc2, 0xd7, 0x1c, 0x17, 0x5e, 0x6f, 0x7a, 0xca,
	0x25, 0x6f, 0x9c, 0x08, 0x18, 0xd2, 0x7f, 0xce, 0x38, 0xaa, 0xd1, 0x89, 0xbf, 0xa5, 0xcb, 0xb4,
	0xb4, 0x7c, 0x67, 0xb1, 0xec, 0xda, 0x4f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x44, 0x5c, 0x14, 0x87,
	0xfa, 0x01, 0x00, 0x00,
}

func (m *HookForwardToHL) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *HookForwardToHL) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *HookForwardToHL) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.HyperlaneTransfer != nil {
		{
			size, err := m.HyperlaneTransfer.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintDt(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *HookForwardToIBC) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *HookForwardToIBC) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *HookForwardToIBC) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.Transfer != nil {
		{
			size, err := m.Transfer.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintDt(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *HLMetadata) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *HLMetadata) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *HLMetadata) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Kaspa) > 0 {
		i -= len(m.Kaspa)
		copy(dAtA[i:], m.Kaspa)
		i = encodeVarintDt(dAtA, i, uint64(len(m.Kaspa)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.HookForwardToIbc) > 0 {
		i -= len(m.HookForwardToIbc)
		copy(dAtA[i:], m.HookForwardToIbc)
		i = encodeVarintDt(dAtA, i, uint64(len(m.HookForwardToIbc)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintDt(dAtA []byte, offset int, v uint64) int {
	offset -= sovDt(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *HookForwardToHL) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.HyperlaneTransfer != nil {
		l = m.HyperlaneTransfer.Size()
		n += 1 + l + sovDt(uint64(l))
	}
	return n
}

func (m *HookForwardToIBC) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Transfer != nil {
		l = m.Transfer.Size()
		n += 1 + l + sovDt(uint64(l))
	}
	return n
}

func (m *HLMetadata) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.HookForwardToIbc)
	if l > 0 {
		n += 1 + l + sovDt(uint64(l))
	}
	l = len(m.Kaspa)
	if l > 0 {
		n += 1 + l + sovDt(uint64(l))
	}
	return n
}

func sovDt(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozDt(x uint64) (n int) {
	return sovDt(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *HookForwardToHL) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDt
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
			return fmt.Errorf("proto: HookForwardToHL: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: HookForwardToHL: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field HyperlaneTransfer", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDt
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
				return ErrInvalidLengthDt
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthDt
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.HyperlaneTransfer == nil {
				m.HyperlaneTransfer = &types.MsgRemoteTransfer{}
			}
			if err := m.HyperlaneTransfer.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipDt(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthDt
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
func (m *HookForwardToIBC) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDt
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
			return fmt.Errorf("proto: HookForwardToIBC: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: HookForwardToIBC: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Transfer", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDt
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
				return ErrInvalidLengthDt
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthDt
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Transfer == nil {
				m.Transfer = &types1.MsgTransfer{}
			}
			if err := m.Transfer.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipDt(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthDt
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
func (m *HLMetadata) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowDt
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
			return fmt.Errorf("proto: HLMetadata: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: HLMetadata: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field HookForwardToIbc", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDt
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthDt
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthDt
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.HookForwardToIbc = append(m.HookForwardToIbc[:0], dAtA[iNdEx:postIndex]...)
			if m.HookForwardToIbc == nil {
				m.HookForwardToIbc = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Kaspa", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowDt
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthDt
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthDt
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Kaspa = append(m.Kaspa[:0], dAtA[iNdEx:postIndex]...)
			if m.Kaspa == nil {
				m.Kaspa = []byte{}
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipDt(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthDt
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
func skipDt(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowDt
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
					return 0, ErrIntOverflowDt
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
					return 0, ErrIntOverflowDt
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
				return 0, ErrInvalidLengthDt
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupDt
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthDt
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthDt        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowDt          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupDt = fmt.Errorf("proto: unexpected end of group")
)
