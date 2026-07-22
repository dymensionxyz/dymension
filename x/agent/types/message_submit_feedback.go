package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// MaxFeedbackScore is the fixed-point score ceiling: 10000 == 100.00%.
const MaxFeedbackScore = 10000

func NewMsgSubmitFeedback(client, agentID string, score uint32, tag1, tag2 string, evidenceSeq uint64) *MsgSubmitFeedback {
	return &MsgSubmitFeedback{
		Client:      client,
		AgentId:     agentID,
		Score:       score,
		Tag1:        tag1,
		Tag2:        tag2,
		EvidenceSeq: evidenceSeq,
	}
}

func (m *MsgSubmitFeedback) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Client); err != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "client address")
	}
	if m.AgentId == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "empty agent id")
	}
	if m.Score > MaxFeedbackScore {
		return errorsmod.Wrapf(ErrInvalidScore, "score %d exceeds max %d", m.Score, MaxFeedbackScore)
	}
	if m.Tag1 == "" {
		return errorsmod.Wrap(ErrInvalidTag, "tag1 is required")
	}
	if err := validateTag(m.Tag1); err != nil {
		return err
	}
	return validateTag(m.Tag2)
}

// validateTag rejects ASCII control chars to keep event/attribute output clean;
// any other UTF-8 is allowed. Length limits are param-driven and enforced in
// the handler.
func validateTag(tag string) error {
	for i := 0; i < len(tag); i++ {
		if tag[i] < 0x20 {
			return errorsmod.Wrap(ErrInvalidTag, "tag contains control characters")
		}
	}
	return nil
}
