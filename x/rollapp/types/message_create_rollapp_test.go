package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/stretchr/testify/require"
)

func TestMsgCreateRollapp_ValidateBasic(t *testing.T) {
	defaultMetadata := TokenMetadata{
		Description: "valid",
		DenomUnits: []*DenomUnit{
			{Denom: "uvalid", Exponent: 0},
			{Denom: "valid", Exponent: 18},
		},
		Base:    "uvalid",
		Display: "valid",
		Name:    "valid",
		Symbol:  "VALID",
	}

	seqDupAddr := sample.AccAddress()

	var tooManyAddresses []string
	for i := 0; i < 200; i++ {
		tooManyAddresses = append(tooManyAddresses, sample.AccAddress())
	}
	var validNumberAddresses []string
	for i := 0; i < 100; i++ {
		validNumberAddresses = append(validNumberAddresses, sample.AccAddress())
	}
	tests := []struct {
		name string
		msg  MsgCreateRollapp
		err  error
	}{
		{
			name: "valid - full features",
			msg: MsgCreateRollapp{
				Creator:               sample.AccAddress(),
				MaxSequencers:         2,
				RollappId:             "dym_100-1",
				PermissionedAddresses: []string{sample.AccAddress(), sample.AccAddress()},
				Metadatas:             []TokenMetadata{defaultMetadata},
				GenesisAccounts: []GenesisAccount{
					{
						Address: sample.AccAddress(),
						Amount:  sdk.NewCoin("valid", sdk.NewInt(1000)),
					},
				},
			},
		},
		{
			name: "invalid rollappID",
			msg: MsgCreateRollapp{
				Creator:       sample.AccAddress(),
				MaxSequencers: 1,
				RollappId:     " ",
			},
			err: ErrInvalidRollappID,
		},
		{
			name: "invalid address",
			msg: MsgCreateRollapp{
				Creator:       "invalid_address",
				MaxSequencers: 1,
				RollappId:     "dym_100-1",
			},
			err: ErrInvalidCreatorAddress,
		},
		{
			name: "valid address",
			msg: MsgCreateRollapp{
				Creator:       sample.AccAddress(),
				MaxSequencers: 1,
				RollappId:     "dym_100-1",
			},
		},
		{
			name: "no max sequencers set",
			msg: MsgCreateRollapp{
				Creator:   sample.AccAddress(),
				RollappId: "dym_100-1",
			},
		},
		{
			name: "valid permissioned addresses",
			msg: MsgCreateRollapp{
				Creator:               sample.AccAddress(),
				MaxSequencers:         2,
				RollappId:             "dym_100-1",
				PermissionedAddresses: []string{sample.AccAddress(), sample.AccAddress()},
			},
		},
		{
			name: "duplicate permissioned addresses",
			msg: MsgCreateRollapp{
				Creator:               sample.AccAddress(),
				MaxSequencers:         2,
				RollappId:             "dym_100-1",
				PermissionedAddresses: []string{seqDupAddr, seqDupAddr},
			},
			err: ErrPermissionedAddressesDuplicate,
		},
		{
			name: "invalid permissioned addresses",
			msg: MsgCreateRollapp{
				Creator:               sample.AccAddress(),
				MaxSequencers:         2,
				RollappId:             "dym_100-1",
				PermissionedAddresses: []string{seqDupAddr, "invalid permissioned address"},
			},
			err: ErrInvalidPermissionedAddress,
		},
		{
			name: "valid token metadata",
			msg: MsgCreateRollapp{
				Creator:       sample.AccAddress(),
				MaxSequencers: 1,
				RollappId:     "dym_100-1",
				Metadatas:     []TokenMetadata{defaultMetadata},
			},
			err: nil,
		},
		{
			name: "invalid token metadata", // just trigger one case to see if validation is done or not
			msg: MsgCreateRollapp{
				Creator:       sample.AccAddress(),
				MaxSequencers: 1,
				RollappId:     "dym_100-1",
				Metadatas: []TokenMetadata{{
					Description: "valid",
					DenomUnits: []*DenomUnit{
						{Denom: "uvalid", Exponent: 0},
						{Denom: "valid", Exponent: 18},
					},
					Base:    "uvalid",
					Display: "valid",
					Name:    "", // empty
					Symbol:  "VALID",
				}},
			},
			err: ErrInvalidTokenMetadata,
		},
		{
			name: "invalid genesis account address",
			msg: MsgCreateRollapp{
				Creator:       sample.AccAddress(),
				MaxSequencers: 1,
				RollappId:     "dym_100-1",
				GenesisAccounts: []GenesisAccount{
					{
						Address: "invalid_address",
						Amount:  sdk.NewCoin("valid", sdk.NewInt(1000)),
					},
				},
			},
			err: ErrInvalidGenesisAccount,
		},
		{
			name: "more addresses than sequencers", // just trigger one case to see if validation is done or not
			msg: MsgCreateRollapp{
				Creator:               sample.AccAddress(),
				RollappId:             "dym_100-1",
				MaxSequencers:         1,
				PermissionedAddresses: validNumberAddresses,
			},
			err: ErrTooManyPermissionedAddresses,
		},
		{
			name: "too many sequencers", // just trigger one case to see if validation is done or not
			msg: MsgCreateRollapp{
				Creator:               sample.AccAddress(),
				RollappId:             "dym_100-1",
				MaxSequencers:         200,
				PermissionedAddresses: tooManyAddresses,
			},
			err: ErrInvalidMaxSequencers,
		},
		{
			name: "max sequencer not set",
			msg: MsgCreateRollapp{
				Creator:   sample.AccAddress(),
				RollappId: "dym_100-1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorContains(t, err, tt.err.Error(), "test %s failed", tt.name)
				return
			}
			require.NoError(t, err)
		})
	}
}
