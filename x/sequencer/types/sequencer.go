package types

func (seq Sequencer) IsBonded() bool {
	if seq.Status != Bonded && seq.Status != Proposer {
		return false
	}
	return true
}
