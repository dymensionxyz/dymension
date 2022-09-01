package types

type SimSequencer struct {
	// sequencerAddress is the bech32-encoded address of the sequencer account.
	SequencerAddress string
	// creator is the bech32-encoded address of the account sent the transaction (sequencer creator)
	Creator string
	// rollappId defines the rollapp to which the sequencer belongs.
	RollappId string `protobuf:"bytes,4,opt,name=rollappId,proto3" json:"rollappId,omitempty"`
}
