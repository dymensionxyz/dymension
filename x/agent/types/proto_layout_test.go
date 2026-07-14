package types_test

import (
	"reflect"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/dymensionxyz/dymension/v3/x/common/tee"
)

func TestAgentProtoDescriptorCompatibility(t *testing.T) {
	require.NotEmpty(t, proto.FileDescriptor("dymensionxyz/dymension/agent/agent.proto"))
	require.NotEmpty(t, proto.FileDescriptor("dymensionxyz/dymension/agent/d.proto"))
	require.NotEmpty(t, proto.FileDescriptor("dymensionxyz/dymension/agent/events.proto"))
}

func TestAgentProtoFieldCompatibility(t *testing.T) {
	tests := []struct {
		message any
		fields  map[string]fieldContract
	}{
		{types.Agent{}, map[string]fieldContract{
			"Id":                  {"bytes,1,opt,name=id,proto3", reflect.TypeOf("")},
			"Policy":              {"bytes,2,opt,name=policy,proto3", reflect.TypeOf(tee.Policy{})},
			"Active":              {"varint,3,opt,name=active,proto3", reflect.TypeOf(false)},
			"ActionSeq":           {"varint,4,opt,name=action_seq,json=actionSeq,proto3", reflect.TypeOf(uint64(0))},
			"Owner":               {"bytes,5,opt,name=owner,proto3", reflect.TypeOf("")},
			"PendingPolicy":       {"bytes,6,opt,name=pending_policy,json=pendingPolicy,proto3", reflect.TypeOf((*tee.Policy)(nil))},
			"PendingPolicyHeight": {"varint,7,opt,name=pending_policy_height,json=pendingPolicyHeight,proto3", reflect.TypeOf(int64(0))},
		}},
		{types.Params{}, map[string]fieldContract{
			"MaxActionBytes":            {"varint,1,opt,name=max_action_bytes,json=maxActionBytes,proto3", reflect.TypeOf(uint64(0))},
			"AgentRegistrationFee":      {"bytes,2,opt,name=agent_registration_fee,json=agentRegistrationFee,proto3", reflect.TypeOf(types.DefaultParams().AgentRegistrationFee)},
			"PolicyRotationDelayBlocks": {"varint,3,opt,name=policy_rotation_delay_blocks,json=policyRotationDelayBlocks,proto3", reflect.TypeOf(uint64(0))},
		}},
		{types.ActionLogEntry{}, map[string]fieldContract{
			"AgentId":     {"bytes,1,opt,name=agent_id,json=agentId,proto3", reflect.TypeOf("")},
			"Seq":         {"varint,2,opt,name=seq,proto3", reflect.TypeOf(uint64(0))},
			"Payload":     {"bytes,3,opt,name=payload,proto3", reflect.TypeOf([]byte(nil))},
			"PayloadHash": {"bytes,4,opt,name=payload_hash,json=payloadHash,proto3", reflect.TypeOf([]byte(nil))},
			"Height":      {"varint,5,opt,name=height,proto3", reflect.TypeOf(int64(0))},
			"Time":        {"bytes,6,opt,name=time,proto3,stdtime", reflect.TypeOf(time.Time{})},
		}},
		{types.EventRegisterAgent{}, map[string]fieldContract{
			"AgentId": {"bytes,1,opt,name=agent_id,json=agentId,proto3", reflect.TypeOf("")},
			"Owner":   {"bytes,2,opt,name=owner,proto3", reflect.TypeOf("")},
		}},
		{types.EventDeactivateAgent{}, map[string]fieldContract{
			"AgentId": {"bytes,1,opt,name=agent_id,json=agentId,proto3", reflect.TypeOf("")},
			"Owner":   {"bytes,2,opt,name=owner,proto3", reflect.TypeOf("")},
		}},
		{types.EventUpdateAgentPolicy{}, map[string]fieldContract{
			"AgentId":          {"bytes,1,opt,name=agent_id,json=agentId,proto3", reflect.TypeOf("")},
			"ActivationHeight": {"varint,2,opt,name=activation_height,json=activationHeight,proto3", reflect.TypeOf(int64(0))},
		}},
	}

	for _, test := range tests {
		t.Run(reflect.TypeOf(test.message).Name(), func(t *testing.T) {
			typ := reflect.TypeOf(test.message)
			require.Len(t, test.fields, typ.NumField())
			for name, contract := range test.fields {
				field, found := typ.FieldByName(name)
				require.True(t, found, name)
				require.Equal(t, contract.protobufTag, field.Tag.Get("protobuf"), name)
				require.Equal(t, contract.goType, field.Type, name)
			}
		})
	}
}

func TestAgentProtoRoundTripCompatibility(t *testing.T) {
	policy := tee.Policy{
		GcpRootCertPem:  "cert",
		PolicyValues:    `{"key":"value"}`,
		PolicyQuery:     "data.agent.allow",
		PolicyStructure: "package agent",
	}
	tests := []struct {
		message proto.Message
		decoded proto.Message
	}{
		{&types.Agent{
			Id: "agent-1", Policy: policy, Active: true, ActionSeq: 7,
			Owner: "dym1owner", PendingPolicy: &policy, PendingPolicyHeight: 42,
		}, &types.Agent{}},
		{&types.Params{
			MaxActionBytes: 1024, AgentRegistrationFee: sdk.NewInt64Coin("adym", 12),
			PolicyRotationDelayBlocks: 99,
		}, &types.Params{}},
		{&types.ActionLogEntry{
			AgentId: "agent-1", Seq: 7, Payload: []byte("payload"),
			PayloadHash: []byte("hash"), Height: 42,
			Time: time.Date(2026, time.July, 14, 12, 34, 56, 789, time.UTC),
		}, &types.ActionLogEntry{}},
		{&types.EventRegisterAgent{AgentId: "agent-1", Owner: "dym1owner"}, &types.EventRegisterAgent{}},
		{&types.EventDeactivateAgent{AgentId: "agent-1", Owner: "dym1owner"}, &types.EventDeactivateAgent{}},
		{&types.EventUpdateAgentPolicy{AgentId: "agent-1", ActivationHeight: 42}, &types.EventUpdateAgentPolicy{}},
	}

	for _, test := range tests {
		t.Run(reflect.TypeOf(test.message).Elem().Name(), func(t *testing.T) {
			bz, err := proto.Marshal(test.message)
			require.NoError(t, err)

			require.NoError(t, proto.Unmarshal(bz, test.decoded))
			require.Equal(t, test.message, test.decoded)
		})
	}
}

type fieldContract struct {
	protobufTag string
	goType      reflect.Type
}
