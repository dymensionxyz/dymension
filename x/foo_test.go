package x

import (
	"flag"
	"testing"

	"pgregory.net/rapid"
)

func TesWiz(t *testing.T) {
	_ = flag.Set("rapid.checks", "50")
	_ = flag.Set("rapid.steps", "50")
	rapid.Check(t, func(t *rapid.T) {
		t.Repeat(map[string]func(*rapid.T){
			"": func(t *rapid.T) {
			},
			"foo": func(t *rapid.T) {
			},
		})
	})
}

type Model struct {
	kasUTXO []int
}

func (m *Model) deposit(x int) {
	m.kasUTXO = append(m.kasUTXO, x)
}

func (m *Model) withdraw(x int) {
}
