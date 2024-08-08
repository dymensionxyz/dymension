package types

// Incentive module event types.
const (
	// EventTypeCreateSequencer is emitted when a sequencer is created
	EventTypeCreateSequencer = "create_sequencer"
	AttributeKeyRollappId    = "rollapp_id"
	AttributeKeySequencer    = "sequencer"
	AttributeKeyBond         = "bond"
	AttributeKeyProposer     = "proposer"

	// EventTypeUnbonding is emitted when a sequencer is unbonding
	EventTypeUnbonding         = "unbonding"
	AttributeKeyCompletionTime = "completion_time"

	// EventTypeNoBondedSequencer is emitted when no bonded sequencer is found for a rollapp
	EventTypeNoBondedSequencer = "no_bonded_sequencer"

	// EventTypeProposerRotated is emitted when a proposer is rotated
	EventTypeProposerRotated = "proposer_rotated"

	// EventTypeUnbonded is emitted when a sequencer is unbonded
	EventTypeUnbonded = "unbonded"

	// EventTypeSlashed is emitted when a sequencer is slashed
	EventTypeSlashed = "slashed"
	// EventTypeJailed is emitted when a sequencer is jailed
	EventTypeJailed = "jailed"
	// EventTypeBondIncreased is emitted when a sequencer's bond is increased
	EventTypeBondIncreased = "bond_increased"
)
