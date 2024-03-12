package types

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	// ProposalTypeCreateDenomMetadata defines the type for a CreateDenomMetadata
	ProposalTypeCreateDenomMetadata = "CreateDenomMetadata"

	// ProposalTypeUpdateDenomMetadata defines the type for a UpdateDenomMetadata
	ProposalTypeUpdateDenomMetadata = "UpdateDenomMetadata"
)

// Assert CreateDenomMetadataProposal implements govtypes.Content at compile-time
var _ govtypes.Content = &CreateDenomMetadataProposal{}
var _ govtypes.Content = &UpdateDenomMetadataProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeCreateDenomMetadata)
	govtypes.RegisterProposalType(ProposalTypeUpdateDenomMetadata)

}

// NewCreateMetadataProposal creates a new create denommetadata proposal.
func NewCreateMetadataProposal(title, description string, denommetadata types.Metadata) *CreateDenomMetadataProposal {
	return &CreateDenomMetadataProposal{
		Title:         title,
		Description:   description,
		TokenMetadata: denommetadata,
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

	err = csp.TokenMetadata.Validate()
	if err != nil {
		return err
	}
	return nil
}

// String implements the Stringer interface.
func (csp CreateDenomMetadataProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Create denommetadata Proposal:
	  Title:       %s
	  Description: %s
	  TokenMetadata: %s
`, csp.Title, csp.Description, &csp.TokenMetadata))
	return b.String()
}

// NewUpdateDenomMetadataProposal creates a new proposal for updating existing token metadata.
func NewUpdateDenomMetadataProposal(title, description string, denommetadata types.Metadata) *UpdateDenomMetadataProposal {
	return &UpdateDenomMetadataProposal{
		Title:         title,
		Description:   description,
		TokenMetadata: denommetadata,
	}
}

// GetTitle returns the title of a community pool spend proposal.
func (csp *UpdateDenomMetadataProposal) GetTitle() string { return csp.Title }

// GetDescription returns the description of a community pool spend proposal.
func (csp *UpdateDenomMetadataProposal) GetDescription() string { return csp.Description }

// GetDescription returns the routing key of a community pool spend proposal.
func (csp *UpdateDenomMetadataProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (csp *UpdateDenomMetadataProposal) ProposalType() string { return ProposalTypeCreateDenomMetadata }

// ValidateBasic runs basic stateless validity checks
func (csp *UpdateDenomMetadataProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(csp)
	if err != nil {
		return err
	}
	err = csp.TokenMetadata.Validate()
	if err != nil {
		return err
	}
	return nil
}

// String implements the Stringer interface.
func (csp UpdateDenomMetadataProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Update denommetadata Proposal:
	  Title:       %s
	  Description: %s
	  TokenMetadata: %s
`, csp.Title, csp.Description, &csp.TokenMetadata))
	return b.String()
}
