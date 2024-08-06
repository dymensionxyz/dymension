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

// CmdQueryBuyOrder is the CLI command for querying Buy-Orders a Dym-Name
func CmdQueryBuyOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "buy-order [target]",
		Aliases: []string{"offer"},
		Short:   "Get list of Buy-Orders corresponding to the target",
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

			var offers []dymnstypes.BuyOffer
			var err error

			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := dymnstypes.NewQueryClient(clientCtx)
			queryCtx := cmd.Context()

			switch targetType {
			case targetTypeById:
				var offer *dymnstypes.BuyOffer
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
				if err := clientCtx.PrintProto(&offer); err != nil {
					return err
				}
			}

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	cmd.Flags().String(flagTargetType, "", fmt.Sprintf("Target type to query for, one of: %s/%s/%s/%s", targetTypeById, targetTypeBuyer, targetTypeOwner, targetTypeDymName))

	return cmd
}

// queryOfferById fetches a Buy-Order by its ID
func queryOfferById(queryClient dymnstypes.QueryClient, ctx context.Context, offerId string) (*dymnstypes.BuyOffer, error) {
	if !dymnstypes.IsValidBuyOfferId(offerId) {
		return nil, fmt.Errorf("input Offer-ID '%s' is not a valid Offer-ID", offerId)
	}

	res, err := queryClient.BuyOfferById(ctx, &dymnstypes.QueryBuyOfferByIdRequest{
		Id: offerId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch buy offer by ID '%s': %w", offerId, err)
	}

	return &res.Offer, nil
}

// queryOffersPlacedByBuyer fetches Buy-Orders placed by a buyer
func queryOffersPlacedByBuyer(queryClient dymnstypes.QueryClient, ctx context.Context, buyer string) ([]dymnstypes.BuyOffer, error) {
	if !dymnsutils.IsValidBech32AccountAddress(buyer, true) {
		return nil, fmt.Errorf("input buyer address '%s' is not a valid bech32 account address", buyer)
	}

	res, err := queryClient.BuyOffersPlacedByAccount(ctx, &dymnstypes.QueryBuyOffersPlacedByAccountRequest{
		Account: buyer,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Buy-Orders placed by buyer '%s': %w", buyer, err)
	}

	return res.Offers, nil
}

// queryOffersOfDymNamesOwnedByOwner fetches all Buy-Orders of all Dym-Names owned by an owner
func queryOffersOfDymNamesOwnedByOwner(queryClient dymnstypes.QueryClient, ctx context.Context, owner string) ([]dymnstypes.BuyOffer, error) {
	if !dymnsutils.IsValidBech32AccountAddress(owner, true) {
		return nil, fmt.Errorf("input owner address '%s' is not a valid bech32 account address", owner)
	}

	res, err := queryClient.BuyOffersOfDymNamesOwnedByAccount(ctx, &dymnstypes.QueryBuyOffersOfDymNamesOwnedByAccountRequest{
		Account: owner,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Buy-Orders of Dym-Names owned by '%s': %w", owner, err)
	}

	return res.Offers, nil
}

// queryOffersByDymName fetches all Buy-Orders of a Dym-Name
func queryOffersByDymName(queryClient dymnstypes.QueryClient, ctx context.Context, dymName string) ([]dymnstypes.BuyOffer, error) {
	if !dymnsutils.IsValidDymName(dymName) {
		return nil, fmt.Errorf("input Dym-Name '%s' is not a valid Dym-Name", dymName)
	}

	res, err := queryClient.BuyOffersByDymName(ctx, &dymnstypes.QueryBuyOffersByDymNameRequest{
		Name: dymName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Buy-Orders of Dym-Name '%s': %w", dymName, err)
	}

	return res.Offers, nil
}
