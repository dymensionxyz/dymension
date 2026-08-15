package types

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MaxValidationResponse is the ERC-8004 response ceiling.
const MaxValidationResponse uint32 = 100

func NewMsgRequestValidation(requester, validatorID, agentID string, evidenceSeq uint64, requestHash []byte, requestURI string) *MsgRequestValidation {
	return &MsgRequestValidation{Requester: requester, ValidatorId: validatorID, AgentId: agentID, EvidenceSeq: evidenceSeq, RequestHash: requestHash, RequestUri: requestURI}
}

func (m *MsgRequestValidation) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Requester); err != nil {
		return err
	}
	if m.ValidatorId == "" || m.AgentId == "" {
		return fmt.Errorf("agent ids must not be empty")
	}
	if len(m.RequestHash) != 32 {
		return errorsmod.Wrapf(ErrInvalidValidationHash, "request hash length %d", len(m.RequestHash))
	}
	return nil
}

func NewMsgRespondValidation(responder string, requestHash []byte, response uint32, responseURI string, responseHash []byte, tag string) *MsgRespondValidation {
	return &MsgRespondValidation{Responder: responder, RequestHash: requestHash, Response: response, ResponseUri: responseURI, ResponseHash: responseHash, Tag: tag}
}

func (m *MsgRespondValidation) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Responder); err != nil {
		return err
	}
	if len(m.RequestHash) != 32 {
		return errorsmod.Wrapf(ErrInvalidValidationHash, "request hash length %d", len(m.RequestHash))
	}
	if m.Response > MaxValidationResponse {
		return errorsmod.Wrapf(ErrInvalidValidationResponse, "response %d exceeds max %d", m.Response, MaxValidationResponse)
	}
	if len(m.ResponseHash) != 0 && len(m.ResponseHash) != 32 {
		return errorsmod.Wrapf(ErrInvalidValidationHash, "response hash length %d", len(m.ResponseHash))
	}
	return nil
}
