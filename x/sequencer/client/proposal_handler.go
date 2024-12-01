package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/client/cli"
)

var PunishSequencerHandler = govclient.NewProposalHandler(cli.NewCmdSubmitPunishSequencerProposal)
