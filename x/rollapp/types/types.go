package types

import (
	common "github.com/dymensionxyz/dymension/v3/x/common/types"
)

type StateStatus common.Status

type GenParams struct {
	ChannelID       string
	RollappID       string
	TokenMetadata   []*TokenMetadata
	GenesisAccounts []*GenesisAccount
}
