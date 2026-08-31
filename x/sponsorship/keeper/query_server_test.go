package keeper_test

import (
	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func (s *KeeperTestSuite) TestQueryMissingEntitiesUseNotFound() {
	queryServer := keeper.NewQueryServer(s.App.SponsorshipKeeper)
	ctx := sdk.WrapSDKContext(s.Ctx)
	voter := apptesting.CreateRandomAccounts(1)[0]

	tests := []struct {
		name  string
		query func() error
	}{
		{
			name: "vote",
			query: func() error {
				_, err := queryServer.Vote(ctx, &types.QueryVoteRequest{Voter: voter.String()})
				return err
			},
		},
		{
			name: "claim estimate",
			query: func() error {
				_, err := queryServer.EstimateClaim(ctx, &types.QueryEstimateClaim{Address: voter.String(), RollappId: "missing"})
				return err
			},
		},
		{
			name: "endorsement",
			query: func() error {
				_, err := queryServer.Endorsement(ctx, &types.QueryEndorsement{RollappId: "missing"})
				return err
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			err := tt.query()
			s.Require().ErrorIs(err, gerrc.ErrNotFound)
			s.Require().ErrorIs(err, collections.ErrNotFound)
		})
	}
}
