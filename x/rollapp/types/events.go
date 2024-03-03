package types

const (
	EventTypeStateUpdate  = "state_update"
	EventTypeStatusChange = "status_change"

	AttributeKeyRollappId      = "rollapp_id"
	AttributeKeyStateInfoIndex = "state_info_index"
	AttributeKeyStartHeight    = "start_height"
	AttributeKeyNumBlocks      = "num_blocks"
	AttributeKeyDAPath         = "da_path"
	AttributeKeyStatus         = "status"

	EventTypeFraud             = "fraud_proposal"
	AttributeKeyFraudHeight    = "fraud_height"
	AttributeKeyFraudSequencer = "fraud_sequencer"
)
