package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/client/cli"
)

var SubmitFraudHandler = govclient.NewProposalHandler(cli.NewCmdSubmitFraudProposal)
