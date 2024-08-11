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
	targetTypeDymName = "name"
	targetTypeAlias   = "alias"
	targetTypeRollApp = "rollapp"
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
%s q %s offer my-name --%s=%s
%s q %s offer dym --%s=%s
%s q %s offer rollapp_1-1 --%s=%s
`,
			version.AppName, dymnstypes.ModuleName, flagTargetType, targetTypeById,
			version.AppName, dymnstypes.ModuleName, flagTargetType, targetTypeBuyer,
			version.AppName, dymnstypes.ModuleName, flagTargetType, targetTypeOwner,
			version.AppName, dymnstypes.ModuleName, flagTargetType, targetTypeDymName,
			version.AppName, dymnstypes.ModuleName, flagTargetType, targetTypeAlias,
			version.AppName, dymnstypes.ModuleName, flagTargetType, targetTypeRollApp,
		),
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetType, err := cmd.Flags().GetString(flagTargetType)
			if err != nil {
				return err
			}

			if targetType == "" {
				return fmt.Errorf("flag --%s is required", flagTargetType)
			}

			var offers []dymnstypes.BuyOrder

			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := dymnstypes.NewQueryClient(clientCtx)
			queryCtx := cmd.Context()

			switch targetType {
			case targetTypeById:
				var offer *dymnstypes.BuyOrder
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
			case targetTypeAlias:
				offers, err = queryOffersByAlias(queryClient, queryCtx, args[0])
			case targetTypeRollApp:
				offers, err = queryOffersOfAliasesLinkedToRollApp(queryClient, queryCtx, args[0])
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

			for i, offer := range offers {
				if i > 0 {
					fmt.Println("___")
				}
				if err := printBuyOrder(offer); err != nil {
					fmt.Printf("Bad offer: %s\n", err.Error())
					fmt.Println("Raw: ", offer)
				}
			}

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	cmd.Flags().String(flagTargetType, "", fmt.Sprintf("Target type to query for, one of: %s/%s/%s/%s/%s/%s", targetTypeById, targetTypeBuyer, targetTypeOwner, targetTypeDymName, targetTypeAlias, targetTypeRollApp))

	return cmd
}

func printBuyOrder(offer dymnstypes.BuyOrder) error {
	if err := offer.Validate(); err != nil {
		return err
	}
	fmt.Printf("ID: %s\n", offer.Id)
	fmt.Printf(" Buyer: %s\n", offer.Buyer)
	fmt.Printf(" Type: %s\n", offer.AssetType.FriendlyString())
	if offer.AssetType == dymnstypes.TypeName {
		fmt.Printf(" Dym-Name: %s\n", offer.AssetId)
	} else if offer.AssetType == dymnstypes.TypeAlias {
		fmt.Printf(" Alias: %s\n", offer.AssetId)
		fmt.Printf(" For RollApp: %s\n", offer.Params[0])
	}
	fmt.Printf(" Offer Price: %s\n", offer.OfferPrice)
	if estAmt, ok := toEstimatedCoinAmount(offer.OfferPrice); ok {
		fmt.Printf("   (~ %s)\n", estAmt)
	}
	fmt.Printf(" Counterparty Offer Price: ")
	if offer.CounterpartyOfferPrice != nil {
		fmt.Printf("%s\n", *offer.CounterpartyOfferPrice)
		if estAmt, ok := toEstimatedCoinAmount(*offer.CounterpartyOfferPrice); ok {
			fmt.Printf("   (~ %s)\n", estAmt)
		}
	} else {
		fmt.Println("None")
	}
	return nil
}

// queryOfferById fetches a Buy-Order by its ID
func queryOfferById(queryClient dymnstypes.QueryClient, ctx context.Context, orderId string) (*dymnstypes.BuyOrder, error) {
	if !dymnstypes.IsValidBuyOrderId(orderId) {
		return nil, fmt.Errorf("input is not a valid Buy-Order ID: %s", orderId)
	}

	res, err := queryClient.BuyOrderById(ctx, &dymnstypes.QueryBuyOrderByIdRequest{
		Id: orderId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch buy offer by ID '%s': %w", orderId, err)
	}

	return &res.BuyOrder, nil
}

// queryOffersPlacedByBuyer fetches Buy-Orders placed by a buyer
func queryOffersPlacedByBuyer(queryClient dymnstypes.QueryClient, ctx context.Context, buyer string) ([]dymnstypes.BuyOrder, error) {
	if !dymnsutils.IsValidBech32AccountAddress(buyer, true) {
		return nil, fmt.Errorf("input buyer address '%s' is not a valid bech32 account address", buyer)
	}

	res, err := queryClient.BuyOrdersPlacedByAccount(ctx, &dymnstypes.QueryBuyOrdersPlacedByAccountRequest{
		Account: buyer,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Buy-Orders placed by buyer '%s': %w", buyer, err)
	}

	return res.BuyOrders, nil
}

// queryOffersOfDymNamesOwnedByOwner fetches all Buy-Orders of all Dym-Names owned by an owner
func queryOffersOfDymNamesOwnedByOwner(queryClient dymnstypes.QueryClient, ctx context.Context, owner string) ([]dymnstypes.BuyOrder, error) {
	if !dymnsutils.IsValidBech32AccountAddress(owner, true) {
		return nil, fmt.Errorf("input owner address is not a valid bech32 account address: %s", owner)
	}

	res, err := queryClient.BuyOrdersOfDymNamesOwnedByAccount(ctx, &dymnstypes.QueryBuyOrdersOfDymNamesOwnedByAccountRequest{
		Account: owner,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Buy-Orders of Dym-Names owned by '%s': %w", owner, err)
	}

	return res.BuyOrders, nil
}

// queryOffersOfAliasesLinkedToRollApp fetches all Buy-Orders of all Aliases linked to a RollApp
func queryOffersOfAliasesLinkedToRollApp(queryClient dymnstypes.QueryClient, ctx context.Context, rollAppId string) ([]dymnstypes.BuyOrder, error) {
	if !dymnsutils.IsValidChainIdFormat(rollAppId) {
		return nil, fmt.Errorf("input RollApp ID is invalid: %s", rollAppId)
	}

	res, err := queryClient.BuyOrdersOfAliasesLinkedToRollApp(ctx, &dymnstypes.QueryBuyOrdersOfAliasesLinkedToRollAppRequest{
		RollappId: rollAppId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Buy-Orders of aliases linked to '%s': %w", rollAppId, err)
	}

	return res.BuyOrders, nil
}

// queryOffersByDymName fetches all Buy-Orders of a Dym-Name
func queryOffersByDymName(queryClient dymnstypes.QueryClient, ctx context.Context, dymName string) ([]dymnstypes.BuyOrder, error) {
	if !dymnsutils.IsValidDymName(dymName) {
		return nil, fmt.Errorf("input is not a valid Dym-Name: %s", dymName)
	}

	res, err := queryClient.BuyOrdersByDymName(ctx, &dymnstypes.QueryBuyOrdersByDymNameRequest{
		Name: dymName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Buy-Orders of Dym-Name '%s': %w", dymName, err)
	}

	return res.BuyOrders, nil
}

func queryOffersByAlias(queryClient dymnstypes.QueryClient, ctx context.Context, alias string) ([]dymnstypes.BuyOrder, error) {
	if !dymnsutils.IsValidAlias(alias) {
		return nil, fmt.Errorf("input is not a valid alias: %s", alias)
	}

	res, err := queryClient.BuyOrdersByAlias(ctx, &dymnstypes.QueryBuyOrdersByAliasRequest{
		Alias: alias,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch Buy-Orders of Alias '%s': %w", alias, err)
	}

	return res.BuyOrders, nil
}
