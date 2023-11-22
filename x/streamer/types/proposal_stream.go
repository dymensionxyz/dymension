package types

import (
	"fmt"
	"strings"
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	// ProposalTypeCreateStream defines the type for a CreateStreamProposal
	ProposalTypeCreateStream = "CreateStream"

	// ProposalTypeStopStream defines the type for a StopStreamProposal
	ProposalTypeStopStream = "StopStream"
)

// Assert CreateStreamProposal implements govtypes.Content at compile-time
var _ govtypes.Content = &CreateStreamProposal{}
var _ govtypes.Content = &StopStreamProposal{}

func init() {
	govtypes.RegisterProposalType(ProposalTypeCreateStream)
	govtypes.RegisterProposalType(ProposalTypeStopStream)

}

// NewCreateStreamProposal creates a new create stream proposal.
//
//nolint:interfacer
func NewCreateStreamProposal(title, description string, coins sdk.Coins, distrToRecords []DistrRecord, startTime time.Time, epochIdentifier string, numEpochsPaidOver uint64) *CreateStreamProposal {
	return &CreateStreamProposal{
		Title:                title,
		Description:          description,
		DistributeToRecords:  distrToRecords,
		Coins:                coins,
		StartTime:            startTime,
		DistrEpochIdentifier: epochIdentifier,
		NumEpochsPaidOver:    numEpochsPaidOver,
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

// NewStopStreamProposal creates a new stop stream proposal.
//
//nolint:interfacer
func NewStopStreamProposal(title, description string, streamId uint64) *StopStreamProposal {
	return &StopStreamProposal{
		Title:       title,
		Description: description,
		StreamId:    streamId,
	}
}

// GetTitle returns the title of a community pool spend proposal.
func (csp *StopStreamProposal) GetTitle() string { return csp.Title }

// GetDescription returns the description of a community pool spend proposal.
func (csp *StopStreamProposal) GetDescription() string { return csp.Description }

// GetDescription returns the routing key of a community pool spend proposal.
func (csp *StopStreamProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (csp *StopStreamProposal) ProposalType() string { return ProposalTypeCreateStream }

// ValidateBasic runs basic stateless validity checks
func (csp *StopStreamProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(csp)
	if err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (csp StopStreamProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Create stream Proposal:
	  Title:       %s
	  Description: %s
	  StreamID:    %d
`, csp.Title, csp.Description, &csp.StreamId))
	return b.String()
}
