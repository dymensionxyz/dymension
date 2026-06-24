package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
)

// The agent module exposes only queries (registry + attested action log).
// Messages (registration, submission) live in their own issues; there is
// nothing to register here yet.
func RegisterCodec(*codec.LegacyAmino) {}

func RegisterInterfaces(cdctypes.InterfaceRegistry) {}
