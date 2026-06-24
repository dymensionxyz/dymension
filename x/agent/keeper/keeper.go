package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/dymensionxyz/dymension/v3/x/common/tee"
)

type Keeper struct {
	// verifier checks TEE attestation tokens. Stored now, used from issue 4.
	verifier   tee.Verifier
	bankKeeper types.BankKeeper

	params collections.Item[types.Params]
	// agents keyed by id
	agents collections.Map[string, types.Agent]
	// actionLog keyed by (agent_id, seq)
	actionLog collections.Map[collections.Pair[string, uint64], types.ActionLogEntry]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	service store.KVStoreService,
	verifier tee.Verifier,
	bankKeeper types.BankKeeper,
) *Keeper {
	sb := collections.NewSchemaBuilder(service)

	return &Keeper{
		verifier:   verifier,
		bankKeeper: bankKeeper,
		params: collections.NewItem(sb, collections.NewPrefix(types.KeyParams),
			types.KeyParams,
			collcompat.ProtoValue[types.Params](cdc)),
		agents: collections.NewMap(sb, collections.NewPrefix(types.KeyAgents),
			types.KeyAgents,
			collections.StringKey,
			collcompat.ProtoValue[types.Agent](cdc)),
		actionLog: collections.NewMap(sb, collections.NewPrefix(types.KeyActionLog),
			types.KeyActionLog,
			collections.PairKeyCodec(collections.StringKey, collections.Uint64Key),
			collcompat.ProtoValue[types.ActionLogEntry](cdc)),
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	p, err := k.params.Get(ctx)
	if err != nil {
		panic(err)
	}
	return p
}

func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	return k.params.Set(ctx, p)
}

func (k Keeper) AgentRegistrationFee(ctx sdk.Context) sdk.Coin {
	return k.GetParams(ctx).AgentRegistrationFee
}

func (k Keeper) HasAgent(ctx sdk.Context, id string) (bool, error) {
	return k.agents.Has(ctx, id)
}

func (k Keeper) GetAgent(ctx sdk.Context, id string) (types.Agent, error) {
	return k.agents.Get(ctx, id)
}

func (k Keeper) SetAgent(ctx sdk.Context, a types.Agent) error {
	return k.agents.Set(ctx, a.Id, a)
}
