package types

import (
	"fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	ProposalTypeUpdateLockdrop  = "UpdateLockdrop"
	ProposalTypeReplaceLockdrop = "ReplaceLockdrop"
)

// Init registers proposals to update and replace pool incentives.
func init() {
	govtypes.RegisterProposalType(ProposalTypeUpdateLockdrop)
	govtypes.RegisterProposalType(ProposalTypeReplaceLockdrop)
}

var (
	_ govtypes.Content = &UpdateLockdropProposal{}
	_ govtypes.Content = &ReplaceLockdropProposal{}
)

// NewReplaceLockdropProposal returns a new instance of a replace pool incentives proposal struct.
func NewReplaceLockdropProposal(title, description string, records []DistrRecord) govtypes.Content {
	return &ReplaceLockdropProposal{
		Title:       title,
		Description: description,
		Records:     records,
	}
}

// GetTitle gets the title of the proposal
func (p *ReplaceLockdropProposal) GetTitle() string { return p.Title }

// GetDescription gets the description of the proposal
func (p *ReplaceLockdropProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the router key for the proposal
func (p *ReplaceLockdropProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal
func (p *ReplaceLockdropProposal) ProposalType() string {
	return ProposalTypeReplaceLockdrop
}

// ValidateBasic validates a governance proposal's abstract and basic contents
func (p *ReplaceLockdropProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}
	if len(p.Records) == 0 {
		return ErrEmptyProposalRecords
	}

	for _, record := range p.Records {
		if err := record.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}

// String returns a string containing the pool incentives proposal.
func (p ReplaceLockdropProposal) String() string {
	// TODO: Make this prettier
	recordsStr := ""
	for _, record := range p.Records {
		recordsStr = recordsStr + fmt.Sprintf("(GaugeId: %d, Weight: %s) ", record.GaugeId, record.Weight.String())
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Replace Pool Incentives Proposal:
  Title:       %s
  Description: %s
  Records:     %s
`, p.Title, p.Description, recordsStr))
	return b.String()
}

// NewReplaceLockdropProposal returns a new instance of a replace pool incentives proposal struct.
func NewUpdateLockdropProposal(title, description string, records []DistrRecord) govtypes.Content {
	return &UpdateLockdropProposal{
		Title:       title,
		Description: description,
		Records:     records,
	}
}

// GetTitle gets the title of the proposal
func (p *UpdateLockdropProposal) GetTitle() string { return p.Title }

// GetDescription gets the description of the proposal
func (p *UpdateLockdropProposal) GetDescription() string { return p.Description }

// ProposalRoute returns the router key for the proposal
func (p *UpdateLockdropProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of the proposal
func (p *UpdateLockdropProposal) ProposalType() string { return ProposalTypeUpdateLockdrop }

// ValidateBasic validates a governance proposal's abstract and basic contents.
func (p *UpdateLockdropProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(p)
	if err != nil {
		return err
	}
	if len(p.Records) == 0 {
		return ErrEmptyProposalRecords
	}

	for _, record := range p.Records {
		if err := record.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}

// String returns a string containing the pool incentives proposal.
func (p UpdateLockdropProposal) String() string {
	// TODO: Make this prettier
	recordsStr := ""
	for _, record := range p.Records {
		recordsStr = recordsStr + fmt.Sprintf("(GaugeId: %d, Weight: %s) ", record.GaugeId, record.Weight.String())
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Update Pool Incentives Proposal:
  Title:       %s
  Description: %s
  Records:     %s
`, p.Title, p.Description, recordsStr))
	return b.String()
}
