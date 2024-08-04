package types

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
func TestDymName_Validate(t *testing.T) {
	t.Run("nil obj", func(t *testing.T) {
		m := (*DymName)(nil)
		require.Error(t, m.Validate())
	})

	tests := []struct {
		name            string
		dymName         string
		owner           string
		controller      string
		expireAt        int64
		configs         []DymNameConfig
		contact         string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:       "pass - valid dym name",
			dymName:    "my-name",
			owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			expireAt:   time.Now().Unix(),
			configs: []DymNameConfig{
				{
					Type:  DymNameConfigType_NAME,
					Path:  "",
					Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				},
				{
					Type:  DymNameConfigType_NAME,
					Path:  "www",
					Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				},
			},
			contact: "contact@example.com",
		},
		{
			name:       "fail - empty name",
			dymName:    "",
			owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			expireAt:   time.Now().Unix(),
			configs: []DymNameConfig{
				{
					Type:  DymNameConfigType_NAME,
					Path:  "",
					Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				},
				{
					Type:  DymNameConfigType_NAME,
					Path:  "www",
					Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				},
			},
			wantErr:         true,
			wantErrContains: "name is empty",
		},
		{
			name:       "fail - bad name",
			dymName:    "-a",
			owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			expireAt:   time.Now().Unix(),
			configs: []DymNameConfig{
				{
					Type:  DymNameConfigType_NAME,
					Path:  "",
					Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				},
				{
					Type:  DymNameConfigType_NAME,
					Path:  "www",
					Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				},
			},
			wantErr:         true,
			wantErrContains: "name is not a valid dym name",
		},
		{
			name:       "fail - empty owner",
			dymName:    "my-name",
			owner:      "",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			expireAt:   time.Now().Unix(),
			configs: []DymNameConfig{
				{
					Type:  DymNameConfigType_NAME,
					Path:  "",
					Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				},
				{
					Type:  DymNameConfigType_NAME,
					Path:  "www",
					Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				},
			},
			wantErr:         true,
			wantErrContains: "owner is empty",
		},
		{
			name:       "fail - bad owner",
			dymName:    "my-name",
			owner:      "dym1fl48vsnmsdzcv85q5",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			expireAt:   time.Now().Unix(),
			configs: []DymNameConfig{
				{
					Type:  DymNameConfigType_NAME,
					Path:  "",
					Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				},
				{
					Type:  DymNameConfigType_NAME,
					Path:  "www",
					Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				},
			},
			wantErr:         true,
			wantErrContains: "owner is not a valid bech32 account address",
		},
		{
			name:       "fail - empty controller",
			dymName:    "my-name",
			owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller: "",
			expireAt:   time.Now().Unix(),
			configs: []DymNameConfig{
				{
					Type:  DymNameConfigType_NAME,
					Path:  "",
					Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				},
				{
					Type:  DymNameConfigType_NAME,
					Path:  "www",
					Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				},
			},
			wantErr:         true,
			wantErrContains: "controller is empty",
		},
		{
			name:       "fail - bad controller",
			dymName:    "my-name",
			owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller: "dym1fl48vsnmsdzcv85q5",
			expireAt:   time.Now().Unix(),
			configs: []DymNameConfig{
				{
					Type:  DymNameConfigType_NAME,
					Path:  "",
					Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				},
				{
					Type:  DymNameConfigType_NAME,
					Path:  "www",
					Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				},
			},
			wantErr:         true,
			wantErrContains: "controller is not a valid bech32 account address",
		},
		{
			name:       "fail - empty expire at",
			dymName:    "my-name",
			owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			expireAt:   0,
			configs: []DymNameConfig{
				{
					Type:  DymNameConfigType_NAME,
					Path:  "",
					Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				},
				{
					Type:  DymNameConfigType_NAME,
					Path:  "www",
					Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				},
			},
			wantErr:         true,
			wantErrContains: "expiry is empty",
		},
		{
			name:       "pass - valid dym name without config",
			dymName:    "my-name",
			owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			expireAt:   time.Now().Unix(),
		},
		{
			name:       "fail - bad config, value must be valid bech32 account address when chain-id is empty",
			dymName:    "my-name",
			owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			expireAt:   time.Now().Unix(),
			configs: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   "dym1fl48vsnmsdzcv85q5d2",
				},
			},
			wantErr:         true,
			wantErrContains: "dym name config value must be a valid bech32 account address",
		},
		{
			name:       "fail - bad config, value must be possible account address when chain-id is not empty",
			dymName:    "my-name",
			owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			expireAt:   time.Now().Unix(),
			configs: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "nim_1122-1",
					Path:    "",
					Value:   "@",
				},
			},
			wantErr:         true,
			wantErrContains: "dym name config value",
		},
		{
			name:       "fail - duplicate config",
			dymName:    "my-name",
			owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			expireAt:   time.Now().Unix(),
			configs: []DymNameConfig{
				{
					Type:  DymNameConfigType_NAME,
					Path:  "www",
					Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
				},
				{
					Type:  DymNameConfigType_NAME,
					Path:  "www",
					Value: "dym1tygms3xhhs3yv487phx3dw4a95jn7t7lnxec2d",
				},
			},
			wantErr:         true,
			wantErrContains: "dym name config is not unique",
		},
		{
			name:       "pass - contact is optional, provided",
			dymName:    "my-name",
			owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			expireAt:   time.Now().Unix(),
			contact:    "contact@example.com",
		},
		{
			name:       "pass - contact is optional, not provided",
			dymName:    "my-name",
			owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			expireAt:   time.Now().Unix(),
			contact:    "",
		},
		{
			name:            "fail - bad contact, too long",
			dymName:         "my-name",
			owner:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			controller:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			expireAt:        time.Now().Unix(),
			contact:         "123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901",
			wantErr:         true,
			wantErrContains: "invalid contact length",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &DymName{
				Name:       tt.dymName,
				Owner:      tt.owner,
				Controller: tt.controller,
				ExpireAt:   tt.expireAt,
				Configs:    tt.configs,
				Contact:    tt.contact,
			}
			err := m.Validate()
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)
		})
	}

	t.Run("maximum number of config", func(t *testing.T) {
		m := &DymName{
			Name:       "a",
			Owner:      "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			Controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			ExpireAt:   1,
		}

		for i := 0; i < MaxConfigSize+1; i++ {
			m.Configs = append(m.Configs, DymNameConfig{
				Type:  DymNameConfigType_NAME,
				Path:  fmt.Sprintf("s%d", i),
				Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			})
		}

		err := m.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "maximum number of configs allowed")
	})
}

func TestDymNameConfig_Validate(t *testing.T) {
	t.Run("nil obj", func(t *testing.T) {
		m := (*DymNameConfig)(nil)
		require.Error(t, m.Validate())
	})

	//goland:noinspection SpellCheckingInspection
	tests := []struct {
		name            string
		Type            DymNameConfigType
		ChainId         string
		Path            string
		Value           string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:    "pass - valid name config",
			Type:    DymNameConfigType_NAME,
			ChainId: "dymension_1100-1",
			Path:    "abc",
			Value:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:  "pass - valid name config with multi-level path",
			Type:  DymNameConfigType_NAME,
			Path:  "abc.def",
			Value: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:    "pass - valid name config with empty path",
			Type:    DymNameConfigType_NAME,
			ChainId: "dymension_1100-1",
			Path:    "",
			Value:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:    "pass - valid name config with empty chain-id",
			Type:    DymNameConfigType_NAME,
			ChainId: "",
			Path:    "abc",
			Value:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:    "pass - valid name config with empty chain-id and path",
			Type:    DymNameConfigType_NAME,
			ChainId: "",
			Path:    "",
			Value:   "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
		},
		{
			name:            "reject - not accept hex address value on host-chain",
			Type:            DymNameConfigType_NAME,
			ChainId:         "",
			Path:            "",
			Value:           "0x1234567890123456789012345678901234567890",
			wantErr:         true,
			wantErrContains: "must be a valid bech32 account address",
		},
		{
			name:            "reject - not accept hex address value on host-chain",
			Type:            DymNameConfigType_NAME,
			ChainId:         "",
			Path:            "a",
			Value:           "0x1234567890123456789012345678901234567890",
			wantErr:         true,
			wantErrContains: "must be a valid bech32 account address",
		},
		{
			name:            "reject - not accept other address value on host-chain",
			Type:            DymNameConfigType_NAME,
			ChainId:         "",
			Path:            "",
			Value:           "t1Rv4exT7bqhZqi2j7xz8bUHDMxwosrjADU",
			wantErr:         true,
			wantErrContains: "dym name config value on host-chain must be lowercase",
		},
		{
			name:            "reject - not accept other address value on host-chain",
			Type:            DymNameConfigType_NAME,
			ChainId:         "",
			Path:            "a",
			Value:           "t1Rv4exT7bqhZqi2j7xz8bUHDMxwosrjADU",
			wantErr:         true,
			wantErrContains: "dym name config value on host-chain must be lowercase",
		},
		{
			name:    "pass - accept hex address value on non-host-chain",
			Type:    DymNameConfigType_NAME,
			ChainId: "another",
			Path:    "",
			Value:   "0x1234567890123456789012345678901234567890",
			wantErr: false,
		},
		{
			name:    "pass - accept hex address value on non-host-chain",
			Type:    DymNameConfigType_NAME,
			ChainId: "another",
			Path:    "a",
			Value:   "0x1234567890123456789012345678901234567890",
			wantErr: false,
		},
		{
			name:    "pass - accept other address value on non-host-chain",
			Type:    DymNameConfigType_NAME,
			ChainId: "another",
			Path:    "",
			Value:   "t1Rv4exT7bqhZqi2j7xz8bUHDMxwosrjADU",
			wantErr: false,
		},
		{
			name:    "pass - accept other address value on non-host-chain",
			Type:    DymNameConfigType_NAME,
			ChainId: "another",
			Path:    "a",
			Value:   "t1Rv4exT7bqhZqi2j7xz8bUHDMxwosrjADU",
			wantErr: false,
		},
		{
			name:            "fail - reject bad address value on non-host-chain",
			Type:            DymNameConfigType_NAME,
			ChainId:         "another",
			Path:            "",
			Value:           "@@",
			wantErr:         true,
			wantErrContains: "dym name config value: @@",
		},
		{
			name:            "fail - reject bad address value on non-host-chain",
			Type:            DymNameConfigType_NAME,
			ChainId:         "another",
			Path:            "a",
			Value:           "@@",
			wantErr:         true,
			wantErrContains: "dym name config value: @@",
		},
		{
			name:            "reject - not accept unknown type",
			Type:            DymNameConfigType_UNKNOWN,
			ChainId:         "",
			Path:            "",
			Value:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "dym name config type must be",
		},
		{
			name:            "reject - bad chain-id",
			Type:            DymNameConfigType_NAME,
			ChainId:         "dymension_",
			Path:            "abc",
			Value:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "dym name config chain id must be a valid chain id format",
		},
		{
			name:            "reject - bad path",
			Type:            DymNameConfigType_NAME,
			ChainId:         "",
			Path:            "-a",
			Value:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "dym name config path must be a valid dym name",
		},
		{
			name:            "reject - bad multi-level path",
			Type:            DymNameConfigType_NAME,
			ChainId:         "",
			Path:            "a.b.",
			Value:           "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "dym name config path must be a valid dym name",
		},
		{
			name:    "pass - value can be empty",
			Type:    DymNameConfigType_NAME,
			ChainId: "",
			Path:    "a",
			Value:   "",
		},
		{
			name:    "pass - value can be empty",
			Type:    DymNameConfigType_NAME,
			ChainId: "",
			Path:    "",
			Value:   "",
		},
		{
			name:            "reject - bad value on host-chain",
			Type:            DymNameConfigType_NAME,
			ChainId:         "",
			Path:            "a",
			Value:           "0x01",
			wantErr:         true,
			wantErrContains: "dym name config value must be a valid bech32 account address",
		},
		{
			name:            "reject - reject value not normalized on host-chain",
			Type:            DymNameConfigType_NAME,
			ChainId:         "",
			Path:            "",
			Value:           "Dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "must be lowercase",
		},
		{
			name:            "reject - reject value not normalized on host-chain",
			Type:            DymNameConfigType_NAME,
			ChainId:         "",
			Path:            "a",
			Value:           "Dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			wantErr:         true,
			wantErrContains: "must be lowercase",
		},
		{
			name:    "pass - allow not-lowercased on non-host-chain",
			Type:    DymNameConfigType_NAME,
			ChainId: "another",
			Path:    "",
			Value:   "t1Rv4exT7bqhZqi2j7xz8bUHDMxwosrjADU",
			wantErr: false,
		},
		{
			name:    "pass - allow not-lowercased on non-host-chain",
			Type:    DymNameConfigType_NAME,
			ChainId: "another",
			Path:    "a",
			Value:   "t1Rv4exT7bqhZqi2j7xz8bUHDMxwosrjADU",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &DymNameConfig{
				Type:    tt.Type,
				ChainId: tt.ChainId,
				Path:    tt.Path,
				Value:   tt.Value,
			}

			err := m.Validate()
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestReverseLookupDymNames_Validate(t *testing.T) {
	t.Run("nil obj", func(t *testing.T) {
		m := (*ReverseLookupDymNames)(nil)
		require.Error(t, m.Validate())
	})

	tests := []struct {
		name            string
		DymNames        []string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:     "valid reverse lookup record",
			DymNames: []string{"my-name", "not-bonded-pool"},
		},
		{
			name:     "allow empty",
			DymNames: []string{},
		},
		{
			name:            "bad dym name",
			DymNames:        []string{"my-name", "-not-bonded-pool"},
			wantErr:         true,
			wantErrContains: "invalid dym name:",
		},
		{
			name:            "bad dym name",
			DymNames:        []string{"-a"},
			wantErr:         true,
			wantErrContains: "invalid dym name:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &ReverseLookupDymNames{
				DymNames: tt.DymNames,
			}

			err := m.Validate()
			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestDymName_IsExpiredAt(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		contextTime time.Time
		wantExpired bool
	}{
		{
			name:        "future",
			contextTime: now.Add(-time.Second),
			wantExpired: false,
		},
		{
			name:        "past",
			contextTime: now.Add(+time.Second),
			wantExpired: true,
		},
		{
			name:        "present",
			contextTime: now,
			wantExpired: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := sdk.Context{}.WithBlockTime(tt.contextTime)
			require.Equal(t, tt.wantExpired, DymName{
				ExpireAt: now.Unix(),
			}.IsExpiredAtCtx(ctx))
		})
	}
}

func TestDymName_IsProhibitedTradingAt(t *testing.T) {
	now := time.Now().UTC()
	require.False(t, DymName{
		ExpireAt: now.Unix() + 10,
	}.IsProhibitedTradingAt(now, 5*time.Second))
	require.True(t, DymName{
		ExpireAt: now.Unix() + 10,
	}.IsProhibitedTradingAt(now, 15*time.Second))
}

func TestDymName_GetSdkEvent(t *testing.T) {
	event := DymName{
		Name:       "a",
		Owner:      "b",
		Controller: "c",
		ExpireAt:   time.Date(2024, 0o1, 0o2, 0o3, 0o4, 0o5, 0, time.UTC).Unix(),
		Configs:    []DymNameConfig{{}, {}},
		Contact:    "contact@example.com",
	}.GetSdkEvent()
	require.NotNil(t, event)
	require.Equal(t, EventTypeSetDymName, event.Type)
	require.Len(t, event.Attributes, 6)
	require.Equal(t, AttributeKeyDymName, event.Attributes[0].Key)
	require.Equal(t, "a", event.Attributes[0].Value)
	require.Equal(t, AttributeKeyDymNameOwner, event.Attributes[1].Key)
	require.Equal(t, "b", event.Attributes[1].Value)
	require.Equal(t, AttributeKeyDymNameController, event.Attributes[2].Key)
	require.Equal(t, "c", event.Attributes[2].Value)
	require.Equal(t, AttributeKeyDymNameExpiryEpoch, event.Attributes[3].Key)
	require.Equal(t, "1704164645", event.Attributes[3].Value)
	require.Equal(t, AttributeKeyDymNameConfigCount, event.Attributes[4].Key)
	require.Equal(t, "2", event.Attributes[4].Value)
	require.Equal(t, AttributeKeyDymNameHasContactDetails, event.Attributes[5].Key)
	require.Equal(t, "true", event.Attributes[5].Value)
}

func TestDymNameConfig_GetIdentity(t *testing.T) {
	tests := []struct {
		name    string
		_type   DymNameConfigType
		chainId string
		path    string
		value   string
		want    string
	}{
		{
			name:    "combination of Type & Chain Id & Path, exclude Value",
			_type:   DymNameConfigType_NAME,
			chainId: "1",
			path:    "2",
			value:   "3",
			want:    "name|1|2",
		},
		{
			name:    "combination of Type & Chain Id & Path, exclude Value",
			_type:   DymNameConfigType_NAME,
			chainId: "1",
			path:    "2",
			value:   "",
			want:    "name|1|2",
		},
		{
			name:    "normalize material fields",
			_type:   DymNameConfigType_NAME,
			chainId: "AaA",
			path:    "bBb",
			value:   "",
			want:    "name|aaa|bbb",
		},
		{
			name:    "use String() of type",
			_type:   DymNameConfigType_UNKNOWN,
			chainId: "1",
			path:    "2",
			want:    "unknown|1|2",
		},
		{
			name:    "use String() of type",
			_type:   DymNameConfigType_NAME,
			chainId: "1",
			path:    "2",
			want:    "name|1|2",
		},
		{
			name:    "respect empty chain-id",
			_type:   DymNameConfigType_NAME,
			chainId: "",
			path:    "2",
			want:    "name||2",
		},
		{
			name:    "respect empty path",
			_type:   DymNameConfigType_NAME,
			chainId: "1",
			path:    "",
			want:    "name|1|",
		},
		{
			name:    "respect empty chain-id and path",
			_type:   DymNameConfigType_NAME,
			chainId: "",
			path:    "",
			want:    "name||",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := DymNameConfig{
				Type:    tt._type,
				ChainId: tt.chainId,
				Path:    tt.path,
				Value:   tt.value,
			}
			require.Equal(t, tt.want, m.GetIdentity())
		})
	}

	t.Run("normalize material fields", func(t *testing.T) {
		require.Equal(t, DymNameConfig{
			ChainId: "AaA",
			Path:    "bBb",
			Value:   "123",
		}.GetIdentity(), DymNameConfig{
			ChainId: "aAa",
			Path:    "BbB",
			Value:   "456",
		}.GetIdentity())
	})
}

func TestDymNameConfig_IsDelete(t *testing.T) {
	require.True(t, DymNameConfig{
		Value: "",
	}.IsDelete(), "if value is empty then it's delete")
	require.False(t, DymNameConfig{
		Value: "1",
	}.IsDelete(), "if value is not empty then it's not delete")
}

//goland:noinspection SpellCheckingInspection
func TestDymName_GetAddressesForReverseMapping(t *testing.T) {
	const dymName = "a"
	const ownerBech32 = "dym1zg69v7yszg69v7yszg69v7yszg69v7ys8xdv96"
	const ownerBech32AtNim = "nim1zg69v7yszg69v7yszg69v7yszg69v7yspkhdt9"
	const ownerHex = "0x1234567890123456789012345678901234567890"
	const bondedPoolBech32 = "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue"
	const bondedPoolHex = "0x4fea76427b8345861e80a3540a8a9d936fd39391"

	const icaBech32 = "dym1zg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg6qrz80ul"
	const icaBech32AtNim = "nim1zg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg69v7yszg6qe9zz9m"
	const icaHex = "0x1234567890123456789012345678901234567890123456789012345678901234"

	tests := []struct {
		name                    string
		configs                 []DymNameConfig
		wantPanic               bool
		wantConfiguredAddresses map[string][]DymNameConfig
		wantFallbackAddresses   map[string][]DymNameConfig
	}{
		{
			name: "pass",
			configs: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   ownerBech32,
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "nim_1122-1",
					Path:    "",
					Value:   ownerBech32AtNim,
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "my-name",
					Value:   bondedPoolBech32,
				},
			},
			wantConfiguredAddresses: map[string][]DymNameConfig{
				ownerBech32: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   ownerBech32,
					},
				},
				ownerBech32AtNim: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "nim_1122-1",
						Path:    "",
						Value:   ownerBech32AtNim,
					},
				},
				bondedPoolBech32: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "my-name",
						Value:   bondedPoolBech32,
					},
				},
			},
			wantFallbackAddresses: map[string][]DymNameConfig{
				ownerHex: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   ownerBech32,
					},
				},
			},
		},
		{
			name: "pass - fallback address is parsed correctly",
			configs: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   ownerBech32,
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "my-name",
					Value:   bondedPoolBech32,
				},
			},
			wantConfiguredAddresses: map[string][]DymNameConfig{
				ownerBech32: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   ownerBech32,
					},
				},
				bondedPoolBech32: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "my-name",
						Value:   bondedPoolBech32,
					},
				},
			},
			wantFallbackAddresses: map[string][]DymNameConfig{
				ownerHex: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   ownerBech32,
					},
				},
			},
		},
		{
			name: "pass - configured bech32 address is kept as is",
			configs: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   ownerBech32,
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "nim_1122-1",
					Path:    "",
					Value:   ownerBech32AtNim, // not dym1, it's nim1
				},
			},
			wantConfiguredAddresses: map[string][]DymNameConfig{
				ownerBech32: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   ownerBech32,
					},
				},
				ownerBech32AtNim: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "nim_1122-1",
						Path:    "",
						Value:   ownerBech32AtNim,
					},
				},
			},
			wantFallbackAddresses: map[string][]DymNameConfig{
				ownerHex: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   ownerBech32,
					},
				},
			},
		},
		{
			name: "pass - able to detect default config address when not configured",
			configs: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "my-name",
					Value:   bondedPoolBech32,
				},
				// not include default config
			},
			wantConfiguredAddresses: map[string][]DymNameConfig{
				ownerBech32: { // default config resolved to owner
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   ownerBech32,
					},
				},
				bondedPoolBech32: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "my-name",
						Value:   bondedPoolBech32,
					},
				},
			},
			wantFallbackAddresses: map[string][]DymNameConfig{
				ownerHex: { // default config resolved to owner
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   ownerBech32,
					},
				},
			},
		},
		{
			name: "pass - respect default config when it is not owner",
			configs: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   bondedPoolBech32, // not the owner
				},
			},
			wantConfiguredAddresses: map[string][]DymNameConfig{
				bondedPoolBech32: { // respect
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   bondedPoolBech32,
					},
				},
			},
			wantFallbackAddresses: map[string][]DymNameConfig{
				bondedPoolHex: { // respect
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   bondedPoolBech32,
					},
				},
			},
		},
		{
			name: "pass - respect default config when it is not owner",
			configs: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   bondedPoolBech32, // not the owner
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "a",
					Value:   bondedPoolBech32,
				},
			},
			wantConfiguredAddresses: map[string][]DymNameConfig{
				bondedPoolBech32: { // respect
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   bondedPoolBech32,
					},
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "a",
						Value:   bondedPoolBech32,
					},
				},
			},
			wantFallbackAddresses: map[string][]DymNameConfig{
				bondedPoolHex: { // respect
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   bondedPoolBech32,
					},
				},
			},
		},
		{
			name: "pass - respect default config when it is not owner",
			configs: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   bondedPoolBech32, // not owner
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "nim_1122-1",
					Path:    "",
					Value:   ownerBech32AtNim, // but this is owner, in different bech32 prefix
				},
			},
			wantConfiguredAddresses: map[string][]DymNameConfig{
				ownerBech32AtNim: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "nim_1122-1",
						Path:    "",
						Value:   ownerBech32AtNim,
					},
				},
				bondedPoolBech32: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   bondedPoolBech32,
					},
				},
			},
			wantFallbackAddresses: map[string][]DymNameConfig{
				bondedPoolHex: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   bondedPoolBech32,
					},
				},
			},
		},
		{
			name: "pass - non-default config will not have fallback records",
			configs: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   ownerBech32,
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "cosmoshub-4",
					Path:    "",
					Value:   "cosmos1tygms3xhhs3yv487phx3dw4a95jn7t7lpm470r",
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "my-name",
					Value:   bondedPoolBech32,
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "blumbus_111-1",
					Path:    "",
					Value:   bondedPoolBech32,
				},
			},
			wantConfiguredAddresses: map[string][]DymNameConfig{
				ownerBech32: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   ownerBech32,
					},
				},
				"cosmos1tygms3xhhs3yv487phx3dw4a95jn7t7lpm470r": {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "cosmoshub-4",
						Path:    "",
						Value:   "cosmos1tygms3xhhs3yv487phx3dw4a95jn7t7lpm470r",
					},
				},
				bondedPoolBech32: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "my-name",
						Value:   bondedPoolBech32,
					},
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "blumbus_111-1",
						Path:    "",
						Value:   bondedPoolBech32,
					},
				},
			},
			wantFallbackAddresses: map[string][]DymNameConfig{
				ownerHex: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   ownerBech32,
					},
				},
			},
		},
		{
			name:      "fail - not accept malformed config",
			configs:   []DymNameConfig{{}},
			wantPanic: true,
		},
		{
			name: "fail - not accept malformed config, not bech32 address value",
			configs: []DymNameConfig{{
				Type:    DymNameConfigType_NAME,
				ChainId: "",
				Path:    "a",
				Value:   "0x1234567890123456789012345678901234567890",
			}},
			wantPanic: true,
		},
		{
			name: "fail - not accept malformed config, default config is not bech32 address of host",
			configs: []DymNameConfig{{
				Type:    DymNameConfigType_NAME,
				ChainId: "",
				Path:    "",
				Value:   ownerBech32AtNim,
			}},
			wantPanic: true,
		},
		{
			name: "fail - not accept malformed config, not valid bech32 address",
			configs: []DymNameConfig{{
				Type:    DymNameConfigType_NAME,
				ChainId: "",
				Path:    "a",
				Value:   ownerBech32 + "a",
			}},
			wantPanic: true,
		},
		{
			name: "fail - not accept malformed config, default config is not bech32 address of host",
			configs: []DymNameConfig{{
				Type:    DymNameConfigType_NAME,
				ChainId: "",
				Path:    "",
				Value:   ownerBech32 + "a",
			}},
			wantPanic: true,
		},
		{
			name: "pass - ignore empty value config",
			configs: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   ownerBech32,
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "nim_1122-1",
					Path:    "",
					Value:   ownerBech32AtNim,
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "my-name",
					Value:   "", // empty value
				},
			},
			wantConfiguredAddresses: map[string][]DymNameConfig{
				ownerBech32: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   ownerBech32,
					},
				},
				ownerBech32AtNim: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "nim_1122-1",
						Path:    "",
						Value:   ownerBech32AtNim,
					},
				},
			},
			wantFallbackAddresses: map[string][]DymNameConfig{
				ownerHex: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   ownerBech32,
					},
				},
			},
		},
		{
			name: "pass - ignore empty value default config",
			configs: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   "", // empty value
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "nim_1122-1",
					Path:    "",
					Value:   ownerBech32AtNim,
				},
			},
			wantConfiguredAddresses: map[string][]DymNameConfig{
				ownerBech32: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   ownerBech32, // detected & automatically filled default config
					},
				},
				ownerBech32AtNim: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "nim_1122-1",
						Path:    "",
						Value:   ownerBech32AtNim,
					},
				},
			},
			wantFallbackAddresses: map[string][]DymNameConfig{
				ownerHex: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   ownerBech32, // detected & automatically filled default config
					},
				},
			},
		},
		{
			name: "pass - allow Interchain Account",
			configs: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   icaBech32,
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "nim_1122-1",
					Path:    "ica",
					Value:   icaBech32AtNim,
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "nim_1122-1",
					Path:    "",
					Value:   ownerBech32AtNim,
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "my-name",
					Value:   bondedPoolBech32,
				},
			},
			wantConfiguredAddresses: map[string][]DymNameConfig{
				ownerBech32AtNim: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "nim_1122-1",
						Path:    "",
						Value:   ownerBech32AtNim,
					},
				},
				bondedPoolBech32: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "my-name",
						Value:   bondedPoolBech32,
					},
				},
				icaBech32: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   icaBech32,
					},
				},
				icaBech32AtNim: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "nim_1122-1",
						Path:    "ica",
						Value:   icaBech32AtNim,
					},
				},
			},
			wantFallbackAddresses: map[string][]DymNameConfig{
				icaHex: {
					{
						Type:    DymNameConfigType_NAME,
						ChainId: "",
						Path:    "",
						Value:   icaBech32,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &DymName{
				Name:       dymName,
				Owner:      ownerBech32,
				Controller: ownerBech32,
				ExpireAt:   1,
				Configs:    tt.configs,
			}

			if tt.wantPanic {
				require.Panics(t, func() {
					_, _ = m.GetAddressesForReverseMapping()
				})

				return
			}

			gotConfiguredAddresses, gotFallbackAddresses := m.GetAddressesForReverseMapping()

			if !reflect.DeepEqual(tt.wantConfiguredAddresses, gotConfiguredAddresses) {
				t.Errorf("gotConfiguredAddresses = %v, want %v", gotConfiguredAddresses, tt.wantConfiguredAddresses)
			}
			if !reflect.DeepEqual(tt.wantFallbackAddresses, gotFallbackAddresses) {
				t.Errorf("gotFallbackAddresses = %v, want %v", gotFallbackAddresses, tt.wantFallbackAddresses)
			}

			if len(gotFallbackAddresses) > 0 {
				require.Len(t, gotFallbackAddresses, 1, "there is only one default config, therefor only one fallback address")
			}
			for fba, cfgs := range gotFallbackAddresses {
				require.True(t, dymnsutils.IsValidHexAddress(fba))
				for _, cfg := range cfgs {
					require.True(t, cfg.IsDefaultNameConfig())
					require.NotEmpty(t, cfg.Value != "")
				}
			}
		})
	}
}

func TestDymNameConfig_IsDefaultNameConfig(t *testing.T) {
	tests := []struct {
		name    string
		_type   DymNameConfigType
		chainId string
		path    string
		value   string
		want    bool
	}{
		{
			name:    "default name config",
			_type:   DymNameConfigType_NAME,
			chainId: "",
			path:    "",
			value:   "x",
			want:    true,
		},
		{
			name:    "default name config, value can be empty",
			_type:   DymNameConfigType_NAME,
			chainId: "",
			path:    "",
			value:   "",
			want:    true,
		},
		{
			name:    "config with type != name is not default name config",
			_type:   DymNameConfigType_UNKNOWN,
			chainId: "",
			path:    "",
			value:   "x",
			want:    false,
		},
		{
			name:    "config with type != name is not default name config",
			_type:   DymNameConfigType_UNKNOWN,
			chainId: "",
			path:    "",
			value:   "",
			want:    false,
		},
		{
			name:    "config with type != name is not default name config",
			_type:   DymNameConfigType_UNKNOWN,
			chainId: "",
			path:    "x",
			value:   "",
			want:    false,
		},
		{
			name:    "non-empty chain-id is not default name config",
			_type:   DymNameConfigType_NAME,
			chainId: "x",
			path:    "",
			value:   "x",
			want:    false,
		},
		{
			name:    "non-empty path is not default name config",
			_type:   DymNameConfigType_NAME,
			chainId: "",
			path:    "x",
			value:   "x",
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := DymNameConfig{
				Type:    tt._type,
				ChainId: tt.chainId,
				Path:    tt.path,
				Value:   tt.value,
			}
			require.Equal(t, tt.want, m.IsDefaultNameConfig())
		})
	}
}

func TestDymNameConfigs_DefaultNameConfigs(t *testing.T) {
	tests := []struct {
		name                  string
		m                     DymNameConfigs
		dropEmptyValueConfigs bool
		want                  DymNameConfigs
	}{
		{
			name: "pass",
			m: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "b",
					Path:    "b",
					Value:   "b",
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   "a",
				},
			},
			want: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   "a",
				},
			},
		},
		{
			name: "pass - empty",
			m:    []DymNameConfig{},
			want: DymNameConfigs{},
		},
		{
			name: "pass - nil",
			m:    nil,
			want: DymNameConfigs{},
		},
		{
			name: "pass - none",
			m: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "b",
					Path:    "b",
					Value:   "b",
				},
			},
			want: DymNameConfigs{},
		},
		{
			name: "pass - multiple of more than one",
			m: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   "a",
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "b",
					Path:    "b",
					Value:   "b",
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   "c",
				},
			},
			want: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   "a",
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   "c",
				},
			},
		},
		{
			name: "pass - name config only",
			m: []DymNameConfig{
				{
					Type:    DymNameConfigType_UNKNOWN,
					ChainId: "",
					Path:    "",
					Value:   "a",
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "b",
					Path:    "b",
					Value:   "b",
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   "c",
				},
			},
			want: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   "c",
				},
			},
		},
		{
			name: "pass - drop empty value configs",
			m: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   "a",
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   "",
				},
			},
			dropEmptyValueConfigs: true,
			want: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   "a",
				},
			},
		},
		{
			name: "pass - drop empty value configs",
			m: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   "a",
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   "",
				},
			},
			dropEmptyValueConfigs: false,
			want: []DymNameConfig{
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   "a",
				},
				{
					Type:    DymNameConfigType_NAME,
					ChainId: "",
					Path:    "",
					Value:   "",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.m.DefaultNameConfigs(tt.dropEmptyValueConfigs)
			if len(tt.want) == 0 {
				require.Empty(t, got)
			} else {
				require.Equal(t, tt.want, got)
			}
		})
	}
}
