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

	// EventTypeFraud is emitted when a fraud evidence is submitted
	EventTypeFraud             = "fraud_proposal"
	AttributeKeyFraudHeight    = "fraud_height"
	AttributeKeyFraudSequencer = "fraud_sequencer"
	AttributeKeyClientID       = "client_id"

	// EventTypeTransferGenesisAllReceived is when all genesis transfers from the rollapp have been received
	EventTypeTransferGenesisAllReceived = "transfer_genesis_all_received"
	// EventTypeTransferGenesisTransfersEnabled is when the bridge is enabled
	EventTypeTransferGenesisTransfersEnabled = "transfer_genesis_transfers_enabled"
	AttributeKeyTransferGenesisNReceived     = "num_transfers_received"
)
