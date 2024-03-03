package types

func (seq Sequencer) IsBonded() bool {
	return seq.Status == Bonded
}

// is proposer
func (seq Sequencer) IsProposer() bool {
	return seq.Proposer
}
