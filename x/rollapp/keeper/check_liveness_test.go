package keeper_test

import (
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
)

func TestLivenessSlashAndJail1(t *testing.T) {
	tests := []struct {
		name         string
		args         keeper.LivenessSlashAndJailArgs
		wantSlashAmt types.Coins
		wantJail     bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotSlashAmt, gotJail := keeper.LivenessSlashAndJail(tt.args)
			if !reflect.DeepEqual(gotSlashAmt, tt.wantSlashAmt) {
				t.Errorf("LivenessSlashAndJail() gotSlashAmt = %v, want %v", gotSlashAmt, tt.wantSlashAmt)
			}
			if gotJail != tt.wantJail {
				t.Errorf("LivenessSlashAndJail() gotJail = %v, want %v", gotJail, tt.wantJail)
			}
		})
	}
}
