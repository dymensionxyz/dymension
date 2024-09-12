package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group sequencer queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdListSequencer())
	cmd.AddCommand(CmdShowSequencer())
	cmd.AddCommand(CmdShowSequencersByRollapp())
	cmd.AddCommand(CmdGetProposerByRollapp())
	cmd.AddCommand(CmdGetNextProposerByRollapp())
	cmd.AddCommand(CmdGetAllProposers())

	return cmd
}
