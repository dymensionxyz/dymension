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

	// EventTypeHardFork is emitted when a fraud evidence is submitted
	EventTypeHardFork             = "hard_fork"
	AttributeKeyNewRevisionHeight = "new_revision_height"
	AttributeKeyClientID          = "client_id"

	// EventTypeTransfersEnabled is when the bridge is enabled
	EventTypeTransfersEnabled = "transfers_enabled"
)
