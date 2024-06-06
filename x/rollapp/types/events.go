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

	// EventTypeTransferGenesisH1 is when all genesis transfers from the rollapp have been received
	EventTypeTransferGenesisH1 = "transfer_genesis_h1"
	// EventTypeTransferGenesisH2 is when the dispute period has elapsed after all genesis transfers from the rollapp have been received
	EventTypeTransferGenesisH2           = "transfer_genesis_h2"
	AttributeKeyTransferGenesisNReceived = "num_transfers_received"
)
