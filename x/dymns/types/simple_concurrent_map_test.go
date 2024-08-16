package types

import (
	"testing"

	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestNewSimpleConcurrentMap(t *testing.T) {
	_ = flag.Set("rapid.checks", "1000")
	_ = flag.Set("rapid.steps", "1000")

	keys := rapid.StringMatching(`^\d{1,2}$`)
	values := rapid.Int64Range(10, 20)
	rapid.Check(t, func(r *rapid.T) {
		m := NewSimpleConcurrentMap[string, int64]()
		model := make(map[string]int64) // model actual storage
		r.Repeat(map[string]func(r *rapid.T){
			"set": func(r *rapid.T) {
				k := keys.Draw(r, "k")
				v := values.Draw(r, "v")
				m.Set(k, v)
				model[k] = v
			},
			"get": func(r *rapid.T) {
				k := keys.Draw(r, "k")

				v1, f1 := m.Get(k)
				v2, f2 := model[k]
				require.Equal(r, v1, v2, "value from store does not match model")
				require.Equal(r, f1, f2, "found from store does not match model")
			},
			"has": func(r *rapid.T) {
				k := keys.Draw(r, "k")

				f1 := m.Has(k)
				_, f2 := model[k]
				require.Equal(r, f1, f2, "found from store does not match model")
			},
			"delete": func(r *rapid.T) {
				k := keys.Draw(r, "k")
				m.Delete(k)
				delete(model, k)
			},
			"clear": func(r *rapid.T) {
				m.Clear()
				clear(model)
			},
		})
	})
}
