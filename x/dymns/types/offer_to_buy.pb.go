// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: dymensionxyz/dymension/dymns/offer_to_buy.proto

package types

import (
	fmt "fmt"
	types "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
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

// OfferToBuy defines an offer to buy a Dym-Name, placed by buyer.
// Buyer will need to deposit the offer amount to the module account.
type OfferToBuy struct {
	// The unique identifier of the offer.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// The name of the Dym-Name willing to buy.
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	// The bech32 address of the account which placed the offer.
	Buyer string `protobuf:"bytes,3,opt,name=buyer,proto3" json:"buyer,omitempty"`
	// The amount of coins willing to pay for the Dym-Name.
	// This amount is deposited to the module account.
	OfferPrice types.Coin `protobuf:"bytes,4,opt,name=offer_price,json=offerPrice,proto3" json:"offer_price"`
	// The price that the Dym-Name owner is willing to sell the Dym-Name for.
	// When this price matches offer_price, the transaction done.
	CounterpartyOfferPrice *types.Coin `protobuf:"bytes,5,opt,name=counterparty_offer_price,json=counterpartyOfferPrice,proto3" json:"counterparty_offer_price,omitempty"`
}

func (m *OfferToBuy) Reset()         { *m = OfferToBuy{} }
func (m *OfferToBuy) String() string { return proto.CompactTextString(m) }
func (*OfferToBuy) ProtoMessage()    {}
func (*OfferToBuy) Descriptor() ([]byte, []int) {
	return fileDescriptor_267456c6f8f3fedf, []int{0}
}
func (m *OfferToBuy) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *OfferToBuy) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_OfferToBuy.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *OfferToBuy) XXX_Merge(src proto.Message) {
	xxx_messageInfo_OfferToBuy.Merge(m, src)
}
func (m *OfferToBuy) XXX_Size() int {
	return m.Size()
}
func (m *OfferToBuy) XXX_DiscardUnknown() {
	xxx_messageInfo_OfferToBuy.DiscardUnknown(m)
}

var xxx_messageInfo_OfferToBuy proto.InternalMessageInfo

func (m *OfferToBuy) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *OfferToBuy) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *OfferToBuy) GetBuyer() string {
	if m != nil {
		return m.Buyer
	}
	return ""
}

func (m *OfferToBuy) GetOfferPrice() types.Coin {
	if m != nil {
		return m.OfferPrice
	}
	return types.Coin{}
}

func (m *OfferToBuy) GetCounterpartyOfferPrice() *types.Coin {
	if m != nil {
		return m.CounterpartyOfferPrice
	}
	return nil
}

// ReverseLookupOfferIds defines a message to reverse lookup offer Ids.
type ReverseLookupOfferIds struct {
	OfferIds []string `protobuf:"bytes,1,rep,name=offer_ids,json=offerIds,proto3" json:"offer_ids,omitempty"`
}

func (m *ReverseLookupOfferIds) Reset()         { *m = ReverseLookupOfferIds{} }
func (m *ReverseLookupOfferIds) String() string { return proto.CompactTextString(m) }
func (*ReverseLookupOfferIds) ProtoMessage()    {}
func (*ReverseLookupOfferIds) Descriptor() ([]byte, []int) {
	return fileDescriptor_267456c6f8f3fedf, []int{1}
}
func (m *ReverseLookupOfferIds) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *ReverseLookupOfferIds) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_ReverseLookupOfferIds.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *ReverseLookupOfferIds) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ReverseLookupOfferIds.Merge(m, src)
}
func (m *ReverseLookupOfferIds) XXX_Size() int {
	return m.Size()
}
func (m *ReverseLookupOfferIds) XXX_DiscardUnknown() {
	xxx_messageInfo_ReverseLookupOfferIds.DiscardUnknown(m)
}

var xxx_messageInfo_ReverseLookupOfferIds proto.InternalMessageInfo

func (m *ReverseLookupOfferIds) GetOfferIds() []string {
	if m != nil {
		return m.OfferIds
	}
	return nil
}

func init() {
	proto.RegisterType((*OfferToBuy)(nil), "dymensionxyz.dymension.dymns.OfferToBuy")
	proto.RegisterType((*ReverseLookupOfferIds)(nil), "dymensionxyz.dymension.dymns.ReverseLookupOfferIds")
}

func init() {
	proto.RegisterFile("dymensionxyz/dymension/dymns/offer_to_buy.proto", fileDescriptor_267456c6f8f3fedf)
}

var fileDescriptor_267456c6f8f3fedf = []byte{
	// 338 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x84, 0x91, 0xcf, 0x6a, 0x2a, 0x31,
	0x14, 0xc6, 0x27, 0xfe, 0xb9, 0x5c, 0x23, 0xdc, 0x45, 0xf0, 0x5e, 0xe6, 0xda, 0x32, 0x15, 0x57,
	0xae, 0x12, 0xd4, 0x3e, 0x40, 0xb1, 0xab, 0x82, 0x60, 0x99, 0x76, 0xd5, 0x8d, 0xcc, 0x9f, 0x68,
	0x43, 0x99, 0x9c, 0x21, 0xc9, 0x88, 0xe9, 0x53, 0xf4, 0xb1, 0x5c, 0xba, 0xec, 0xa6, 0xa5, 0xe8,
	0x8b, 0x94, 0xc9, 0x88, 0x75, 0x53, 0xba, 0x3b, 0xdf, 0xf9, 0xe6, 0xf7, 0x7d, 0x43, 0x0e, 0x66,
	0xa9, 0xcd, 0xb8, 0xd4, 0x02, 0xe4, 0xda, 0x3e, 0x7f, 0x89, 0x72, 0x92, 0x9a, 0xc1, 0x62, 0xc1,
	0xd5, 0xdc, 0xc0, 0x3c, 0x2e, 0x2c, 0xcd, 0x15, 0x18, 0x20, 0xe7, 0xa7, 0x00, 0x3d, 0x0a, 0xea,
	0x80, 0x6e, 0x67, 0x09, 0x4b, 0x70, 0x1f, 0xb2, 0x72, 0xaa, 0x98, 0x6e, 0x90, 0x80, 0xce, 0x40,
	0xb3, 0x38, 0xd2, 0x9c, 0xad, 0x86, 0x31, 0x37, 0xd1, 0x90, 0x25, 0x20, 0x64, 0xe5, 0xf7, 0xdf,
	0x10, 0xc6, 0xb3, 0xb2, 0xea, 0x1e, 0x26, 0x85, 0x25, 0x7f, 0x70, 0x4d, 0xa4, 0x3e, 0xea, 0xa1,
	0x41, 0x2b, 0xac, 0x89, 0x94, 0x10, 0xdc, 0x90, 0x51, 0xc6, 0xfd, 0x9a, 0xdb, 0xb8, 0x99, 0x74,
	0x70, 0x33, 0x2e, 0x2c, 0x57, 0x7e, 0xdd, 0x2d, 0x2b, 0x41, 0xae, 0x70, 0xbb, 0xfa, 0xe5, 0x5c,
	0x89, 0x84, 0xfb, 0x8d, 0x1e, 0x1a, 0xb4, 0x47, 0xff, 0x69, 0x55, 0x4f, 0xcb, 0x7a, 0x7a, 0xa8,
	0xa7, 0xd7, 0x20, 0xe4, 0xa4, 0xb1, 0x79, 0xbf, 0xf0, 0x42, 0xec, 0x98, 0xdb, 0x12, 0x21, 0x77,
	0xd8, 0x4f, 0xa0, 0x90, 0x86, 0xab, 0x3c, 0x52, 0xc6, 0xce, 0x4f, 0xe3, 0x9a, 0x3f, 0xc4, 0x85,
	0xff, 0x4e, 0xd1, 0xd9, 0x31, 0xb4, 0x7f, 0x89, 0xff, 0x86, 0x7c, 0xc5, 0x95, 0xe6, 0x53, 0x80,
	0xa7, 0x22, 0x77, 0xd6, 0x4d, 0xaa, 0xc9, 0x19, 0x6e, 0x55, 0x05, 0x22, 0xd5, 0x3e, 0xea, 0xd5,
	0x07, 0xad, 0xf0, 0x37, 0x1c, 0xcc, 0xc9, 0x74, 0xb3, 0x0b, 0xd0, 0x76, 0x17, 0xa0, 0x8f, 0x5d,
	0x80, 0x5e, 0xf6, 0x81, 0xb7, 0xdd, 0x07, 0xde, 0xeb, 0x3e, 0xf0, 0x1e, 0x46, 0x4b, 0x61, 0x1e,
	0x8b, 0x98, 0x26, 0x90, 0x7d, 0x77, 0xbf, 0xd5, 0x98, 0xad, 0x0f, 0x47, 0x34, 0x36, 0xe7, 0x3a,
	0xfe, 0xe5, 0x9e, 0x7a, 0xfc, 0x19, 0x00, 0x00, 0xff, 0xff, 0x21, 0x56, 0x02, 0x81, 0xf1, 0x01,
	0x00, 0x00,
}

func (m *OfferToBuy) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *OfferToBuy) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *OfferToBuy) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.CounterpartyOfferPrice != nil {
		{
			size, err := m.CounterpartyOfferPrice.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintOfferToBuy(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x2a
	}
	{
		size, err := m.OfferPrice.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintOfferToBuy(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x22
	if len(m.Buyer) > 0 {
		i -= len(m.Buyer)
		copy(dAtA[i:], m.Buyer)
		i = encodeVarintOfferToBuy(dAtA, i, uint64(len(m.Buyer)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.Name) > 0 {
		i -= len(m.Name)
		copy(dAtA[i:], m.Name)
		i = encodeVarintOfferToBuy(dAtA, i, uint64(len(m.Name)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Id) > 0 {
		i -= len(m.Id)
		copy(dAtA[i:], m.Id)
		i = encodeVarintOfferToBuy(dAtA, i, uint64(len(m.Id)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *ReverseLookupOfferIds) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *ReverseLookupOfferIds) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *ReverseLookupOfferIds) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.OfferIds) > 0 {
		for iNdEx := len(m.OfferIds) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.OfferIds[iNdEx])
			copy(dAtA[i:], m.OfferIds[iNdEx])
			i = encodeVarintOfferToBuy(dAtA, i, uint64(len(m.OfferIds[iNdEx])))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintOfferToBuy(dAtA []byte, offset int, v uint64) int {
	offset -= sovOfferToBuy(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *OfferToBuy) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Id)
	if l > 0 {
		n += 1 + l + sovOfferToBuy(uint64(l))
	}
	l = len(m.Name)
	if l > 0 {
		n += 1 + l + sovOfferToBuy(uint64(l))
	}
	l = len(m.Buyer)
	if l > 0 {
		n += 1 + l + sovOfferToBuy(uint64(l))
	}
	l = m.OfferPrice.Size()
	n += 1 + l + sovOfferToBuy(uint64(l))
	if m.CounterpartyOfferPrice != nil {
		l = m.CounterpartyOfferPrice.Size()
		n += 1 + l + sovOfferToBuy(uint64(l))
	}
	return n
}

func (m *ReverseLookupOfferIds) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.OfferIds) > 0 {
		for _, s := range m.OfferIds {
			l = len(s)
			n += 1 + l + sovOfferToBuy(uint64(l))
		}
	}
	return n
}

func sovOfferToBuy(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozOfferToBuy(x uint64) (n int) {
	return sovOfferToBuy(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *OfferToBuy) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowOfferToBuy
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
			return fmt.Errorf("proto: OfferToBuy: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: OfferToBuy: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Id", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOfferToBuy
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthOfferToBuy
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthOfferToBuy
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Id = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Name", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOfferToBuy
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthOfferToBuy
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthOfferToBuy
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Name = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Buyer", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOfferToBuy
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthOfferToBuy
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthOfferToBuy
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Buyer = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OfferPrice", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOfferToBuy
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
				return ErrInvalidLengthOfferToBuy
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthOfferToBuy
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.OfferPrice.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CounterpartyOfferPrice", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOfferToBuy
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
				return ErrInvalidLengthOfferToBuy
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthOfferToBuy
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.CounterpartyOfferPrice == nil {
				m.CounterpartyOfferPrice = &types.Coin{}
			}
			if err := m.CounterpartyOfferPrice.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipOfferToBuy(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthOfferToBuy
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
func (m *ReverseLookupOfferIds) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowOfferToBuy
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
			return fmt.Errorf("proto: ReverseLookupOfferIds: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: ReverseLookupOfferIds: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OfferIds", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowOfferToBuy
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthOfferToBuy
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthOfferToBuy
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.OfferIds = append(m.OfferIds, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipOfferToBuy(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthOfferToBuy
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
func skipOfferToBuy(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowOfferToBuy
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
					return 0, ErrIntOverflowOfferToBuy
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
					return 0, ErrIntOverflowOfferToBuy
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
				return 0, ErrInvalidLengthOfferToBuy
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupOfferToBuy
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthOfferToBuy
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthOfferToBuy        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowOfferToBuy          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupOfferToBuy = fmt.Errorf("proto: unexpected end of group")
)