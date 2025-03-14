// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: dymensionxyz/dymension/sponsorship/events.proto

package types

import (
	cosmossdk_io_math "cosmossdk.io/math"
	fmt "fmt"
	_ "github.com/cosmos/cosmos-proto"
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

type EventUpdateParams struct {
	Authority string `protobuf:"bytes,1,opt,name=authority,proto3" json:"authority,omitempty"`
	NewParams Params `protobuf:"bytes,2,opt,name=new_params,json=newParams,proto3" json:"new_params"`
	OldParams Params `protobuf:"bytes,3,opt,name=old_params,json=oldParams,proto3" json:"old_params"`
}

func (m *EventUpdateParams) Reset()         { *m = EventUpdateParams{} }
func (m *EventUpdateParams) String() string { return proto.CompactTextString(m) }
func (*EventUpdateParams) ProtoMessage()    {}
func (*EventUpdateParams) Descriptor() ([]byte, []int) {
	return fileDescriptor_b80e9ef6d6e7fb59, []int{0}
}
func (m *EventUpdateParams) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventUpdateParams) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventUpdateParams.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventUpdateParams) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventUpdateParams.Merge(m, src)
}
func (m *EventUpdateParams) XXX_Size() int {
	return m.Size()
}
func (m *EventUpdateParams) XXX_DiscardUnknown() {
	xxx_messageInfo_EventUpdateParams.DiscardUnknown(m)
}

var xxx_messageInfo_EventUpdateParams proto.InternalMessageInfo

func (m *EventUpdateParams) GetAuthority() string {
	if m != nil {
		return m.Authority
	}
	return ""
}

func (m *EventUpdateParams) GetNewParams() Params {
	if m != nil {
		return m.NewParams
	}
	return Params{}
}

func (m *EventUpdateParams) GetOldParams() Params {
	if m != nil {
		return m.OldParams
	}
	return Params{}
}

type EventVote struct {
	Voter        string       `protobuf:"bytes,1,opt,name=voter,proto3" json:"voter,omitempty"`
	Vote         Vote         `protobuf:"bytes,2,opt,name=vote,proto3" json:"vote"`
	Distribution Distribution `protobuf:"bytes,3,opt,name=distribution,proto3" json:"distribution"`
}

func (m *EventVote) Reset()         { *m = EventVote{} }
func (m *EventVote) String() string { return proto.CompactTextString(m) }
func (*EventVote) ProtoMessage()    {}
func (*EventVote) Descriptor() ([]byte, []int) {
	return fileDescriptor_b80e9ef6d6e7fb59, []int{1}
}
func (m *EventVote) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventVote) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventVote.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventVote) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventVote.Merge(m, src)
}
func (m *EventVote) XXX_Size() int {
	return m.Size()
}
func (m *EventVote) XXX_DiscardUnknown() {
	xxx_messageInfo_EventVote.DiscardUnknown(m)
}

var xxx_messageInfo_EventVote proto.InternalMessageInfo

func (m *EventVote) GetVoter() string {
	if m != nil {
		return m.Voter
	}
	return ""
}

func (m *EventVote) GetVote() Vote {
	if m != nil {
		return m.Vote
	}
	return Vote{}
}

func (m *EventVote) GetDistribution() Distribution {
	if m != nil {
		return m.Distribution
	}
	return Distribution{}
}

type EventRevokeVote struct {
	Voter        string       `protobuf:"bytes,1,opt,name=voter,proto3" json:"voter,omitempty"`
	Distribution Distribution `protobuf:"bytes,2,opt,name=distribution,proto3" json:"distribution"`
}

func (m *EventRevokeVote) Reset()         { *m = EventRevokeVote{} }
func (m *EventRevokeVote) String() string { return proto.CompactTextString(m) }
func (*EventRevokeVote) ProtoMessage()    {}
func (*EventRevokeVote) Descriptor() ([]byte, []int) {
	return fileDescriptor_b80e9ef6d6e7fb59, []int{2}
}
func (m *EventRevokeVote) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventRevokeVote) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventRevokeVote.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventRevokeVote) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventRevokeVote.Merge(m, src)
}
func (m *EventRevokeVote) XXX_Size() int {
	return m.Size()
}
func (m *EventRevokeVote) XXX_DiscardUnknown() {
	xxx_messageInfo_EventRevokeVote.DiscardUnknown(m)
}

var xxx_messageInfo_EventRevokeVote proto.InternalMessageInfo

func (m *EventRevokeVote) GetVoter() string {
	if m != nil {
		return m.Voter
	}
	return ""
}

func (m *EventRevokeVote) GetDistribution() Distribution {
	if m != nil {
		return m.Distribution
	}
	return Distribution{}
}

type EventVotingPowerUpdate struct {
	Voter           string                `protobuf:"bytes,1,opt,name=voter,proto3" json:"voter,omitempty"`
	Validator       string                `protobuf:"bytes,2,opt,name=validator,proto3" json:"validator,omitempty"`
	Distribution    Distribution          `protobuf:"bytes,3,opt,name=distribution,proto3" json:"distribution"`
	VotePruned      bool                  `protobuf:"varint,4,opt,name=vote_pruned,json=votePruned,proto3" json:"vote_pruned,omitempty"`
	NewVotingPower  cosmossdk_io_math.Int `protobuf:"bytes,5,opt,name=new_voting_power,json=newVotingPower,proto3,customtype=cosmossdk.io/math.Int" json:"new_voting_power"`
	VotingPowerDiff cosmossdk_io_math.Int `protobuf:"bytes,6,opt,name=voting_power_diff,json=votingPowerDiff,proto3,customtype=cosmossdk.io/math.Int" json:"voting_power_diff"`
}

func (m *EventVotingPowerUpdate) Reset()         { *m = EventVotingPowerUpdate{} }
func (m *EventVotingPowerUpdate) String() string { return proto.CompactTextString(m) }
func (*EventVotingPowerUpdate) ProtoMessage()    {}
func (*EventVotingPowerUpdate) Descriptor() ([]byte, []int) {
	return fileDescriptor_b80e9ef6d6e7fb59, []int{3}
}
func (m *EventVotingPowerUpdate) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *EventVotingPowerUpdate) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_EventVotingPowerUpdate.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *EventVotingPowerUpdate) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EventVotingPowerUpdate.Merge(m, src)
}
func (m *EventVotingPowerUpdate) XXX_Size() int {
	return m.Size()
}
func (m *EventVotingPowerUpdate) XXX_DiscardUnknown() {
	xxx_messageInfo_EventVotingPowerUpdate.DiscardUnknown(m)
}

var xxx_messageInfo_EventVotingPowerUpdate proto.InternalMessageInfo

func (m *EventVotingPowerUpdate) GetVoter() string {
	if m != nil {
		return m.Voter
	}
	return ""
}

func (m *EventVotingPowerUpdate) GetValidator() string {
	if m != nil {
		return m.Validator
	}
	return ""
}

func (m *EventVotingPowerUpdate) GetDistribution() Distribution {
	if m != nil {
		return m.Distribution
	}
	return Distribution{}
}

func (m *EventVotingPowerUpdate) GetVotePruned() bool {
	if m != nil {
		return m.VotePruned
	}
	return false
}

func init() {
	proto.RegisterType((*EventUpdateParams)(nil), "dymensionxyz.dymension.sponsorship.EventUpdateParams")
	proto.RegisterType((*EventVote)(nil), "dymensionxyz.dymension.sponsorship.EventVote")
	proto.RegisterType((*EventRevokeVote)(nil), "dymensionxyz.dymension.sponsorship.EventRevokeVote")
	proto.RegisterType((*EventVotingPowerUpdate)(nil), "dymensionxyz.dymension.sponsorship.EventVotingPowerUpdate")
}

func init() {
	proto.RegisterFile("dymensionxyz/dymension/sponsorship/events.proto", fileDescriptor_b80e9ef6d6e7fb59)
}

var fileDescriptor_b80e9ef6d6e7fb59 = []byte{
	// 516 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x94, 0xc1, 0x4e, 0x13, 0x41,
	0x18, 0xc7, 0x3b, 0x50, 0x88, 0xfd, 0x30, 0x22, 0x1b, 0x34, 0x2b, 0x87, 0x6d, 0xd3, 0x53, 0xa3,
	0x61, 0xd7, 0x88, 0x21, 0x5e, 0x6d, 0xf0, 0xc0, 0x89, 0x66, 0x0d, 0x9a, 0x70, 0xd9, 0x6c, 0x99,
	0xe9, 0x76, 0x42, 0x77, 0xbe, 0xcd, 0xcc, 0x74, 0x4b, 0x7d, 0x0a, 0x5f, 0xc0, 0x47, 0xf0, 0xc6,
	0x43, 0x70, 0x24, 0x9c, 0x8c, 0x26, 0xc4, 0xb4, 0x67, 0xdf, 0xc1, 0xec, 0xec, 0x42, 0x17, 0x13,
	0xa5, 0x10, 0xb9, 0xcd, 0xcc, 0x7e, 0xff, 0xff, 0xfc, 0xfe, 0x5f, 0x66, 0x3f, 0xf0, 0xe8, 0x38,
	0x66, 0x42, 0x71, 0x14, 0xc7, 0xe3, 0x4f, 0xb3, 0x8d, 0xa7, 0x12, 0x14, 0x0a, 0xa5, 0xea, 0xf3,
	0xc4, 0x63, 0x29, 0x13, 0x5a, 0xb9, 0x89, 0x44, 0x8d, 0x56, 0xb3, 0x2c, 0x70, 0xaf, 0x36, 0x6e,
	0x49, 0xb0, 0xb1, 0x1e, 0x61, 0x84, 0xa6, 0xdc, 0xcb, 0x56, 0xb9, 0x72, 0xe3, 0xd9, 0x21, 0xaa,
	0x18, 0x55, 0x90, 0x7f, 0xc8, 0x37, 0xc5, 0xa7, 0xd7, 0x73, 0x50, 0x94, 0xd6, 0xb9, 0xaa, 0xf9,
	0x8b, 0xc0, 0xda, 0xbb, 0x8c, 0x6d, 0x3f, 0xa1, 0xa1, 0x66, 0x9d, 0x50, 0x86, 0xb1, 0xb2, 0xb6,
	0xa1, 0x16, 0x0e, 0x75, 0x1f, 0x25, 0xd7, 0x63, 0x9b, 0x34, 0x48, 0xab, 0xd6, 0xb6, 0xcf, 0x4f,
	0x36, 0xd7, 0x8b, 0x0b, 0xdf, 0x52, 0x2a, 0x99, 0x52, 0xef, 0xb5, 0xe4, 0x22, 0xf2, 0x67, 0xa5,
	0xd6, 0x1e, 0x80, 0x60, 0xa3, 0x20, 0x31, 0x2e, 0xf6, 0x42, 0x83, 0xb4, 0x56, 0x5e, 0x3d, 0x77,
	0x6f, 0x4e, 0xeb, 0xe6, 0xf7, 0xb6, 0xab, 0xa7, 0x17, 0xf5, 0x8a, 0x5f, 0x13, 0x6c, 0x54, 0x80,
	0xec, 0x01, 0xe0, 0x80, 0x5e, 0x1a, 0x2e, 0xde, 0xd5, 0x10, 0x07, 0x34, 0x3f, 0x68, 0xfe, 0x20,
	0x50, 0x33, 0x79, 0x3f, 0xa0, 0x66, 0x96, 0x0b, 0x4b, 0x29, 0x6a, 0x26, 0x6f, 0xcc, 0x98, 0x97,
	0x59, 0x6d, 0xa8, 0x66, 0x8b, 0x22, 0x59, 0x6b, 0x1e, 0x90, 0xec, 0x9e, 0x02, 0xc3, 0x68, 0xad,
	0x03, 0x78, 0x48, 0xb9, 0xd2, 0x92, 0x77, 0x87, 0x9a, 0xa3, 0x28, 0x42, 0xbd, 0x9c, 0xc7, 0x6b,
	0xa7, 0xa4, 0x2b, 0x3c, 0xaf, 0x79, 0x35, 0xbf, 0x10, 0x58, 0x35, 0xe9, 0x7c, 0x96, 0xe2, 0x11,
	0xbb, 0x53, 0xc6, 0x3f, 0xf9, 0x16, 0xfe, 0x23, 0xdf, 0xd7, 0x45, 0x78, 0x7a, 0xd9, 0x7d, 0x2e,
	0xa2, 0x0e, 0x8e, 0x98, 0xcc, 0x1f, 0xde, 0xad, 0x31, 0xb7, 0xa1, 0x96, 0x86, 0x03, 0x4e, 0x43,
	0x8d, 0xd2, 0x30, 0xfe, 0xf3, 0x89, 0x5e, 0x95, 0xde, 0x67, 0xfb, 0xad, 0x3a, 0xac, 0x64, 0x70,
	0x41, 0x22, 0x87, 0x82, 0x51, 0xbb, 0xda, 0x20, 0xad, 0x07, 0x3e, 0x64, 0x47, 0x1d, 0x73, 0x62,
	0xed, 0xc3, 0xe3, 0xec, 0xff, 0x48, 0x4d, 0xfa, 0x20, 0xc9, 0xe2, 0xdb, 0x4b, 0x86, 0xfd, 0x45,
	0x66, 0xf7, 0xfd, 0xa2, 0xfe, 0x24, 0xe7, 0x57, 0xf4, 0xc8, 0xe5, 0xe8, 0xc5, 0xa1, 0xee, 0xbb,
	0xbb, 0x42, 0x9f, 0x9f, 0x6c, 0x42, 0x11, 0x6c, 0x57, 0x68, 0xff, 0x91, 0x60, 0xa3, 0x52, 0x07,
	0xad, 0x8f, 0xb0, 0x56, 0xb6, 0x0c, 0x28, 0xef, 0xf5, 0xec, 0xe5, 0xdb, 0xfb, 0xae, 0xa6, 0x33,
	0xd3, 0x1d, 0xde, 0xeb, 0xb5, 0xfd, 0xd3, 0x89, 0x43, 0xce, 0x26, 0x0e, 0xf9, 0x39, 0x71, 0xc8,
	0xe7, 0xa9, 0x53, 0x39, 0x9b, 0x3a, 0x95, 0x6f, 0x53, 0xa7, 0x72, 0xf0, 0x26, 0xe2, 0xba, 0x3f,
	0xec, 0xba, 0x87, 0x18, 0xff, 0x6d, 0xfc, 0xa5, 0x5b, 0xde, 0xf1, 0xb5, 0xe9, 0xa3, 0xc7, 0x09,
	0x53, 0xdd, 0x65, 0x33, 0x78, 0xb6, 0x7e, 0x07, 0x00, 0x00, 0xff, 0xff, 0x04, 0x08, 0x76, 0xa5,
	0x36, 0x05, 0x00, 0x00,
}

func (m *EventUpdateParams) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventUpdateParams) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventUpdateParams) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.OldParams.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintEvents(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	{
		size, err := m.NewParams.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintEvents(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Authority) > 0 {
		i -= len(m.Authority)
		copy(dAtA[i:], m.Authority)
		i = encodeVarintEvents(dAtA, i, uint64(len(m.Authority)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *EventVote) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventVote) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventVote) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Distribution.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintEvents(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	{
		size, err := m.Vote.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintEvents(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Voter) > 0 {
		i -= len(m.Voter)
		copy(dAtA[i:], m.Voter)
		i = encodeVarintEvents(dAtA, i, uint64(len(m.Voter)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *EventRevokeVote) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventRevokeVote) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventRevokeVote) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size, err := m.Distribution.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintEvents(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.Voter) > 0 {
		i -= len(m.Voter)
		copy(dAtA[i:], m.Voter)
		i = encodeVarintEvents(dAtA, i, uint64(len(m.Voter)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *EventVotingPowerUpdate) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *EventVotingPowerUpdate) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *EventVotingPowerUpdate) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	{
		size := m.VotingPowerDiff.Size()
		i -= size
		if _, err := m.VotingPowerDiff.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintEvents(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x32
	{
		size := m.NewVotingPower.Size()
		i -= size
		if _, err := m.NewVotingPower.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintEvents(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x2a
	if m.VotePruned {
		i--
		if m.VotePruned {
			dAtA[i] = 1
		} else {
			dAtA[i] = 0
		}
		i--
		dAtA[i] = 0x20
	}
	{
		size, err := m.Distribution.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintEvents(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if len(m.Validator) > 0 {
		i -= len(m.Validator)
		copy(dAtA[i:], m.Validator)
		i = encodeVarintEvents(dAtA, i, uint64(len(m.Validator)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.Voter) > 0 {
		i -= len(m.Voter)
		copy(dAtA[i:], m.Voter)
		i = encodeVarintEvents(dAtA, i, uint64(len(m.Voter)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintEvents(dAtA []byte, offset int, v uint64) int {
	offset -= sovEvents(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *EventUpdateParams) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Authority)
	if l > 0 {
		n += 1 + l + sovEvents(uint64(l))
	}
	l = m.NewParams.Size()
	n += 1 + l + sovEvents(uint64(l))
	l = m.OldParams.Size()
	n += 1 + l + sovEvents(uint64(l))
	return n
}

func (m *EventVote) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Voter)
	if l > 0 {
		n += 1 + l + sovEvents(uint64(l))
	}
	l = m.Vote.Size()
	n += 1 + l + sovEvents(uint64(l))
	l = m.Distribution.Size()
	n += 1 + l + sovEvents(uint64(l))
	return n
}

func (m *EventRevokeVote) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Voter)
	if l > 0 {
		n += 1 + l + sovEvents(uint64(l))
	}
	l = m.Distribution.Size()
	n += 1 + l + sovEvents(uint64(l))
	return n
}

func (m *EventVotingPowerUpdate) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.Voter)
	if l > 0 {
		n += 1 + l + sovEvents(uint64(l))
	}
	l = len(m.Validator)
	if l > 0 {
		n += 1 + l + sovEvents(uint64(l))
	}
	l = m.Distribution.Size()
	n += 1 + l + sovEvents(uint64(l))
	if m.VotePruned {
		n += 2
	}
	l = m.NewVotingPower.Size()
	n += 1 + l + sovEvents(uint64(l))
	l = m.VotingPowerDiff.Size()
	n += 1 + l + sovEvents(uint64(l))
	return n
}

func sovEvents(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozEvents(x uint64) (n int) {
	return sovEvents(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *EventUpdateParams) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvents
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
			return fmt.Errorf("proto: EventUpdateParams: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventUpdateParams: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Authority", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
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
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Authority = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NewParams", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
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
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.NewParams.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field OldParams", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
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
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.OldParams.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipEvents(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthEvents
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
func (m *EventVote) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvents
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
			return fmt.Errorf("proto: EventVote: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventVote: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Voter", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
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
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Voter = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Vote", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
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
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Vote.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Distribution", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
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
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Distribution.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipEvents(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthEvents
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
func (m *EventRevokeVote) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvents
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
			return fmt.Errorf("proto: EventRevokeVote: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventRevokeVote: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Voter", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
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
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Voter = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Distribution", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
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
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Distribution.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipEvents(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthEvents
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
func (m *EventVotingPowerUpdate) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowEvents
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
			return fmt.Errorf("proto: EventVotingPowerUpdate: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: EventVotingPowerUpdate: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Voter", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
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
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Voter = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Validator", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
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
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Validator = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Distribution", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
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
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.Distribution.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field VotePruned", wireType)
			}
			var v int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.VotePruned = bool(v != 0)
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NewVotingPower", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
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
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.NewVotingPower.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field VotingPowerDiff", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowEvents
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
				return ErrInvalidLengthEvents
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthEvents
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.VotingPowerDiff.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipEvents(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthEvents
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
func skipEvents(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowEvents
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
					return 0, ErrIntOverflowEvents
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
					return 0, ErrIntOverflowEvents
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
				return 0, ErrInvalidLengthEvents
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupEvents
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthEvents
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthEvents        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowEvents          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupEvents = fmt.Errorf("proto: unexpected end of group")
)
