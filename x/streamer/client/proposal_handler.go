package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	"github.com/dymensionxyz/dymension/x/streamer/client/cli"
)

var (
	CreateStreamHandler  = govclient.NewProposalHandler(cli.NewCmdSubmitCreateStreamProposal)
	StopStreamHandler    = govclient.NewProposalHandler(cli.NewCmdSubmitStopStreamProposal)
	ReplaceStreamHandler = govclient.NewProposalHandler(cli.NewCmdSubmitReplaceStreamDistributionProposal)
)
