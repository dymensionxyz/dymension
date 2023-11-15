package types

import (
	"fmt"
	"strings"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	// ProposalTypeCreateStream defines the type for a CreateStreamProposal
	ProposalTypeCreateStream = "CreateStream"
)

// Assert CreateStreamProposal implements govtypes.Content at compile-time
var _ govtypes.Content = &CreateStreamProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeCreateStream)
}

// NewCreateStreamProposal creates a new community pool spend proposal.
//
//nolint:interfacer
func NewCreateStreamProposal(title, description string, stream Stream) *CreateStreamProposal {
	return &CreateStreamProposal{
		Title:                title,
		Description:          description,
		DistributeToRecords:  stream.DistributeTo.Records,
		Coins:                stream.Coins,
		StartTime:            stream.StartTime,
		DistrEpochIdentifier: stream.DistrEpochIdentifier,
		NumEpochsPaidOver:    stream.NumEpochsPaidOver,
	}
}

// GetTitle returns the title of a community pool spend proposal.
func (csp *CreateStreamProposal) GetTitle() string { return csp.Title }

// GetDescription returns the description of a community pool spend proposal.
func (csp *CreateStreamProposal) GetDescription() string { return csp.Description }

// GetDescription returns the routing key of a community pool spend proposal.
func (csp *CreateStreamProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (csp *CreateStreamProposal) ProposalType() string { return ProposalTypeCreateStream }

// ValidateBasic runs basic stateless validity checks
func (csp *CreateStreamProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(csp)
	if err != nil {
		return err
	}

	for _, record := range csp.DistributeToRecords {
		err := record.ValidateBasic()
		if err != nil {
			return err
		}
	}

	if !csp.Coins.IsAllPositive() {
		return fmt.Errorf("all coins %s must be positive", csp.Coins)
	}

	if csp.NumEpochsPaidOver <= 0 {
		return fmt.Errorf("numEpochsPaidOver must be greater than 0")
	}
	return nil
}

// String implements the Stringer interface.
func (csp CreateStreamProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Create stream Proposal:
	  Title:       %s
	  Description: %s
	  DistributeTo: %v
	  Coins:       %s
	  StartTime:   %s
	  EpochIdentifier:   %s
	  NumEpochsPaidOver:   %d
`, csp.Title, csp.Description, &csp.DistributeToRecords, csp.Coins, csp.StartTime, csp.DistrEpochIdentifier, csp.NumEpochsPaidOver))
	return b.String()
}
