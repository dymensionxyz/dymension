package keeper

import (
	"cosmossdk.io/collections"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/kas/types"
)

func InitGenesis(ctx sdk.Context, k *Keeper, g types.GenesisState) {
	if err := k.bootstrapped.Set(ctx, g.Bootstrapped); err != nil {
		panic(err)
	}

	if g.Mailbox != "" {
		if err := k.mailbox.Set(ctx, g.Mailbox); err != nil {
			panic(err)
		}
	}
	if g.Ism != "" {
		if err := k.ism.Set(ctx, g.Ism); err != nil {
			panic(err)
		}
	}
	if g.Outpoint != nil {
		if err := k.outpoint.Set(ctx, *g.Outpoint); err != nil {
			panic(err)
		}
	}
	for _, w := range g.ProcessedWithdrawals {
		if err := k.SetProcessedWithdrawal(ctx, *w); err != nil {
			panic(err)
		}
	}
}

func ExportGenesis(ctx sdk.Context, k *Keeper) *types.GenesisState {
	var err error
	g := types.GenesisState{}

	g.Bootstrapped, err = k.bootstrapped.Get(ctx)
	if err != nil {
		panic(err)
	}

	mailbox, err := k.mailbox.Get(ctx)
	if err == nil {
		g.Mailbox = mailbox
	}

	ism, err := k.ism.Get(ctx)
	if err == nil {
		g.Ism = ism
	}

	outpoint, err := k.outpoint.Get(ctx)
	if err == nil {
		g.Outpoint = &outpoint
	}

	err = k.processedWithdrawals.Walk(ctx, nil, func(key collections.Pair[uint64, []byte]) (stop bool, err error) {
		// see https://github.com/dymensionxyz/hyperlane-cosmos/blob/fb914a5ba702f70a428a475968b886891cb1ad77/x/core/keeper/genesis.go#L61
		g.ProcessedWithdrawals = append(g.ProcessedWithdrawals, &types.WithdrawalID{
			MessageId: hyperutil.HexAddress(key.K2()).String(),
		})
		return false, nil
	})
	if err != nil {
		panic(err)
	}

	return &g
}
