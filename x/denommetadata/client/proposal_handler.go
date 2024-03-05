package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/client/cli"
)

var (
	CreateDenomMetadataHandler = govclient.NewProposalHandler(cli.NewCmdSubmitCreateDenomMetadataProposal)
)
