package types

// Incentive module event types.
const (
	// EventTypeCreateSequencer is emitted when a sequencer is created
	// Attributes:
	// - AttributeKeyRollappId
	// - AttributeKeySequencer
	// - AttributeKeyBond
	// - AttributeKeyProposer
	EventTypeCreateSequencer = "create_sequencer"

	// EventTypeRotationStarted is emitted when a rotation is started (after notice period)
	// Attributes:
	// - AttributeKeyRollappId
	// - AttributeKeyNextProposer
	// - AttributeKeyRewardAddr
	// - AttributeKeyWhitelistedRelayers
	EventTypeRotationStarted = "proposer_rotation_started"

	// EventTypeProposerRotated is emitted when a proposer is rotated
	// Attributes:
	// - AttributeKeyRollappId
	// - AttributeKeySequencer
	EventTypeProposerRotated = "proposer_rotated"

	// EventTypeNoticePeriodStarted is emitted when a sequencer's notice period starts
	// Attributes:
	// - AttributeKeyRollappId
	// - AttributeKeySequencer
	// - AttributeNextProposer
	// - AttributeKeyCompletionTime
	EventTypeNoticePeriodStarted = "notice_period_started"

	// EventTypeUnbonding is emitted when a sequencer is unbonding
	EventTypeUnbonding = "unbonding"

	// EventTypeUnbonded is emitted when a sequencer is unbonded
	EventTypeUnbonded = "unbonded"

	// EventTypeSlashed is emitted when a sequencer is slashed
	EventTypeSlashed = "slashed"
	// EventTypeJailed is emitted when a sequencer is jailed
	EventTypeJailed = "jailed"
	// EventTypeBondIncreased is emitted when a sequencer's bond is increased
	EventTypeBondIncreased = "bond_increased"

	AttributeKeyRollappId           = "rollapp_id"
	AttributeKeySequencer           = "sequencer"
	AttributeKeyBond                = "bond"
	AttributeKeyProposer            = "proposer"
	AttributeKeyNextProposer        = "next_proposer"
	AttributeKeyRewardAddr          = "reward_addr"
	AttributeKeyWhitelistedRelayers = "whitelisted_relayers"
	AttributeKeyCompletionTime      = "completion_time"
)
