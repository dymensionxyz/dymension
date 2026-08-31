package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func TestMsgRollappFraudProposalValidateBasic(t *testing.T) {
	valid := types.MsgRollappFraudProposal{
		Authority:              sample.AccAddress(),
		RollappId:              "rollapp_1234-1",
		FraudHeight:            1,
		FraudRevision:          0,
		PunishSequencerAddress: sample.AccAddress(),
		Rewardee:               sample.AccAddress(),
	}

	tests := []struct {
		name        string
		modify      func(*types.MsgRollappFraudProposal)
		wantErr     error
		wantMessage string
	}{
		{
			name: "valid message with revision zero",
		},
		{
			name: "empty rollapp ID",
			modify: func(msg *types.MsgRollappFraudProposal) {
				msg.RollappId = ""
			},
			wantErr: gerrc.ErrInvalidArgument,
		},
		{
			name: "malformed rollapp ID",
			modify: func(msg *types.MsgRollappFraudProposal) {
				msg.RollappId = "malformed"
			},
			wantErr: gerrc.ErrInvalidArgument,
		},
		{
			name: "zero fraud height",
			modify: func(msg *types.MsgRollappFraudProposal) {
				msg.FraudHeight = 0
			},
			wantErr: gerrc.ErrInvalidArgument,
		},
		{
			name: "malformed punish sequencer address",
			modify: func(msg *types.MsgRollappFraudProposal) {
				msg.PunishSequencerAddress = "malformed"
			},
			wantErr: gerrc.ErrInvalidArgument,
		},
		{
			name: "empty optional punish sequencer address",
			modify: func(msg *types.MsgRollappFraudProposal) {
				msg.PunishSequencerAddress = ""
			},
		},
		{
			name: "malformed rewardee reports rewardee value",
			modify: func(msg *types.MsgRollappFraudProposal) {
				msg.Rewardee = "malformed-rewardee"
			},
			wantErr:     gerrc.ErrInvalidArgument,
			wantMessage: "malformed-rewardee",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := valid
			if tt.modify != nil {
				tt.modify(&msg)
			}

			err := msg.ValidateBasic()
			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}

			require.ErrorIs(t, err, tt.wantErr)
			if tt.wantMessage != "" {
				require.ErrorContains(t, err, tt.wantMessage)
			}
		})
	}
}
