package types

func (seq Sequencer) IsBonded() bool {
	if seq.Status != Bonded {
		return false
	}
	return true
}

// is proposer
func (seq Sequencer) IsProposer() bool {
	return seq.Proposer
}
