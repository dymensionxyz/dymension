package types

const (
	EventTypeStateUpdate  = "state_update"
	EventTypeStatusChange = "status_change"

	AttributeKeyRollappId      = "rollapp_id"
	AttributeRollappIBCdenom   = "ibc_denom"
	AttributeKeyStateInfoIndex = "state_info_index"
	AttributeKeyStartHeight    = "start_height"
	AttributeKeyNumBlocks      = "num_blocks"
	AttributeKeyDAPath         = "da_path"
	AttributeKeyStatus         = "status"

	// EventTypeFraud is emitted when a fraud evidence is submitted
	EventTypeFraud             = "fraud_proposal"
	AttributeKeyFraudHeight    = "fraud_height"
	AttributeKeyFraudSequencer = "fraud_sequencer"
	AttributeKeyClientID       = "client_id"

	// EventTypeTransfersEnabled is when the bridge is enabled
	EventTypeTransfersEnabled = "transfers_enabled"
)
