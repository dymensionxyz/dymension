package client

import (
	"github.com/dymensionxyz/dymension/x/lockdrop/client/cli"

	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
)

var (
	UpdateLockdropHandler  = govclient.NewProposalHandler(cli.NewCmdSubmitUpdateLockdropProposal)
	ReplaceLockdropHandler = govclient.NewProposalHandler(cli.NewCmdSubmitReplaceLockdropProposal)
)
