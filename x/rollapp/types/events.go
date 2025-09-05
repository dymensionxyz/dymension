package types

const (
	EventTypeStateUpdate  = "state_update"
	EventTypeStatusChange = "status_change"

	AttributeKeyRollappId          = "rollapp_id"
	AttributeRollappIBCdenom       = "ibc_denom"
	AttributeKeyStateInfoIndex     = "state_info_index"
	AttributeKeyStartHeight        = "start_height"
	AttributeKeyNumBlocks          = "num_blocks"
	AttributeKeyDAPath             = "da_path"
	AttributeKeyStatus             = "status"
	AttributeKeyCanonicalChannelId = "canonical_channel_id"

	// EventTypeHardFork is emitted when a fraud evidence is submitted
	EventTypeHardFork             = "hard_fork"
	AttributeKeyNewRevisionHeight = "new_revision_height"
	AttributeKeyClientID          = "client_id"

	// EventTypeTransfersEnabled is when the bridge is enabled
	EventTypeTransfersEnabled = "transfers_enabled"
	
	// EventTypeTEEFastFinalization is emitted when TEE attestation fast finalizes states
	EventTypeTEEFastFinalization = "tee_fast_finalization"
	EventTypeStateFinalized       = "state_finalized"
	AttributeKeyStateIndex        = "state_index"
	AttributeKeyStatesFinalized   = "states_finalized"
)
