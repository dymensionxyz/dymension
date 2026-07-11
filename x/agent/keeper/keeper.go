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
	// verifier is injected so the attestation seam can be faked in tests; the
	// real verifier does GCP PKI + rego evaluation.
	verifier   tee.Verifier
	bankKeeper types.BankKeeper

	params     collections.Item[types.Params]
	agents     collections.Map[string, types.Agent]
	actionLog  collections.Map[collections.Pair[string, uint64], types.ActionLogEntry]
	feedback   collections.Map[collections.Pair[string, string], types.Feedback]
	reputation collections.Map[string, types.Reputation]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	service store.KVStoreService,
	verifier tee.Verifier,
	bankKeeper types.BankKeeper,
) *Keeper {
	sb := collections.NewSchemaBuilder(service)

	k := &Keeper{
		verifier:   verifier,
		bankKeeper: bankKeeper,
		params: collections.NewItem(sb, collections.NewPrefix(types.KeyParams),
			"params", collcompat.ProtoValue[types.Params](cdc)),
		agents: collections.NewMap(sb, collections.NewPrefix(types.KeyAgents),
			"agents", collections.StringKey, collcompat.ProtoValue[types.Agent](cdc)),
		actionLog: collections.NewMap(sb, collections.NewPrefix(types.KeyActionLog),
			"action_log", collections.PairKeyCodec(collections.StringKey, collections.Uint64Key),
			collcompat.ProtoValue[types.ActionLogEntry](cdc)),
		feedback: collections.NewMap(sb, collections.NewPrefix(types.KeyFeedback),
			"feedback", collections.PairKeyCodec(collections.StringKey, collections.StringKey),
			collcompat.ProtoValue[types.Feedback](cdc)),
		reputation: collections.NewMap(sb, collections.NewPrefix(types.KeyReputation),
			"reputation", collections.StringKey, collcompat.ProtoValue[types.Reputation](cdc)),
	}

	if _, err := sb.Build(); err != nil {
		panic(err)
	}
	return k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
