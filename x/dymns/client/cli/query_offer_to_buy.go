package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/spf13/cobra"

	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

const (
	flagTargetType = "target-type"

	targetTypeById    = "offer-id"
	targetTypeBuyer   = "buyer"
	targetTypeOwner   = "owner"
	targetTypeDymName = "dym-name"
)

func CmdQueryOfferToBuy() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "offer-to-buy [Dym-Name]",
		Aliases: []string{"offer"},
		Short:   "Get list of offers to buy a Dym-Name.",
		Example: fmt.Sprintf(
			`%s q %s offer 1 --%s=%s
%s q %s offer dym1buyer --%s=%s
%s q %s offer dym1owner --%s=%s
%s q %s offer myname --%s=%s
`,
			version.AppName, dymnstypes.ModuleName, flagTargetType, targetTypeById,
			version.AppName, dymnstypes.ModuleName, flagTargetType, targetTypeBuyer,
			version.AppName, dymnstypes.ModuleName, flagTargetType, targetTypeOwner,
			version.AppName, dymnstypes.ModuleName, flagTargetType, targetTypeDymName,
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetType, _ := cmd.Flags().GetString(flagTargetType)

			if targetType == "" {
				return fmt.Errorf("flag --%s is required", flagTargetType)
			}

			var offers []dymnstypes.OfferToBuy
			var err error

			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := dymnstypes.NewQueryClient(clientCtx)
			queryCtx := cmd.Context()

			switch targetType {
			case targetTypeById:
				var offer *dymnstypes.OfferToBuy
				offer, err = queryOfferById(queryClient, queryCtx, args[0])
				if err == nil && offer != nil {
					offers = append(offers, *offer)
				}
			case targetTypeBuyer:
				offers, err = queryOffersPlacedByBuyer(queryClient, queryCtx, args[0])
			case targetTypeOwner:
				offers, err = queryOffersOfDymNamesOwnedByOwner(queryClient, queryCtx, args[0])
			case targetTypeDymName:
				offers, err = queryOffersByDymName(queryClient, queryCtx, args[0])
			default:
				return fmt.Errorf("invalid target type: %s", targetType)
			}

			if err != nil {
				return err
			}

			if len(offers) == 0 {
				fmt.Println("No offers found")
				return nil
			}

			for _, offer := range offers {
				return clientCtx.PrintProto(&offer)
			}

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	cmd.Flags().String(flagTargetType, "", fmt.Sprintf("Target type to query for, one of: %s/%s/%s/%s", targetTypeById, targetTypeBuyer, targetTypeOwner, targetTypeDymName))

	return cmd
}

func queryOfferById(queryClient dymnstypes.QueryClient, ctx context.Context, offerId string) (*dymnstypes.OfferToBuy, error) {
	if !dymnsutils.IsValidBuyNameOfferId(offerId) {
		return nil, fmt.Errorf("input Offer-ID '%s' is not a valid Offer-ID", offerId)
	}

	res, err := queryClient.OfferToBuyById(ctx, &dymnstypes.QueryOfferToBuyByIdRequest{
		Id: offerId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Offer-To-Buy by ID '%s': %w", offerId, err)
	}

	return &res.Offer, nil
}

func queryOffersPlacedByBuyer(queryClient dymnstypes.QueryClient, ctx context.Context, buyer string) ([]dymnstypes.OfferToBuy, error) {
	if !dymnsutils.IsValidBech32AccountAddress(buyer, true) {
		return nil, fmt.Errorf("input buyer address '%s' is not a valid bech32 account address", buyer)
	}

	res, err := queryClient.OffersToBuyPlacedByAccount(ctx, &dymnstypes.QueryOffersToBuyPlacedByAccountRequest{
		Account: buyer,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Offers-To-Buy placed by buyer '%s': %w", buyer, err)
	}

	return res.Offers, nil
}

func queryOffersOfDymNamesOwnedByOwner(queryClient dymnstypes.QueryClient, ctx context.Context, owner string) ([]dymnstypes.OfferToBuy, error) {
	if !dymnsutils.IsValidBech32AccountAddress(owner, true) {
		return nil, fmt.Errorf("input owner address '%s' is not a valid bech32 account address", owner)
	}

	res, err := queryClient.OffersToBuyOfDymNamesOwnedByAccount(ctx, &dymnstypes.QueryOffersToBuyOfDymNamesOwnedByAccountRequest{
		Account: owner,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Offers-To-Buy of Dym-Names owned by '%s': %w", owner, err)
	}

	return res.Offers, nil
}

func queryOffersByDymName(queryClient dymnstypes.QueryClient, ctx context.Context, dymName string) ([]dymnstypes.OfferToBuy, error) {
	if !dymnsutils.IsValidDymName(dymName) {
		return nil, fmt.Errorf("input Dym-Name '%s' is not a valid Dym-Name", dymName)
	}

	res, err := queryClient.OffersToBuyByDymName(ctx, &dymnstypes.QueryOffersToBuyByDymNameRequest{
		Name: dymName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Offers-To-Buy of Dym-Name '%s': %w", dymName, err)
	}

	return res.Offers, nil
}
