package types

import (
	"testing"

	"cosmossdk.io/math"
	hyputil "github.com/bcp-innovations/hyperlane-cosmos/util"
	"github.com/stretchr/testify/require"
)

func TestDefaultGenesis(t *testing.T) {
	genesis := DefaultGenesis()
	require.NotNil(t, genesis)
	require.NotNil(t, genesis.FeeHooks)
	require.NotNil(t, genesis.AggregationHooks)
	require.Empty(t, genesis.FeeHooks)
	require.Empty(t, genesis.AggregationHooks)

	// Most importantly, the default genesis should be valid
	require.NoError(t, genesis.Validate())
}

func TestGenesisState_Validate(t *testing.T) {
	validHookId1 := mustHexFromString("0x0000000000000000000000007fa9385be102ac3eac297483dd6233d62b3e1496")
	validHookId2 := mustHexFromString("0x080ef1c2cd394de78363ecb0a466c934b57de4abb5604a0684e571990eb7b073")
	validOwner := CreateRandomAccount().String()
	validFee := HLAssetFee{
		TokenID:     "0x0000000000000000000000007fa9385be102ac3eac297483dd6233d62b3e1496",
		InboundFee:  math.LegacyMustNewDecFromStr("0.01"),
		OutboundFee: math.LegacyMustNewDecFromStr("0.02"),
	}

	tests := []struct {
		name     string
		genesis  GenesisState
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid empty genesis",
			genesis: GenesisState{
				FeeHooks:         []HLFeeHook{},
				AggregationHooks: []AggregationHook{},
			},
			wantErr: false,
		},
		{
			name: "valid genesis with fee hooks",
			genesis: GenesisState{
				FeeHooks: []HLFeeHook{
					{
						Id:    validHookId1,
						Owner: validOwner,
						Fees:  []HLAssetFee{validFee},
					},
				},
				AggregationHooks: []AggregationHook{},
			},
			wantErr: false,
		},
		{
			name: "valid genesis with aggregation hooks",
			genesis: GenesisState{
				FeeHooks: []HLFeeHook{},
				AggregationHooks: []AggregationHook{
					{
						Id:      validHookId1,
						Owner:   validOwner,
						HookIds: []hyputil.HexAddress{validHookId2},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid genesis with both hook types",
			genesis: GenesisState{
				FeeHooks: []HLFeeHook{
					{
						Id:    validHookId1,
						Owner: validOwner,
						Fees:  []HLAssetFee{validFee},
					},
				},
				AggregationHooks: []AggregationHook{
					{
						Id:      validHookId2,
						Owner:   validOwner,
						HookIds: []hyputil.HexAddress{validHookId1},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid genesis with empty fees",
			genesis: GenesisState{
				FeeHooks: []HLFeeHook{
					{
						Id:    validHookId1,
						Owner: validOwner,
						Fees:  []HLAssetFee{}, // empty fees are allowed
					},
				},
				AggregationHooks: []AggregationHook{},
			},
			wantErr: false,
		},
		{
			name: "valid genesis with empty hook ids",
			genesis: GenesisState{
				FeeHooks: []HLFeeHook{},
				AggregationHooks: []AggregationHook{
					{
						Id:      validHookId1,
						Owner:   validOwner,
						HookIds: []hyputil.HexAddress{}, // empty hook ids are allowed
					},
				},
			},
			wantErr: false,
		},
		{
			name: "duplicate hook IDs between fee hooks",
			genesis: GenesisState{
				FeeHooks: []HLFeeHook{
					{
						Id:    validHookId1,
						Owner: validOwner,
						Fees:  []HLAssetFee{validFee},
					},
					{
						Id:    validHookId1, // duplicate ID
						Owner: validOwner,
						Fees:  []HLAssetFee{validFee},
					},
				},
				AggregationHooks: []AggregationHook{},
			},
			wantErr: true,
			errMsg:  "duplicate hook id",
		},
		{
			name: "duplicate hook IDs between aggregation hooks",
			genesis: GenesisState{
				FeeHooks: []HLFeeHook{},
				AggregationHooks: []AggregationHook{
					{
						Id:      validHookId1,
						Owner:   validOwner,
						HookIds: []hyputil.HexAddress{validHookId2},
					},
					{
						Id:      validHookId1, // duplicate ID
						Owner:   validOwner,
						HookIds: []hyputil.HexAddress{validHookId2},
					},
				},
			},
			wantErr: true,
			errMsg:  "duplicate hook id",
		},
		{
			name: "duplicate hook IDs between fee and aggregation hooks",
			genesis: GenesisState{
				FeeHooks: []HLFeeHook{
					{
						Id:    validHookId1,
						Owner: validOwner,
						Fees:  []HLAssetFee{validFee},
					},
				},
				AggregationHooks: []AggregationHook{
					{
						Id:      validHookId1, // duplicate ID with fee hook
						Owner:   validOwner,
						HookIds: []hyputil.HexAddress{validHookId2},
					},
				},
			},
			wantErr: true,
			errMsg:  "duplicate hook id",
		},
		{
			name: "invalid fee hook - invalid owner",
			genesis: GenesisState{
				FeeHooks: []HLFeeHook{
					{
						Id:    validHookId1,
						Owner: "invalid-address",
						Fees:  []HLAssetFee{validFee},
					},
				},
				AggregationHooks: []AggregationHook{},
			},
			wantErr: true,
			errMsg:  "invalid owner",
		},
		{
			name: "invalid fee hook - invalid fee",
			genesis: GenesisState{
				FeeHooks: []HLFeeHook{
					{
						Id:    validHookId1,
						Owner: validOwner,
						Fees: []HLAssetFee{
							{
								TokenID:     "", // invalid empty token ID
								InboundFee:  math.LegacyMustNewDecFromStr("0.01"),
								OutboundFee: math.LegacyMustNewDecFromStr("0.02"),
							},
						},
					},
				},
				AggregationHooks: []AggregationHook{},
			},
			wantErr: true,
			errMsg:  "token id cannot be empty",
		},
		{
			name: "invalid aggregation hook - invalid owner",
			genesis: GenesisState{
				FeeHooks: []HLFeeHook{},
				AggregationHooks: []AggregationHook{
					{
						Id:      validHookId1,
						Owner:   "invalid-address",
						HookIds: []hyputil.HexAddress{validHookId2},
					},
				},
			},
			wantErr: true,
			errMsg:  "owner address is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.genesis.Validate()
			if tt.wantErr {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}