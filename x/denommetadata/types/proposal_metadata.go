package types

import (
	"fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	// ProposalTypeCreateStream defines the type for a CreateStreamProposal
	ProposalTypeCreateDenomMetadata = "CreateDenomMetadata"

	// ProposalTypeTerminateStream defines the type for a UpdateDenomMetadata
	ProposalTypeUpdateDenomMetadata = "UpdateDenomMetadata"
)

// Assert CreateStreamProposal implements govtypes.Content at compile-time
var _ govtypes.Content = &CreateDenomMetadataProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeCreateDenomMetadata)

}

// NewCreateMetadataProposal creates a new create stream proposal.
func NewCreateMetadataProposal(title, description string, denommetadata *TokenMetadata) *CreateDenomMetadataProposal {
	return &CreateDenomMetadataProposal{
		Title:         title,
		Description:   description,
		TokenMetadata: *denommetadata,
	}
}

// GetTitle returns the title of a community pool spend proposal.
func (csp *CreateDenomMetadataProposal) GetTitle() string { return csp.Title }

// GetDescription returns the description of a community pool spend proposal.
func (csp *CreateDenomMetadataProposal) GetDescription() string { return csp.Description }

// GetDescription returns the routing key of a community pool spend proposal.
func (csp *CreateDenomMetadataProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (csp *CreateDenomMetadataProposal) ProposalType() string { return ProposalTypeCreateDenomMetadata }

// ValidateBasic runs basic stateless validity checks
func (csp *CreateDenomMetadataProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(csp)
	if err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (csp CreateDenomMetadataProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Create stream Proposal:
	  Title:       %s
	  Description: %s
	  TokenMetadata: %s
`, csp.Title, csp.Description, &csp.TokenMetadata))
	return b.String()
}
