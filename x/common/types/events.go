package types

// RollappPacket attributes
const (
	AttributeKeyRollappId                = "rollapp_id"
	AttributeKeyPacketStatus             = "status"
	AttributeKeyPacketSourcePort         = "source_port"
	AttributeKeyPacketSourceChannel      = "source_channel"
	AttributeKeyPacketDestinationPort    = "destination_port"
	AttributeKeyPacketDestinationChannel = "destination_channel"
	AttributeKeyPacketSequence           = "packet_sequence"
	AttributeKeyPacketProofHeight        = "proof_height"
	AttributeKeyPacketType               = "type"
	AttributeKeyPacketAcknowledgement    = "acknowledgement"
	AttributeKeyPacketError              = "error"
)

// FungibleTokenPacketData attributes
const (
	AttributeKeyPacketDataDenom    = "packet_data_denom"
	AttributeKeyPacketDataAmount   = "packet_data_amount"
	AttributeKeyPacketDataSender   = "packet_data_sender"
	AttributeKeyPacketDataReceiver = "packet_data_receiver"
	AttributeKeyPacketDataMemo     = "packet_data_memo"
)
