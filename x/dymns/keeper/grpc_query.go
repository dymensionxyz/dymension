package keeper

import (
	"context"
	"slices"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ dymnstypes.QueryServer = queryServer{}

type queryServer struct {
	Keeper
}

// NewQueryServerImpl returns an implementation of the QueryServer interface
func NewQueryServerImpl(keeper Keeper) dymnstypes.QueryServer {
	return &queryServer{Keeper: keeper}
}

// Params queries the parameters of the module.
func (q queryServer) Params(goCtx context.Context, _ *dymnstypes.QueryParamsRequest) (*dymnstypes.QueryParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	params := q.GetParams(ctx)

	return &dymnstypes.QueryParamsResponse{Params: params}, nil
}

// DymName queries a Dym-Name by its name.
func (q queryServer) DymName(goCtx context.Context, req *dymnstypes.QueryDymNameRequest) (*dymnstypes.QueryDymNameResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	dymName := q.GetDymNameWithExpirationCheck(ctx, req.DymName)

	return &dymnstypes.QueryDymNameResponse{DymName: dymName}, nil
}

// ResolveDymNameAddresses resolves multiple Dym-Name Addresses to account address of each pointing to.
//
// For example:
//   - "my-name@dym" => "dym1a..."
//   - "another.my-name@dym" => "dym1b..."
//   - "my-name@nim" => "nim1..."
//   - (extra format) "0x1234...6789@nim" => "nim1..."
//   - (extra format) "dym1a...@nim" => "nim1..."
func (q queryServer) ResolveDymNameAddresses(goCtx context.Context, req *dymnstypes.ResolveDymNameAddressesRequest) (*dymnstypes.ResolveDymNameAddressesResponse, error) {
	if req == nil || len(req.Addresses) == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if addrCount := len(req.Addresses); addrCount > dymnstypes.LimitMaxElementsInApiRequest {
		return nil, status.Errorf(codes.InvalidArgument, "too many input addresses: %d > %d", addrCount, dymnstypes.LimitMaxElementsInApiRequest)
	}

	// There is a phishing attack vector like this: dym1.....@dym
	// With the current implementation, it is limited to 20 characters per name/sub-name
	// so, it is easier to recognize: dym1234.5678@dym

	ctx := sdk.UnwrapSDKContext(goCtx)

	var result []dymnstypes.ResultDymNameAddress
	for _, address := range req.Addresses {
		resolvedAddress, err := q.ResolveByDymNameAddress(ctx, address)

		r := dymnstypes.ResultDymNameAddress{
			Address: address,
		}

		if err != nil {
			r.Error = err.Error()
		} else {
			r.ResolvedAddress = resolvedAddress
		}

		result = append(result, r)
	}

	return &dymnstypes.ResolveDymNameAddressesResponse{
		ResolvedAddresses: result,
	}, nil
}

// DymNamesOwnedByAccount queries the Dym-Names owned by an account.
func (q queryServer) DymNamesOwnedByAccount(goCtx context.Context, req *dymnstypes.QueryDymNamesOwnedByAccountRequest) (*dymnstypes.QueryDymNamesOwnedByAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	dymNames, err := q.GetDymNamesOwnedBy(ctx, req.Owner)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &dymnstypes.QueryDymNamesOwnedByAccountResponse{
		DymNames: dymNames,
	}, nil
}

// SellOrder queries the active SO of a Dym-Name/Alias.
func (q queryServer) SellOrder(goCtx context.Context, req *dymnstypes.QuerySellOrderRequest) (*dymnstypes.QuerySellOrderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var assetType dymnstypes.AssetType
	switch req.AssetType {
	case dymnstypes.TypeName.PrettyName():
		if !dymnsutils.IsValidDymName(req.AssetId) {
			return nil, status.Errorf(codes.InvalidArgument, "invalid Dym-Name: %s", req.AssetId)
		}
		assetType = dymnstypes.TypeName
	case dymnstypes.TypeAlias.PrettyName():
		if !dymnsutils.IsValidAlias(req.AssetId) {
			return nil, status.Errorf(codes.InvalidArgument, "invalid alias: %s", req.AssetId)
		}
		assetType = dymnstypes.TypeAlias
	default:
		return nil, status.Errorf(codes.InvalidArgument, "invalid asset type: %s", req.AssetType)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	so := q.GetSellOrder(ctx, req.AssetId, assetType)
	if so == nil {
		return nil, status.Errorf(codes.NotFound, "no active Sell Order for %s '%s' at this moment", assetType.PrettyName(), req.AssetId)
	}

	return &dymnstypes.QuerySellOrderResponse{
		Result: *so,
	}, nil
}

// EstimateRegisterName estimates the cost to register a Dym-Name.
func (q queryServer) EstimateRegisterName(goCtx context.Context, req *dymnstypes.EstimateRegisterNameRequest) (*dymnstypes.EstimateRegisterNameResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if !dymnsutils.IsValidDymName(req.Name) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid dym name: %s", req.Name)
	}

	if req.Duration < 1 {
		return nil, status.Error(codes.InvalidArgument, "duration must be at least 1 year")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	existingDymNameRecord := q.GetDymName(ctx, req.Name) // can be nil if not registered before

	if existingDymNameRecord != nil && existingDymNameRecord.Owner != req.Owner {
		// check take-over permission
		if !existingDymNameRecord.IsExpiredAtCtx(ctx) {
			return nil, status.Errorf(
				codes.PermissionDenied,
				"you are not the owner of '%s'", req.Name,
			)
		}

		// we ignore the grace period since this is just an estimation
	}

	estimation := EstimateRegisterName(
		q.PriceParams(ctx),
		req.Name,
		existingDymNameRecord,
		req.Owner,
		req.Duration,
	)
	return &estimation, nil
}

// EstimateRegisterAlias estimates the cost to register an Alias.
func (q queryServer) EstimateRegisterAlias(goCtx context.Context, req *dymnstypes.EstimateRegisterAliasRequest) (*dymnstypes.EstimateRegisterAliasResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if !dymnsutils.IsValidAlias(req.Alias) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid alias: %s", req.Alias)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if req.RollappId != "" && req.Owner != "" {
		rollApp, found := q.rollappKeeper.GetRollapp(ctx, req.RollappId)
		if !found {
			return nil, status.Errorf(codes.NotFound, "RollApp not found: %s", req.RollappId)
		}

		if rollApp.Owner != req.Owner {
			return nil, status.Errorf(codes.PermissionDenied, "not the owner of the RollApp")
		}
	}

	if !q.CanUseAliasForNewRegistration(ctx, req.Alias) {
		return nil, status.Errorf(codes.AlreadyExists, "alias already taken: %s", req.Alias)
	}

	estimation := EstimateRegisterAlias(req.Alias, q.PriceParams(ctx))

	return &estimation, nil
}

// ReverseResolveAddress resolves multiple account addresses to Dym-Name Addresses which point to each.
// This function may return multiple possible Dym-Name-Addresses those point to each of the input address.
//
// For example: when we have "my-name@dym" resolves to "dym1a..."
// so reverse resolve will return "my-name@dym" when input is "dym1a..."
func (q queryServer) ReverseResolveAddress(goCtx context.Context, req *dymnstypes.ReverseResolveAddressRequest) (*dymnstypes.ReverseResolveAddressResponse, error) {
	if req == nil || len(req.Addresses) == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if addrCount := len(req.Addresses); addrCount > dymnstypes.LimitMaxElementsInApiRequest {
		return nil, status.Errorf(codes.InvalidArgument, "too many input addresses: %d > %d", addrCount, dymnstypes.LimitMaxElementsInApiRequest)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	workingChainId := req.WorkingChainId
	if workingChainId == "" {
		workingChainId = ctx.ChainID()
	}

	result := make(map[string]dymnstypes.ReverseResolveAddressResult)
	// Describe usage of Go Map: non-consensus state, for querying purpose only.

	addErrorResult := func(address string, err error) {
		result[address] = dymnstypes.ReverseResolveAddressResult{
			Error: err.Error(),
		}
	}

	addResult := func(address string, candidates []dymnstypes.ReverseResolvedDymNameAddress) {
		dymNameAddress := make([]string, 0, len(candidates))
		for _, candidate := range candidates {
			dymNameAddress = append(dymNameAddress, candidate.String())
		}

		result[address] = dymnstypes.ReverseResolveAddressResult{
			Candidates: dymNameAddress,
		}
	}

	for _, address := range req.Addresses {
		if !dymnsutils.PossibleAccountRegardlessChain(address) {
			// Simply ignore invalid address.
			// Invalid address is not included here to prevent wasting resources due to bad requests.
			continue
		}

		candidates, err := q.ReverseResolveDymNameAddress(ctx, address, workingChainId)
		if err != nil {
			addErrorResult(address, err)
			continue
		}

		addResult(address, candidates)
	}

	return &dymnstypes.ReverseResolveAddressResponse{
		Result:         result,
		WorkingChainId: workingChainId,
	}, nil
}

// TranslateAliasOrChainIdToChainId tries to translate an alias/handle to a chain id.
// If an alias/handle can not be translated to chain-id, it is treated as a chain-id and returns.
func (q queryServer) TranslateAliasOrChainIdToChainId(goCtx context.Context, req *dymnstypes.QueryTranslateAliasOrChainIdToChainIdRequest) (*dymnstypes.QueryTranslateAliasOrChainIdToChainIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.AliasOrChainId == "" {
		return nil, status.Error(codes.InvalidArgument, "empty alias or chain id")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	resolvedToChainId, success := q.tryResolveChainIdOrAliasToChainId(ctx, req.AliasOrChainId)
	if !success {
		resolvedToChainId = req.AliasOrChainId
	}

	return &dymnstypes.QueryTranslateAliasOrChainIdToChainIdResponse{
		ChainId: resolvedToChainId,
	}, nil
}

// BuyOrderById queries a Buy-Order by its id.
func (q queryServer) BuyOrderById(goCtx context.Context, req *dymnstypes.QueryBuyOrderByIdRequest) (*dymnstypes.QueryBuyOrderByIdResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if !dymnstypes.IsValidBuyOrderId(req.Id) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid Buy-Order ID: %s", req.Id)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	buyOrder := q.GetBuyOrder(ctx, req.Id)
	if buyOrder == nil {
		return nil, status.Error(codes.NotFound, "buy order not found")
	}

	return &dymnstypes.QueryBuyOrderByIdResponse{
		BuyOrder: *buyOrder,
	}, nil
}

// BuyOrdersPlacedByAccount queries the all the buy orders placed by an account.
func (q queryServer) BuyOrdersPlacedByAccount(goCtx context.Context, req *dymnstypes.QueryBuyOrdersPlacedByAccountRequest) (*dymnstypes.QueryBuyOrdersPlacedByAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	buyOrders, err := q.GetBuyOrdersByBuyer(ctx, req.Account)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &dymnstypes.QueryBuyOrdersPlacedByAccountResponse{
		BuyOrders: buyOrders,
	}, nil
}

// BuyOrdersByDymName queries all the buy orders of a Dym-Name.
func (q queryServer) BuyOrdersByDymName(goCtx context.Context, req *dymnstypes.QueryBuyOrdersByDymNameRequest) (*dymnstypes.QueryBuyOrdersByDymNameResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	buyOrders, err := q.GetBuyOrdersOfDymName(ctx, req.Name)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &dymnstypes.QueryBuyOrdersByDymNameResponse{
		BuyOrders: buyOrders,
	}, nil
}

// BuyOrdersOfDymNamesOwnedByAccount queries all the buy orders of all Dym-Names owned by an account.
func (q queryServer) BuyOrdersOfDymNamesOwnedByAccount(goCtx context.Context, req *dymnstypes.QueryBuyOrdersOfDymNamesOwnedByAccountRequest) (*dymnstypes.QueryBuyOrdersOfDymNamesOwnedByAccountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	ownedDymNames, err := q.GetDymNamesOwnedBy(ctx, req.Account)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	buyOrders := make([]dymnstypes.BuyOrder, 0)
	for _, dymName := range ownedDymNames {
		buyOrdersOfDymName, err := q.GetBuyOrdersOfDymName(ctx, dymName.Name)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}

		buyOrders = append(buyOrders, buyOrdersOfDymName...)
	}

	return &dymnstypes.QueryBuyOrdersOfDymNamesOwnedByAccountResponse{
		BuyOrders: buyOrders,
	}, nil
}

// Alias queries the chain_id associated as well as the Sell-Order and Buy-Order IDs relates to the alias.
func (q queryServer) Alias(goCtx context.Context, req *dymnstypes.QueryAliasRequest) (*dymnstypes.QueryAliasResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	chainId, success := q.tryResolveChainIdOrAliasToChainId(ctx, req.Alias)
	if !success {
		return nil, status.Errorf(codes.NotFound, "alias not found: %s", req.Alias)
	}

	if chainId == req.Alias {
		return nil, status.Errorf(codes.NotFound, "alias not found: %s", req.Alias)
	}

	var foundSellOrder bool
	var buyOrderIds []string
	var aliasesOfSameChain []string

	if q.IsAliasPresentsInParamsAsAliasOrChainId(ctx, req.Alias) {
		for _, aliasesOfChainId := range q.ChainsParams(ctx).AliasesOfChainIds {
			if chainId != aliasesOfChainId.ChainId {
				continue
			}

			aliasesOfSameChain = aliasesOfChainId.Aliases
			break
		}
	} else {
		foundSellOrder = q.GetSellOrder(ctx, req.Alias, dymnstypes.TypeAlias) != nil

		aliasToBuyOrderIdsRvlKey := dymnstypes.AliasToBuyOrderIdsRvlKey(req.Alias)
		buyOrderIds = q.GenericGetReverseLookupBuyOrderIdsRecord(ctx, aliasToBuyOrderIdsRvlKey).OrderIds

		aliasesOfSameChain = q.GetAliasesOfRollAppId(ctx, chainId)
	}

	if len(aliasesOfSameChain) > 0 {
		// exclude the alias itself
		aliasesOfSameChain = slices.DeleteFunc(aliasesOfSameChain, func(alias string) bool {
			return alias == req.Alias
		})
	}

	return &dymnstypes.QueryAliasResponse{
		ChainId:          chainId,
		FoundSellOrder:   foundSellOrder,
		BuyOrderIds:      buyOrderIds,
		SameChainAliases: aliasesOfSameChain,
	}, nil
}

// Aliases queries all the aliases for a chain id or all chains.
func (q queryServer) Aliases(goCtx context.Context, req *dymnstypes.QueryAliasesRequest) (*dymnstypes.QueryAliasesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.ChainId != "" && !dymnsutils.IsValidChainIdFormat(req.ChainId) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid chain id: %s", req.ChainId)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if req.ChainId != "" {
		aliases := q.GetEffectiveAliasesByChainId(ctx, req.ChainId)
		resp := dymnstypes.QueryAliasesResponse{}
		if len(aliases) > 0 {
			resp.AliasesByChainId = map[string]dymnstypes.MultipleAliases{
				req.ChainId: {
					Aliases: aliases,
				},
			}
		}
		return &resp, nil
	}

	aliasesByChainId := make(map[string]dymnstypes.MultipleAliases)
	for _, aliasesOfChainId := range q.ChainsParams(ctx).AliasesOfChainIds {
		if len(aliasesOfChainId.Aliases) < 1 {
			continue
		}
		aliasesByChainId[aliasesOfChainId.ChainId] = dymnstypes.MultipleAliases{
			Aliases: aliasesOfChainId.Aliases,
		}
	}

	rollAppsWithAliases := q.GetAllRollAppsWithAliases(ctx)
	if len(rollAppsWithAliases) > 0 {
		reservedAliases := q.GetAllAliasAndChainIdInParams(ctx)

		for _, rollAppWithAliases := range rollAppsWithAliases {
			aliases := rollAppWithAliases.Aliases

			// Remove the preserved aliases from record.
			// Please read the `processActiveAliasSellOrders` method (hooks.go) for more information.
			aliases = slices.DeleteFunc(aliases, func(a string) bool {
				_, found := reservedAliases[a]
				return found
			})

			if len(aliases) < 1 {
				continue
			}

			existingAliases := aliasesByChainId[rollAppWithAliases.ChainId]
			aliasesByChainId[rollAppWithAliases.ChainId] = dymnstypes.MultipleAliases{
				Aliases: append(existingAliases.Aliases, aliases...),
			}
		}
	}

	return &dymnstypes.QueryAliasesResponse{
		AliasesByChainId: aliasesByChainId,
	}, nil
}

// BuyOrdersByAlias queries all the buy orders of an Alias.
func (q queryServer) BuyOrdersByAlias(goCtx context.Context, req *dymnstypes.QueryBuyOrdersByAliasRequest) (*dymnstypes.QueryBuyOrdersByAliasResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if !dymnsutils.IsValidAlias(req.Alias) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid alias: %s", req.Alias)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var buyOrders []dymnstypes.BuyOrder
	if !q.IsAliasPresentsInParamsAsAliasOrChainId(ctx, req.Alias) {
		// We ignore the aliases which presents in the params because they are prohibited from trading.
		// Please read the `processActiveAliasSellOrders` method (hooks.go) for more information.

		var err error
		buyOrders, err = q.GetBuyOrdersOfAlias(ctx, req.Alias)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &dymnstypes.QueryBuyOrdersByAliasResponse{
		BuyOrders: buyOrders,
	}, nil
}

// BuyOrdersOfAliasesLinkedToRollApp queries all the buy orders of all Aliases linked to a RollApp.
func (q queryServer) BuyOrdersOfAliasesLinkedToRollApp(goCtx context.Context, req *dymnstypes.QueryBuyOrdersOfAliasesLinkedToRollAppRequest) (*dymnstypes.QueryBuyOrdersOfAliasesLinkedToRollAppResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if !dymnsutils.IsValidChainIdFormat(req.RollappId) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid RollApp ID: %s", req.RollappId)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !q.IsRollAppId(ctx, req.RollappId) {
		return nil, status.Errorf(codes.NotFound, "RollApp not found: %s", req.RollappId)
	}

	var allBuyOrders []dymnstypes.BuyOrder

	aliases := q.GetAliasesOfRollAppId(ctx, req.RollappId)
	for _, alias := range aliases {
		if q.IsAliasPresentsInParamsAsAliasOrChainId(ctx, alias) {
			// ignore
			// Please read the `processActiveAliasSellOrders` method (hooks.go) for more information.
			continue
		}

		buyOrders, err := q.GetBuyOrdersOfAlias(ctx, alias)
		if err != nil {
			return nil, status.Error(codes.Internal, errorsmod.Wrapf(err, "alias: %s", alias).Error())
		}

		allBuyOrders = append(allBuyOrders, buyOrders...)
	}

	return &dymnstypes.QueryBuyOrdersOfAliasesLinkedToRollAppResponse{
		BuyOrders: allBuyOrders,
	}, nil
}
