package lightclient

import (
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	tmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
)

const (
	Type = "drs"
	Foo  = "drs"
)

var _ exported.ClientState = &ClientState{}

type ClientState struct {
	*tmtypes.ClientState
}
