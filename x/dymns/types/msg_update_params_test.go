package types

import (
	"fmt"
	"testing"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgUpdateParams_ValidateBasic(t *testing.T) {
	defaultPriceParams := DefaultPriceParams()
	defaultChainsParams := DefaultChainsParams()
	defaultMiscParams := DefaultMiscParams()

	trueAndFalse := []bool{true, false}
	for _, updatePriceParams := range trueAndFalse {
		for _, updateChainsParams := range trueAndFalse {
			for _, updateMiscParams := range trueAndFalse {
				if !updatePriceParams && !updateChainsParams && !updateMiscParams {
					continue
				}

				t.Run(
					fmt.Sprintf(
						"pass - all valid, random params provided: %t|%t|%t",
						updatePriceParams,
						updateChainsParams,
						updateMiscParams,
					),
					func(t *testing.T) {
						m := &MsgUpdateParams{
							Authority: sample.AccAddress(),
						}

						if updatePriceParams {
							m.NewPriceParams = &defaultPriceParams
						}

						if updateChainsParams {
							m.NewChainsParams = &defaultChainsParams
						}

						if updateMiscParams {
							m.NewMiscParams = &defaultMiscParams
						}

						err := m.ValidateBasic()
						require.NoError(t, err)
					})
			}
		}
	}

	tests := []struct {
		name            string
		authority       string
		newPriceParams  *PriceParams
		newChainsParams *ChainsParams
		newMiscParams   *MiscParams
		wantErr         bool
		wantErrContains string
	}{
		{
			name:            "fail - not update any params",
			authority:       sample.AccAddress(),
			newPriceParams:  nil,
			newChainsParams: nil,
			newMiscParams:   nil,
			wantErr:         true,
			wantErrContains: "at least one of the new params must be provided",
		},
		{
			name:            "fail - bad authority",
			authority:       "0x1",
			newPriceParams:  &defaultPriceParams,
			newChainsParams: nil,
			newMiscParams:   nil,
			wantErr:         true,
			wantErrContains: "authority is not a valid bech32 address",
		},
		{
			name:            "fail - bad price params",
			authority:       sample.AccAddress(),
			newPriceParams:  &PriceParams{},
			newChainsParams: nil,
			newMiscParams:   nil,
			wantErr:         true,
			wantErrContains: "failed to validate new price params",
		},
		{
			name:           "fail - bad chain params",
			authority:      sample.AccAddress(),
			newPriceParams: nil,
			newChainsParams: &ChainsParams{
				AliasesOfChainIds: []AliasesOfChainId{
					{
						ChainId: "@@@",
					},
				},
			},
			newMiscParams:   nil,
			wantErr:         true,
			wantErrContains: "failed to validate new chains params",
		},
		{
			name:            "fail - bad misc params",
			authority:       sample.AccAddress(),
			newPriceParams:  nil,
			newChainsParams: nil,
			newMiscParams:   &MiscParams{},
			wantErr:         true,
			wantErrContains: "failed to validate new misc params",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MsgUpdateParams{
				Authority:       tt.authority,
				NewPriceParams:  tt.newPriceParams,
				NewChainsParams: tt.newChainsParams,
				NewMiscParams:   tt.newMiscParams,
			}
			err := m.ValidateBasic()
			if tt.wantErr {
				require.ErrorContains(t, err, tt.wantErrContains)
				return
			}

			require.NoError(t, err)
		})
	}
}
