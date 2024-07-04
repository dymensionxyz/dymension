package rotation

import (
	"testing"

	"pgregory.net/rapid"
)

const (
	seq1 = "seq1"
	seq2 = "seq2"
)

type state struct {
	h       uint64
	signer  string
	nextSeq string
}

type model struct {
	states           map[uint64]state
	ack              uint64
	sentChangePacket bool
}

func (m *model) SubmitState(s state) {
}

func (m *model) SubmitAck(h uint64) {
	if !m.sentChangePacket || ack != 0 {
		return
	}
}

func (m *model) SubmitFraud() {
}

func (m *model) ChangeSequencer() {
	m.sentChangePacket = true
}

func TestRotationRapid(t *testing.T) {
	m := model{}
	m.states = make(map[uint64]state)
	m.states[0] = state{
		h:       0,
		signer:  seq1,
		nextSeq: seq1,
	}

	seq := rapid.SampledFrom([]string{seq1, seq2})
	height := rapid.IntRange(1, 100)

	f := func(r *rapid.T) {
		ops := map[string]func(*rapid.T){
			"submit state": func(t *rapid.T) {
			},
			"submit ack": func(t *rapid.T) {
			},
			"submit fraud": func(t *rapid.T) {
			},
			"change sequencer": func(t *rapid.T) {
			},
		}
		r.Repeat(ops)
	}

	rapid.Check(t, f)
}
