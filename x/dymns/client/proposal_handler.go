package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/dymensionxyz/dymension/v3/x/dymns/client/cli"
)

var (
	MigrateChainIdsProposalHandler = govclient.NewProposalHandler(cli.NewMigrateChainIdsCmd)
	UpdateAliasesProposalHandler   = govclient.NewProposalHandler(cli.NewUpdateAliasesCmd)
)
