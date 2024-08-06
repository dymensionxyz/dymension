package types

import (
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	// ProposalTypeChangeVMType defines the type for a ChangeVMTypeProposal
	ProposalTypeChangeVMType = "ChangeVMType"
)

// Assert ChangeVMTypeProposal implements govtypes.Content at compile-time
var _ govtypes.Content = &ChangeVMTypeProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeChangeVMType)
}

// NewChangeVMTypeProposal creates a new change vm type proposal.
//
//nolint:interfacer
func NewChangeVMTypeProposal(title, description, rollappId string, vmType string) *ChangeVMTypeProposal {
	return &ChangeVMTypeProposal{
		Title:       title,
		Description: description,
		RollappId:   rollappId,
		VmType:      vmType,
	}
}

// GetTitle returns the title of a change vm type proposal.
func (sfp *ChangeVMTypeProposal) GetTitle() string { return sfp.Title }

// GetDescription returns the description of a change vm type proposal.
func (sfp *ChangeVMTypeProposal) GetDescription() string { return sfp.Description }

// ProposalRoute returns the routing key of a change vm type proposal.
func (sfp *ChangeVMTypeProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a change vm type proposal.
func (sfp *ChangeVMTypeProposal) ProposalType() string { return ProposalTypeChangeVMType }

// ValidateBasic runs basic stateless validity checks
func (sfp *ChangeVMTypeProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(sfp)
	if err != nil {
		return err
	}

	if sfp.RollappId == "" {
		return ErrInvalidRollappID
	}

	if vmt, ok := Rollapp_VMType_value[sfp.VmType]; !ok || vmt == 0 {
		return ErrInvalidVMType
	}

	return nil
}

// String implements the Stringer interface.
func (sfp ChangeVMTypeProposal) String() string {
	return sfp.Description
}
