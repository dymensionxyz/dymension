package types

// Incentive module event types.
const (
	// EventTypeCreateSequencer is emitted when a sequencer is created
	EventTypeCreateSequencer = "create_sequencer"

	// EventTypeNoBondedSequencer is emitted when no bonded sequencer is found for a rollapp
	EventTypeNoBondedSequencer = "no_bonded_sequencer"

	// EventTypeRotationStarted is emitted when a rotation is started (after notice period)
	EventTypeRotationStarted = "rotation_started"

	// EventTypeProposerRotated is emitted when a proposer is rotated
	EventTypeProposerRotated = "proposer_rotated"

	// EventTypeNoticePeriodStarted is emitted when a sequencer's notice period starts
	EventTypeNoticePeriodStarted = "notice_period_started"

	// EventTypeUnbonding is emitted when a sequencer is unbonding
	EventTypeUnbonding = "unbonding"

	// EventTypeUnbonded is emitted when a sequencer is unbonded
	EventTypeUnbonded = "unbonded"

	// EventTypeSlashed is emitted when a sequencer is slashed
	EventTypeSlashed = "slashed"

	AttributeKeyRollappId      = "rollapp_id"
	AttributeKeySequencer      = "sequencer"
	AttributeKeyBond           = "bond"
	AttributeKeyProposer       = "proposer"
	AttributeKeyNextProposer   = "next_proposer"
	AttributeKeyCompletionTime = "completion_time"
)
