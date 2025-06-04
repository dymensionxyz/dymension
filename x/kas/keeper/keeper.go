package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/log"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	hypercorekeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/keeper"
	"github.com/dymensionxyz/dymension/v3/internal/collcompat"
	"github.com/dymensionxyz/dymension/v3/x/kas/types"
)

type Keeper struct {
	authority string // authority is the x/gov module account
	cdc       codec.BinaryCodec

	hypercoreK *hypercorekeeper.Keeper

	// Is this module fully bootstrapped, i.e. ready to use?
	bootstrapped collections.Item[bool]

	mailbox collections.Item[string] // HexAddress format
	ism     collections.Item[string] // HexAddress format

	// The Kaspa escrow outpoint which must be used in all TXs. It's updated only on confirmations.
	outpoint collections.Item[types.TransactionOutpoint]

	// Tracks the processed withdrawals to avoid double relaying. May only update when updating outpoint too.
	processedWithdrawals collections.KeySet[collections.Pair[uint64, []byte]]
}

func NewKeeper(
	cdc codec.BinaryCodec,
	service store.KVStoreService,
	authority string,
	hypercoreK *hypercorekeeper.Keeper,
) *Keeper {
	_, err := sdk.AccAddressFromBech32(authority)
	if err != nil {
		panic(fmt.Errorf("invalid x/sequencer authority address: %w", err))
	}
	sb := collections.NewSchemaBuilder(service)

	bootstrapped := collections.NewItem(sb, collections.NewPrefix(types.KeyBootstrapped),
		types.KeyBootstrapped,
		collections.BoolValue)

	ism := collections.NewItem(sb, collections.NewPrefix(types.KeyISM),
		types.KeyISM,
		collections.StringValue)

	mailbox := collections.NewItem(sb, collections.NewPrefix(types.KeyMailbox),
		types.KeyMailbox,
		collections.StringValue)

	outpoint := collections.NewItem(sb, collections.NewPrefix(types.KeyOutpoint),
		types.KeyOutpoint,
		collcompat.ProtoValue[types.TransactionOutpoint](cdc))

	processedWithdrawals := collections.NewKeySet(sb, collections.NewPrefix(types.KeyProcessedWithdrawals),
		types.KeyProcessedWithdrawals,
		collections.PairKeyCodec(collections.Uint64Key, collections.BytesKey))

	return &Keeper{
		cdc:                  cdc,
		authority:            authority,
		hypercoreK:           hypercoreK,
		bootstrapped:         bootstrapped,
		ism:                  ism,
		mailbox:              mailbox,
		outpoint:             outpoint,
		processedWithdrawals: processedWithdrawals,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) TransactionsEnabled(ctx sdk.Context) bool {
	ret, err := k.bootstrapped.Get(ctx)
	if err != nil {
		panic(err)
	}
	return ret
}
