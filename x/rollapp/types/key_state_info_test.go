package types

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/dymensionxyz/sdk-utils/utils/urand"
)

// nolint: gosec
func TestStateInfoIndexFromKey(t *testing.T) {
	index := StateInfoIndex{
		RollappId: urand.RollappID(),
		Index:     rand.Uint64(),
	}

	type args struct {
		key []byte
	}

	tests := []struct {
		name string
		args args
		want StateInfoIndex
	}{
		{
			name: "Test StateInfoIndexFromKey",
			args: args{
				key: StateInfoKey(index),
			},
			want: index,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StateInfoIndexFromKey(tt.args.key); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StateInfoIndexFromKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
