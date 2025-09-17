package types

import (
	"testing"

	"cosmossdk.io/math"
	hyputil "github.com/bcp-innovations/hyperlane-cosmos/util"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

// mustHexFromString is a helper function for tests
func mustHexFromString(s string) hyputil.HexAddress {
	addr, err := hyputil.DecodeHexAddress(s)
	if err != nil {
		panic(err)
	}
	return addr
}

func TestMsgCreateBridgingFeeHook_ValidateBasic(t *testing.T) {
	validOwner := CreateRandomAccount().String()
	validFee := HLAssetFee{
		TokenID:     "0x0000000000000000000000007fa9385be102ac3eac297483dd6233d62b3e1496",
		InboundFee:  math.LegacyMustNewDecFromStr("0.01"),
		OutboundFee: math.LegacyMustNewDecFromStr("0.02"),
	}

	tests := []struct {
		name    string
		msg     MsgCreateBridgingFeeHook
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: MsgCreateBridgingFeeHook{
				Owner: validOwner,
				Fees:  []HLAssetFee{validFee},
			},
			wantErr: false,
		},
		{
			name: "invalid owner address",
			msg: MsgCreateBridgingFeeHook{
				Owner: "invalid-address",
				Fees:  []HLAssetFee{validFee},
			},
			wantErr: true,
			errMsg:  "must be a valid bech32 address",
		},
		{
			name: "empty owner address",
			msg: MsgCreateBridgingFeeHook{
				Owner: "",
				Fees:  []HLAssetFee{validFee},
			},
			wantErr: true,
			errMsg:  "must be a valid bech32 address",
		},
		{
			name: "empty fees - allowed to disable all fees",
			msg: MsgCreateBridgingFeeHook{
				Owner: validOwner,
				Fees:  []HLAssetFee{},
			},
			wantErr: false,
		},
		{
			name: "duplicate fees - allowed for same token",
			msg: MsgCreateBridgingFeeHook{
				Owner: validOwner,
				Fees: []HLAssetFee{
					validFee,
					validFee, // duplicate is allowed
				},
			},
			wantErr: false,
		},
		{
			name: "invalid fee - empty token ID",
			msg: MsgCreateBridgingFeeHook{
				Owner: validOwner,
				Fees: []HLAssetFee{{
					TokenID:     "",
					InboundFee:  math.LegacyMustNewDecFromStr("0.01"),
					OutboundFee: math.LegacyMustNewDecFromStr("0.02"),
				}},
			},
			wantErr: true,
			errMsg:  "token id cannot be empty",
		},
		{
			name: "invalid fee - negative inbound fee",
			msg: MsgCreateBridgingFeeHook{
				Owner: validOwner,
				Fees: []HLAssetFee{{
					TokenID:     "0x0000000000000000000000007fa9385be102ac3eac297483dd6233d62b3e1496",
					InboundFee:  math.LegacyMustNewDecFromStr("-0.01"),
					OutboundFee: math.LegacyMustNewDecFromStr("0.02"),
				}},
			},
			wantErr: true,
			errMsg:  "inbound fee cannot be negative",
		},
		{
			name: "invalid fee - negative outbound fee",
			msg: MsgCreateBridgingFeeHook{
				Owner: validOwner,
				Fees: []HLAssetFee{{
					TokenID:     "0x0000000000000000000000007fa9385be102ac3eac297483dd6233d62b3e1496",
					InboundFee:  math.LegacyMustNewDecFromStr("0.01"),
					OutboundFee: math.LegacyMustNewDecFromStr("-0.02"),
				}},
			},
			wantErr: true,
			errMsg:  "outbound fee cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgSetBridgingFeeHook_ValidateBasic(t *testing.T) {
	validOwner := CreateRandomAccount().String()
	validNewOwner := CreateRandomAccount().String()
	validHookId := mustHexFromString("0x0000000000000000000000007fa9385be102ac3eac297483dd6233d62b3e1496")
	validFee := HLAssetFee{
		TokenID:     "0x0000000000000000000000007fa9385be102ac3eac297483dd6233d62b3e1496",
		InboundFee:  math.LegacyMustNewDecFromStr("0.01"),
		OutboundFee: math.LegacyMustNewDecFromStr("0.02"),
	}

	tests := []struct {
		name    string
		msg     MsgSetBridgingFeeHook
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: MsgSetBridgingFeeHook{
				Id:                validHookId,
				Owner:             validOwner,
				Fees:              []HLAssetFee{validFee},
				NewOwner:          validNewOwner,
				RenounceOwnership: false,
			},
			wantErr: false,
		},
		{
			name: "valid message with ownership renouncement",
			msg: MsgSetBridgingFeeHook{
				Id:                validHookId,
				Owner:             validOwner,
				Fees:              []HLAssetFee{validFee},
				NewOwner:          "",
				RenounceOwnership: true,
			},
			wantErr: false,
		},
		{
			name: "invalid owner address",
			msg: MsgSetBridgingFeeHook{
				Id:                validHookId,
				Owner:             "invalid-address",
				Fees:              []HLAssetFee{validFee},
				NewOwner:          validNewOwner,
				RenounceOwnership: false,
			},
			wantErr: true,
			errMsg:  "must be a valid bech32 address",
		},
		{
			name: "invalid new owner address",
			msg: MsgSetBridgingFeeHook{
				Id:                validHookId,
				Owner:             validOwner,
				Fees:              []HLAssetFee{validFee},
				NewOwner:          "invalid-address",
				RenounceOwnership: false,
			},
			wantErr: true,
			errMsg:  "must be a valid bech32 address",
		},
		{
			name: "cannot both renounce ownership and set new owner",
			msg: MsgSetBridgingFeeHook{
				Id:                validHookId,
				Owner:             validOwner,
				Fees:              []HLAssetFee{validFee},
				NewOwner:          validNewOwner,
				RenounceOwnership: true,
			},
			wantErr: true,
			errMsg:  "cannot both renounce ownership and set new owner",
		},
		{
			name: "empty fees - allowed to disable all fees",
			msg: MsgSetBridgingFeeHook{
				Id:                validHookId,
				Owner:             validOwner,
				Fees:              []HLAssetFee{},
				NewOwner:          validNewOwner,
				RenounceOwnership: false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgCreateAggregationHook_ValidateBasic(t *testing.T) {
	validOwner := CreateRandomAccount().String()
	validHookId1 := mustHexFromString("0x0000000000000000000000007fa9385be102ac3eac297483dd6233d62b3e1496")
	validHookId2 := mustHexFromString("0x080ef1c2cd394de78363ecb0a466c934b57de4abb5604a0684e571990eb7b073")

	tests := []struct {
		name    string
		msg     MsgCreateAggregationHook
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: MsgCreateAggregationHook{
				Owner:   validOwner,
				HookIds: []hyputil.HexAddress{validHookId1, validHookId2},
			},
			wantErr: false,
		},
		{
			name: "valid message with single hook",
			msg: MsgCreateAggregationHook{
				Owner:   validOwner,
				HookIds: []hyputil.HexAddress{validHookId1},
			},
			wantErr: false,
		},
		{
			name: "empty hooks - allowed to disable aggregation",
			msg: MsgCreateAggregationHook{
				Owner:   validOwner,
				HookIds: []hyputil.HexAddress{},
			},
			wantErr: false,
		},
		{
			name: "duplicate hooks - allowed for same hook",
			msg: MsgCreateAggregationHook{
				Owner: validOwner,
				HookIds: []hyputil.HexAddress{
					validHookId1,
					validHookId1, // duplicate is allowed
				},
			},
			wantErr: false,
		},
		{
			name: "invalid owner address",
			msg: MsgCreateAggregationHook{
				Owner:   "invalid-address",
				HookIds: []hyputil.HexAddress{validHookId1},
			},
			wantErr: true,
			errMsg:  "must be a valid bech32 address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgSetAggregationHook_ValidateBasic(t *testing.T) {
	validOwner := CreateRandomAccount().String()
	validNewOwner := CreateRandomAccount().String()
	validHookId := mustHexFromString("0x0000000000000000000000007fa9385be102ac3eac297483dd6233d62b3e1496")
	validHookId1 := mustHexFromString("0x10df2f89cb24ed6078fc3949b4870e94a7e32e40e8d8c6b7bd74ccc2c933d760")
	validHookId2 := mustHexFromString("0x080ef1c2cd394de78363ecb0a466c934b57de4abb5604a0684e571990eb7b073")

	tests := []struct {
		name    string
		msg     MsgSetAggregationHook
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid message",
			msg: MsgSetAggregationHook{
				Id:                validHookId,
				Owner:             validOwner,
				HookIds:           []hyputil.HexAddress{validHookId1, validHookId2},
				NewOwner:          validNewOwner,
				RenounceOwnership: false,
			},
			wantErr: false,
		},
		{
			name: "valid message with ownership renouncement",
			msg: MsgSetAggregationHook{
				Id:                validHookId,
				Owner:             validOwner,
				HookIds:           []hyputil.HexAddress{validHookId1},
				NewOwner:          "",
				RenounceOwnership: true,
			},
			wantErr: false,
		},
		{
			name: "empty hooks - allowed to disable aggregation",
			msg: MsgSetAggregationHook{
				Id:                validHookId,
				Owner:             validOwner,
				HookIds:           []hyputil.HexAddress{},
				NewOwner:          validNewOwner,
				RenounceOwnership: false,
			},
			wantErr: false,
		},
		{
			name: "duplicate hooks - allowed for same hook",
			msg: MsgSetAggregationHook{
				Id:      validHookId,
				Owner:   validOwner,
				HookIds: []hyputil.HexAddress{validHookId1, validHookId1},
				NewOwner:          validNewOwner,
				RenounceOwnership: false,
			},
			wantErr: false,
		},
		{
			name: "invalid owner address",
			msg: MsgSetAggregationHook{
				Id:                validHookId,
				Owner:             "invalid-address",
				HookIds:           []hyputil.HexAddress{validHookId1},
				NewOwner:          validNewOwner,
				RenounceOwnership: false,
			},
			wantErr: true,
			errMsg:  "must be a valid bech32 address",
		},
		{
			name: "invalid new owner address",
			msg: MsgSetAggregationHook{
				Id:                validHookId,
				Owner:             validOwner,
				HookIds:           []hyputil.HexAddress{validHookId1},
				NewOwner:          "invalid-address",
				RenounceOwnership: false,
			},
			wantErr: true,
			errMsg:  "must be a valid bech32 address",
		},
		{
			name: "cannot both renounce ownership and set new owner",
			msg: MsgSetAggregationHook{
				Id:                validHookId,
				Owner:             validOwner,
				HookIds:           []hyputil.HexAddress{validHookId1},
				NewOwner:          validNewOwner,
				RenounceOwnership: true,
			},
			wantErr: true,
			errMsg:  "cannot both renounce ownership and set new owner",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// A copy from apptesting to avoid cyclic imports
func CreateRandomAccount() sdk.AccAddress {
	pk := ed25519.GenPrivKey().PubKey()
	return sdk.AccAddress(pk.Address())
}
