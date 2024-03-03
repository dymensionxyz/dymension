package types

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	// ProposalTypeSubmitFraud defines the type for a SubmitFraudProposal
	ProposalTypeSubmitFraud = "SubmitFraud"
)

// Assert SubmitFraudProposal implements govtypes.Content at compile-time
var _ govtypes.Content = &SubmitFraudProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeSubmitFraud)
}

// NewSubmitFraudProposal creates a new submit fraud proposal.
//
//nolint:interfacer
func NewSubmitFraudProposal(title, description, rollappId string, height uint64, seqaddr string) *SubmitFraudProposal {
	return &SubmitFraudProposal{
		Title:                      title,
		Description:                description,
		RollappId:                  rollappId,
		FraudelentHeight:           height,
		FraudelentSequencerAddress: seqaddr,
	}
}

// GetTitle returns the title of a submit fraud proposal.
func (sfp *SubmitFraudProposal) GetTitle() string { return sfp.Title }

// GetDescription returns the description of a submit fraud proposal.
func (sfp *SubmitFraudProposal) GetDescription() string { return sfp.Description }

// ProposalRoute returns the routing key of a submit fraud proposal.
func (sfp *SubmitFraudProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a submit fraud proposal.
func (sfp *SubmitFraudProposal) ProposalType() string { return ProposalTypeSubmitFraud }

// ValidateBasic runs basic stateless validity checks
func (sfp *SubmitFraudProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(sfp)
	if err != nil {
		return err
	}

	if sfp.RollappId == "" {
		return ErrInvalidRollappID
	}

	if sfp.FraudelentHeight == 0 {
		return ErrInvalidHeight
	}

	if sfp.FraudelentSequencerAddress == "" {
		return ErrInvalidSequencer
	}

	return nil
}

// String implements the Stringer interface.
func (sfp SubmitFraudProposal) String() string {
	return sfp.Description
}
