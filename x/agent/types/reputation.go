package types

// Add counts fb's score into the aggregate, including the attested bucket
// when fb came from an attested client.
func (r *Reputation) Add(fb Feedback) {
	r.Count++
	r.ScoreSum += uint64(fb.Score)
	if fb.FromAttestedClient {
		r.AttestedCount++
		r.AttestedScoreSum += uint64(fb.Score)
	}
}

// Remove undoes a prior Add of fb. Callers must only remove feedback that is
// currently counted, so the subtractions cannot underflow.
func (r *Reputation) Remove(fb Feedback) {
	r.Count--
	r.ScoreSum -= uint64(fb.Score)
	if fb.FromAttestedClient {
		r.AttestedCount--
		r.AttestedScoreSum -= uint64(fb.Score)
	}
}
