// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: dymensionxyz/dymension/iro/params.proto

package types

import (
	cosmossdk_io_math "cosmossdk.io/math"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
	_ "github.com/cosmos/cosmos-sdk/types"
	_ "github.com/cosmos/gogoproto/gogoproto"
	proto "github.com/cosmos/gogoproto/proto"
	github_com_cosmos_gogoproto_types "github.com/cosmos/gogoproto/types"
	_ "google.golang.org/protobuf/types/known/durationpb"
	_ "google.golang.org/protobuf/types/known/timestamppb"
	io "io"
	math "math"
	math_bits "math/bits"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// Params is a module parameters.
type Params struct {
	TakerFee cosmossdk_io_math.LegacyDec `protobuf:"bytes,1,opt,name=taker_fee,json=takerFee,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"taker_fee"`
	// The fee for creating a plan in rollapp tokens (e.g 1000000000000000000 for
	// 1Token fee) must be > 1 The cost in `liquidity_denom` is charged from the
	// creator
	CreationFee cosmossdk_io_math.Int `protobuf:"bytes,2,opt,name=creation_fee,json=creationFee,proto3,customtype=cosmossdk.io/math.Int" json:"creation_fee"`
	// Minimum plan duration
	// when creating a plan, the plan duration should be greater than or equal to
	// this value plan duration is from the start time to the pre-launch time
	MinPlanDuration time.Duration `protobuf:"bytes,3,opt,name=min_plan_duration,json=minPlanDuration,proto3,stdduration" json:"min_plan_duration"`
	// The minimum time after settlement when the incentives distribution will
	// start
	IncentivesMinStartTimeAfterSettlement time.Duration `protobuf:"bytes,4,opt,name=incentives_min_start_time_after_settlement,json=incentivesMinStartTimeAfterSettlement,proto3,stdduration" json:"incentives_min_start_time_after_settlement"`
	// The minimum number of epochs over which the incentives will be paid
	IncentivesMinNumEpochsPaidOver uint64 `protobuf:"varint,5,opt,name=incentives_min_num_epochs_paid_over,json=incentivesMinNumEpochsPaidOver,proto3" json:"incentives_min_num_epochs_paid_over,omitempty"`
	// The minimum part of the raised liquidity that must be used for pool
	// bootstrapping the other part goes to the founder
	MinLiquidityPart   cosmossdk_io_math.LegacyDec `protobuf:"bytes,6,opt,name=min_liquidity_part,json=minLiquidityPart,proto3,customtype=cosmossdk.io/math.LegacyDec" json:"min_liquidity_part"`
	MinVestingDuration time.Duration               `protobuf:"bytes,7,opt,name=min_vesting_duration,json=minVestingDuration,proto3,stdduration" json:"min_vesting_duration"`
	// Minimum start time after settlement to start vesting
	MinVestingStartTimeAfterSettlement time.Duration `protobuf:"bytes,8,opt,name=min_vesting_start_time_after_settlement,json=minVestingStartTimeAfterSettlement,proto3,stdduration" json:"min_vesting_start_time_after_settlement"`
}

func (m *Params) Reset()         { *m = Params{} }
func (m *Params) String() string { return proto.CompactTextString(m) }
func (*Params) ProtoMessage()    {}
func (*Params) Descriptor() ([]byte, []int) {
	return fileDescriptor_321dd4e17bb4cbec, []int{0}
}
func (m *Params) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Params) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Params.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Params) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Params.Merge(m, src)
}
func (m *Params) XXX_Size() int {
	return m.Size()
}
func (m *Params) XXX_DiscardUnknown() {
	xxx_messageInfo_Params.DiscardUnknown(m)
}

var xxx_messageInfo_Params proto.InternalMessageInfo

func (m *Params) GetMinPlanDuration() time.Duration {
	if m != nil {
		return m.MinPlanDuration
	}
	return 0
}

func (m *Params) GetIncentivesMinStartTimeAfterSettlement() time.Duration {
	if m != nil {
		return m.IncentivesMinStartTimeAfterSettlement
	}
	return 0
}

func (m *Params) GetIncentivesMinNumEpochsPaidOver() uint64 {
	if m != nil {
		return m.IncentivesMinNumEpochsPaidOver
	}
	return 0
}

func (m *Params) GetMinVestingDuration() time.Duration {
	if m != nil {
		return m.MinVestingDuration
	}
	return 0
}

func (m *Params) GetMinVestingStartTimeAfterSettlement() time.Duration {
	if m != nil {
		return m.MinVestingStartTimeAfterSettlement
	}
	return 0
}

func init() {
	proto.RegisterType((*Params)(nil), "dymensionxyz.dymension.iro.Params")
}

func init() {
	proto.RegisterFile("dymensionxyz/dymension/iro/params.proto", fileDescriptor_321dd4e17bb4cbec)
}

var fileDescriptor_321dd4e17bb4cbec = []byte{
	// 538 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x93, 0xc1, 0x6e, 0xd3, 0x4e,
	0x10, 0xc6, 0xe3, 0xff, 0xbf, 0x2d, 0xad, 0x8b, 0x04, 0x58, 0x45, 0x4a, 0x83, 0xe4, 0x44, 0x45,
	0xa8, 0x11, 0x08, 0x2f, 0xa1, 0x4f, 0x40, 0x14, 0x90, 0x0a, 0x25, 0x8d, 0x52, 0xe0, 0xc0, 0x65,
	0xb5, 0xb1, 0x27, 0xce, 0xaa, 0xd9, 0x5d, 0xb3, 0x3b, 0xb6, 0x12, 0x0e, 0x3c, 0x03, 0x47, 0x1e,
	0x84, 0x87, 0xe8, 0xb1, 0xe2, 0x84, 0x38, 0x14, 0x94, 0xbc, 0x04, 0x47, 0x64, 0x3b, 0x4e, 0x42,
	0x51, 0x51, 0xc4, 0x6d, 0x47, 0xf3, 0xcd, 0xef, 0x1b, 0x7d, 0xa3, 0xb5, 0xf7, 0x83, 0xb1, 0x00,
	0x69, 0xb8, 0x92, 0xa3, 0xf1, 0x7b, 0x32, 0x2f, 0x08, 0xd7, 0x8a, 0x44, 0x4c, 0x33, 0x61, 0xbc,
	0x48, 0x2b, 0x54, 0x4e, 0x65, 0x59, 0xe8, 0xcd, 0x0b, 0x8f, 0x6b, 0x55, 0xd9, 0x09, 0x55, 0xa8,
	0x32, 0x19, 0x49, 0x5f, 0xf9, 0x44, 0xa5, 0x1a, 0x2a, 0x15, 0x0e, 0x81, 0x64, 0x55, 0x2f, 0xee,
	0x13, 0xe4, 0x02, 0x0c, 0x32, 0x11, 0xcd, 0x04, 0xee, 0x65, 0x41, 0x10, 0x6b, 0x86, 0x29, 0x74,
	0xd6, 0xf7, 0x95, 0x11, 0xca, 0x90, 0x1e, 0x33, 0x40, 0x92, 0x46, 0x0f, 0x90, 0x35, 0x88, 0xaf,
	0x78, 0xd1, 0xdf, 0xcd, 0xfb, 0x34, 0x77, 0xce, 0x8b, 0xbc, 0xb5, 0xf7, 0x73, 0xdd, 0xde, 0xe8,
	0x64, 0xeb, 0x3b, 0x6d, 0x7b, 0x0b, 0xd9, 0x29, 0x68, 0xda, 0x07, 0x28, 0x5b, 0x35, 0xab, 0xbe,
	0xd5, 0x6c, 0x9c, 0x5d, 0x54, 0x4b, 0xdf, 0x2e, 0xaa, 0x77, 0xf2, 0x19, 0x13, 0x9c, 0x7a, 0x5c,
	0x11, 0xc1, 0x70, 0xe0, 0x1d, 0x41, 0xc8, 0xfc, 0x71, 0x0b, 0xfc, 0x2f, 0x9f, 0x1f, 0xda, 0x33,
	0x64, 0x0b, 0xfc, 0xee, 0x66, 0xc6, 0x78, 0x06, 0xe0, 0xb4, 0xed, 0xeb, 0xbe, 0x86, 0x6c, 0xcf,
	0x0c, 0xf9, 0x5f, 0x86, 0x7c, 0x30, 0x43, 0xde, 0xfe, 0x13, 0x79, 0x28, 0x71, 0x09, 0x76, 0x28,
	0xb1, 0xbb, 0x5d, 0x00, 0x52, 0xde, 0xb1, 0x7d, 0x4b, 0x70, 0x49, 0xa3, 0x21, 0x93, 0xb4, 0x08,
	0xa0, 0xfc, 0x7f, 0xcd, 0xaa, 0x6f, 0x3f, 0xde, 0xf5, 0xf2, 0x84, 0xbc, 0x22, 0x21, 0xaf, 0x35,
	0x13, 0x34, 0x37, 0x53, 0xbf, 0x4f, 0xdf, 0xab, 0x56, 0xf7, 0x86, 0xe0, 0xb2, 0x33, 0x64, 0xb2,
	0x68, 0x39, 0x1f, 0xec, 0xfb, 0x5c, 0xfa, 0x20, 0x91, 0x27, 0x60, 0x68, 0xca, 0x36, 0xc8, 0x34,
	0xd2, 0x34, 0x7e, 0xca, 0xfa, 0x08, 0x9a, 0x1a, 0x40, 0x1c, 0x82, 0x00, 0x89, 0xe5, 0xb5, 0xd5,
	0x9d, 0xee, 0x2d, 0xb0, 0x2f, 0xb9, 0x3c, 0x49, 0xa1, 0xaf, 0xb8, 0x80, 0x27, 0x29, 0xf2, 0x64,
	0x4e, 0x74, 0x5e, 0xd8, 0x77, 0x2f, 0xf9, 0xcb, 0x58, 0x50, 0x88, 0x94, 0x3f, 0x30, 0x34, 0x62,
	0x3c, 0xa0, 0x2a, 0x01, 0x5d, 0x5e, 0xaf, 0x59, 0xf5, 0xb5, 0xae, 0xfb, 0x1b, 0xb3, 0x1d, 0x8b,
	0xa7, 0x99, 0xae, 0xc3, 0x78, 0x70, 0x9c, 0x80, 0x76, 0xa8, 0xed, 0xa4, 0x84, 0x21, 0x7f, 0x17,
	0xf3, 0x80, 0xe3, 0x98, 0x46, 0x4c, 0x63, 0x79, 0xe3, 0x5f, 0xcf, 0x78, 0x53, 0x70, 0x79, 0x54,
	0xb0, 0x3a, 0x4c, 0xa3, 0xf3, 0xda, 0xde, 0x49, 0x0d, 0x12, 0x30, 0xc8, 0x65, 0xb8, 0xb8, 0xc0,
	0xb5, 0xd5, 0x73, 0x49, 0x37, 0x7c, 0x93, 0xcf, 0xcf, 0x8f, 0x30, 0xb2, 0xf7, 0x97, 0xb1, 0x7f,
	0xbb, 0xc0, 0xe6, 0xea, 0x4e, 0x7b, 0x0b, 0xa7, 0xab, 0xe2, 0x6f, 0x3e, 0x3f, 0x9b, 0xb8, 0xd6,
	0xf9, 0xc4, 0xb5, 0x7e, 0x4c, 0x5c, 0xeb, 0xe3, 0xd4, 0x2d, 0x9d, 0x4f, 0xdd, 0xd2, 0xd7, 0xa9,
	0x5b, 0x7a, 0xfb, 0x28, 0xe4, 0x38, 0x88, 0x7b, 0x9e, 0xaf, 0x04, 0xb9, 0xe2, 0xdb, 0x27, 0x07,
	0x64, 0x94, 0xfd, 0x7d, 0x1c, 0x47, 0x60, 0x7a, 0x1b, 0xd9, 0x32, 0x07, 0xbf, 0x02, 0x00, 0x00,
	0xff, 0xff, 0x8d, 0x37, 0x2d, 0x84, 0x26, 0x04, 0x00, 0x00,
}

func (m *Params) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Params) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *Params) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	n1, err1 := github_com_cosmos_gogoproto_types.StdDurationMarshalTo(m.MinVestingStartTimeAfterSettlement, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.MinVestingStartTimeAfterSettlement):])
	if err1 != nil {
		return 0, err1
	}
	i -= n1
	i = encodeVarintParams(dAtA, i, uint64(n1))
	i--
	dAtA[i] = 0x42
	n2, err2 := github_com_cosmos_gogoproto_types.StdDurationMarshalTo(m.MinVestingDuration, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.MinVestingDuration):])
	if err2 != nil {
		return 0, err2
	}
	i -= n2
	i = encodeVarintParams(dAtA, i, uint64(n2))
	i--
	dAtA[i] = 0x3a
	{
		size := m.MinLiquidityPart.Size()
		i -= size
		if _, err := m.MinLiquidityPart.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x32
	if m.IncentivesMinNumEpochsPaidOver != 0 {
		i = encodeVarintParams(dAtA, i, uint64(m.IncentivesMinNumEpochsPaidOver))
		i--
		dAtA[i] = 0x28
	}
	n3, err3 := github_com_cosmos_gogoproto_types.StdDurationMarshalTo(m.IncentivesMinStartTimeAfterSettlement, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.IncentivesMinStartTimeAfterSettlement):])
	if err3 != nil {
		return 0, err3
	}
	i -= n3
	i = encodeVarintParams(dAtA, i, uint64(n3))
	i--
	dAtA[i] = 0x22
	n4, err4 := github_com_cosmos_gogoproto_types.StdDurationMarshalTo(m.MinPlanDuration, dAtA[i-github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.MinPlanDuration):])
	if err4 != nil {
		return 0, err4
	}
	i -= n4
	i = encodeVarintParams(dAtA, i, uint64(n4))
	i--
	dAtA[i] = 0x1a
	{
		size := m.CreationFee.Size()
		i -= size
		if _, err := m.CreationFee.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	{
		size := m.TakerFee.Size()
		i -= size
		if _, err := m.TakerFee.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintParams(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0xa
	return len(dAtA) - i, nil
}

func encodeVarintParams(dAtA []byte, offset int, v uint64) int {
	offset -= sovParams(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *Params) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = m.TakerFee.Size()
	n += 1 + l + sovParams(uint64(l))
	l = m.CreationFee.Size()
	n += 1 + l + sovParams(uint64(l))
	l = github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.MinPlanDuration)
	n += 1 + l + sovParams(uint64(l))
	l = github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.IncentivesMinStartTimeAfterSettlement)
	n += 1 + l + sovParams(uint64(l))
	if m.IncentivesMinNumEpochsPaidOver != 0 {
		n += 1 + sovParams(uint64(m.IncentivesMinNumEpochsPaidOver))
	}
	l = m.MinLiquidityPart.Size()
	n += 1 + l + sovParams(uint64(l))
	l = github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.MinVestingDuration)
	n += 1 + l + sovParams(uint64(l))
	l = github_com_cosmos_gogoproto_types.SizeOfStdDuration(m.MinVestingStartTimeAfterSettlement)
	n += 1 + l + sovParams(uint64(l))
	return n
}

func sovParams(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozParams(x uint64) (n int) {
	return sovParams(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Params) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowParams
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
			return fmt.Errorf("proto: Params: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: Params: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field TakerFee", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.TakerFee.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CreationFee", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.CreationFee.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinPlanDuration", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdDurationUnmarshal(&m.MinPlanDuration, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field IncentivesMinStartTimeAfterSettlement", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdDurationUnmarshal(&m.IncentivesMinStartTimeAfterSettlement, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 5:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field IncentivesMinNumEpochsPaidOver", wireType)
			}
			m.IncentivesMinNumEpochsPaidOver = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.IncentivesMinNumEpochsPaidOver |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinLiquidityPart", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.MinLiquidityPart.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinVestingDuration", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdDurationUnmarshal(&m.MinVestingDuration, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field MinVestingStartTimeAfterSettlement", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowParams
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
				return ErrInvalidLengthParams
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthParams
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_cosmos_gogoproto_types.StdDurationUnmarshal(&m.MinVestingStartTimeAfterSettlement, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipParams(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthParams
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
func skipParams(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowParams
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
					return 0, ErrIntOverflowParams
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
					return 0, ErrIntOverflowParams
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
				return 0, ErrInvalidLengthParams
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupParams
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthParams
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthParams        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowParams          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupParams = fmt.Errorf("proto: unexpected end of group")
)
