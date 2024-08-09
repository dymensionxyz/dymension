package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/dymensionxyz/dymension/v3/x/dymns/client/cli"
)

var (
	// MigrateChainIdsProposalHandler is the proposal handler for migrating chain-ids
	// in module params and configurations of non-expired Dym-Names.
	MigrateChainIdsProposalHandler = govclient.NewProposalHandler(cli.NewMigrateChainIdsCmd)
	// UpdateAliasesProposalHandler is the proposal handler for updating aliases of chain-ids.
	UpdateAliasesProposalHandler = govclient.NewProposalHandler(cli.NewUpdateAliasesCmd)
)
