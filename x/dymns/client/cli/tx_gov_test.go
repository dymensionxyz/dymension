package cli

import (
	"testing"

	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/evmos/ethermint/x/evm/types"
	"github.com/stretchr/testify/require"
)

func Test_parseMigrateChainIdsProposal(t *testing.T) {
	//goland:noinspection SpellCheckingInspection
	testCases := []struct {
		name         string
		metadataFile string
		wantErr      bool
		want         []dymnstypes.MigrateChainId
	}{
		{
			name:         "fail - invalid file name",
			metadataFile: "",
			wantErr:      true,
		},
		{
			name:         "fail - invalid content",
			metadataFile: "test_proposals/mcid_invalid_update_chain_id_proposal_test.json",
			wantErr:      true,
		},
		{
			name:         "pass - update single",
			metadataFile: "test_proposals/mcid_update_single_chain_id_proposal_test.json",
			wantErr:      false,
			want: []dymnstypes.MigrateChainId{
				{
					PreviousChainId: "cosmoshub-3",
					NewChainId:      "cosmoshub-4",
				},
			},
		},
		{
			name:         "pass - update multiple",
			metadataFile: "test_proposals/mcid_update_multiple_chain_ids_proposal_test.json",
			wantErr:      false,
			want: []dymnstypes.MigrateChainId{
				{
					PreviousChainId: "cosmoshub-3",
					NewChainId:      "cosmoshub-4",
				},
				{
					PreviousChainId: "columbus-4",
					NewChainId:      "columbus-5",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			proposal, err := parseMigrateChainIdsProposal(types.AminoCdc, tc.metadataFile)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Len(t, proposal.Replacement, len(tc.want))
			require.Equal(t, tc.want, proposal.Replacement)
		})
	}
}

func Test_parseUpdateAliasesProposal(t *testing.T) {
	testCases := []struct {
		name         string
		metadataFile string
		wantErr      bool
		wantAdd      []dymnstypes.UpdateAlias
		wantRemove   []dymnstypes.UpdateAlias
	}{
		{
			name:         "fail - invalid file name",
			metadataFile: "",
			wantErr:      true,
		},
		{
			name:         "fail - invalid content",
			metadataFile: "test_proposals/uac_invalid_update_aliases_proposal_test.json",
			wantErr:      true,
		},
		{
			name:         "pass - update single",
			metadataFile: "test_proposals/uac_update_aliases_single_add_proposal_test.json",
			wantErr:      false,
			wantAdd: []dymnstypes.UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym",
				},
			},
		},
		{
			name:         "pass - update multiple",
			metadataFile: "test_proposals/uac_update_aliases_multiple_add_remove_proposal_test.json",
			wantErr:      false,
			wantAdd: []dymnstypes.UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dym",
				},
				{
					ChainId: "blumbus_111-1",
					Alias:   "bb",
				},
			},
			wantRemove: []dymnstypes.UpdateAlias{
				{
					ChainId: "dymension_1100-1",
					Alias:   "dymension",
				},
				{
					ChainId: "froopyland_111-1",
					Alias:   "fl",
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			proposal, err := parseUpdateAliasesProposal(types.AminoCdc, tc.metadataFile)
			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Len(t, proposal.Add, len(tc.wantAdd))
			if len(tc.wantAdd) > 0 {
				require.Equal(t, tc.wantAdd, proposal.Add)
			}

			require.Len(t, proposal.Remove, len(tc.wantRemove))
			if len(tc.wantRemove) > 0 {
				require.Equal(t, tc.wantRemove, proposal.Remove)
			}
		})
	}
}
