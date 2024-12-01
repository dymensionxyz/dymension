package types

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

const (
	// ProposalTypePunishSequencer defines the type for a PunishSequencerProposal
	ProposalTypePunishSequencer = "PunishSequencer"
)

// Assert PunishSequencerProposal implements govtypes.Content at compile-time
var (
	_ govtypes.Content = &PunishSequencerProposal{}
)

func init() {
	govtypes.RegisterProposalType(ProposalTypePunishSequencer)
}

// NewPunishSequencerProposal creates a new punish sequencer proposal.
func NewPunishSequencerProposal(
	title, description, punishSequencerAddress, rewardee string,
) *PunishSequencerProposal {
	return &PunishSequencerProposal{
		Title:                  title,
		Description:            description,
		PunishSequencerAddress: punishSequencerAddress,
		Rewardee:               rewardee,
	}
}

// GetTitle returns the title of a community pool spend proposal.
func (csp *PunishSequencerProposal) GetTitle() string { return csp.Title }

// GetDescription returns the description of a community pool spend proposal.
func (csp *PunishSequencerProposal) GetDescription() string { return csp.Description }

// GetDescription returns the routing key of a community pool spend proposal.
func (csp *PunishSequencerProposal) ProposalRoute() string { return RouterKey }

// ProposalType returns the type of a community pool spend proposal.
func (csp *PunishSequencerProposal) ProposalType() string { return ProposalTypePunishSequencer }

// ValidateBasic runs basic stateless validity checks
func (csp *PunishSequencerProposal) ValidateBasic() error {
	err := govtypes.ValidateAbstract(csp)
	if err != nil {
		return err
	}

	if len(csp.PunishSequencerAddress) == 0 {
		return fmt.Errorf("punish sequencer address cannot be empty")
	}

	if len(csp.Rewardee) == 0 {
		return fmt.Errorf("rewardee address cannot be empty")
	}
	return nil
}

// String implements the Stringer interface.
func (csp PunishSequencerProposal) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf(`Punish Sequencer Proposal:
	  Title:       %s
	  Description: %s
	  PunishSequencerAddress: %s
	  Rewardee: %s
`, csp.Title, csp.Description, csp.PunishSequencerAddress, csp.Rewardee))
	return b.String()
}

// MustRewardee returns acc address if rewardee field is not empty
func (csp PunishSequencerProposal) MustRewardee() *sdk.AccAddress {
	if csp.Rewardee == "" {
		return nil
	}
	rewardee, _ := sdk.AccAddressFromBech32(csp.Rewardee)
	return &rewardee
}
