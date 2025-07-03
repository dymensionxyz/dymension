package types

import (
	"fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	// ProposalTypeReplaceStreamDistribution defines the type for a ReplaceStreamDistributionmProposal
	ProposalTypeReplaceStreamDistribution = "ReplaceStreamDistribution"

	// ProposalTypeUpdateStreamDistribution defines the type for a UpdateStreamDistributionmProposal
	ProposalTypeUpdateStreamDistribution = "UpdateStreamDistribution"
)

// Assert ReplaceStreamDistributionProposal implements govtypes.Content at compile-time
var (
	_ govtypes.Content = &ReplaceStreamDistributionProposal{}
	_ govtypes.Content = &UpdateStreamDistributionProposal{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypeReplaceStreamDistribution)
	govtypes.RegisterProposalType(ProposalTypeUpdateStreamDistribution)
}

// NewReplaceStreamDistributionProposal creates a new ReplaceStreamDistribution proposal.
//
//nolint:interfacer
func NewReplaceStreamDistributionProposal(title, description string, streamId uint64, distrToRecords []DistrRecord) *ReplaceStreamDistributionProposal {
	return &ReplaceStreamDistributionProposal{
		Title:       title,
		Description: description,
		StreamId:    streamId,
		Records:     distrToRecords,
	}
}

// GetTitle returns the title of a community pool spend proposal.
func (csp *ReplaceStreamDistributionProposal) GetTitle() string { return csp.Title }

// GetDescription returns the description of a community pool spend proposal.
func (csp *ReplaceStreamDistributionProposal) GetDescription() string { return csp.Description }

// GetDescription returns the routing key of a community pool spend proposal.
func (csp *ReplaceStreamDistributionProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (csp *ReplaceStreamDistributionProposal) ProposalType() string {
	return ProposalTypeReplaceStreamDistribution
}

// ValidateBasic runs basic stateless validity checks
func (csp *ReplaceStreamDistributionProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(csp)
	if err != nil {
		return err
	}

	for _, record := range csp.Records {
		err := record.ValidateBasic()
		if err != nil {
			return err
		}
	}

	return nil
}

// String implements the Stringer interface.
func (csp ReplaceStreamDistributionProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Create stream Proposal:
	  Title:       %s
	  Description: %s
	  StreamID:    %d
	  Records:     %v
`, csp.Title, csp.Description, csp.StreamId, csp.Records))
	return b.String()
}

// NewUpdateStreamDistributionProposal creates a new UpdateStreamDistribution proposal.
//
//nolint:interfacer
func NewUpdateStreamDistributionProposal(title, description string, streamId uint64, distrToRecords []DistrRecord) *UpdateStreamDistributionProposal {
	return &UpdateStreamDistributionProposal{
		Title:       title,
		Description: description,
		StreamId:    streamId,
		Records:     distrToRecords,
	}
}

// GetTitle returns the title of a community pool spend proposal.
func (csp *UpdateStreamDistributionProposal) GetTitle() string { return csp.Title }

// GetDescription returns the description of a community pool spend proposal.
func (csp *UpdateStreamDistributionProposal) GetDescription() string { return csp.Description }

// GetDescription returns the routing key of a community pool spend proposal.
func (csp *UpdateStreamDistributionProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (csp *UpdateStreamDistributionProposal) ProposalType() string {
	return ProposalTypeReplaceStreamDistribution
}

// ValidateBasic runs basic stateless validity checks
func (csp *UpdateStreamDistributionProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(csp)
	if err != nil {
		return err
	}

	for _, record := range csp.Records {
		err := record.ValidateBasic()
		if err != nil {
			return err
		}
	}

	return nil
}

// String implements the Stringer interface.
func (csp UpdateStreamDistributionProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Create stream Proposal:
	  Title:       %s
	  Description: %s
	  StreamID:    %d
	  Records:     %v
`, csp.Title, csp.Description, csp.StreamId, csp.Records))
	return b.String()
}
