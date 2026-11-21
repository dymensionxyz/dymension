package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/dymensionxyz/dymension/v3/utils/uinv"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

var invs = uinv.NamedFuncsList[Keeper]{
	{Name: "escrow-solvency", Func: InvariantEscrowSolvency},
	{Name: "validation-response-integrity", Func: InvariantValidationResponseIntegrity},
	{Name: "recipient-spend-within-limit", Func: InvariantRecipientSpendWithinLimit},
}

func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	invs.RegisterInvariants(types.ModuleName, ir, k)
}

func InvariantValidationResponseIntegrity(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		counts := map[string]uint64{}
		if err := k.validationResponses.Walk(ctx, nil, func(key collections.Pair[[]byte, uint64], _ types.ValidationResponse) (bool, error) {
			if has, err := k.validationRequests.Has(ctx, key.K1()); err != nil {
				return true, err
			} else if !has {
				return true, fmt.Errorf("orphaned validation response: %x/%d", key.K1(), key.K2())
			}
			counts[string(key.K1())]++
			return false, nil
		}); err != nil {
			return err
		}
		return k.validationRequests.Walk(ctx, nil, func(hash []byte, req types.ValidationRequest) (bool, error) {
			if counts[string(hash)] != req.ResponseCount {
				return true, fmt.Errorf("validation response count mismatch: %x: stored %d actual %d", hash, req.ResponseCount, counts[string(hash)])
			}
			return false, nil
		})
	})
}

// DO NOT DELETE
func AllInvariants(k Keeper) sdk.Invariant {
	return invs.All(types.ModuleName, k)
}

// InvariantEscrowSolvency checks that the agent module account holds at least
// the sum of all per-agent escrow ledger balances, and that no ledger entry is
// invalid or empty (runtime removes zero balances).
func InvariantEscrowSolvency(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		total := sdk.NewCoins()
		var errs []error
		if err := k.escrows.Walk(ctx, nil, func(agentID string, e types.AgentEscrow) (stop bool, err error) {
			if err := e.Balance.Validate(); err != nil {
				errs = append(errs, fmt.Errorf("invalid escrow balance: agent %s: %w", agentID, err))
				return false, nil
			}
			if e.Balance.IsZero() {
				errs = append(errs, fmt.Errorf("empty escrow ledger entry: agent %s", agentID))
				return false, nil
			}
			total = total.Add(e.Balance...)
			return false, nil
		}); err != nil {
			return err
		}
		if len(errs) > 0 {
			return fmt.Errorf("escrow ledger: %v", errs)
		}

		moduleBalance := k.bankKeeper.GetAllBalances(ctx, authtypes.NewModuleAddress(types.ModuleName))
		if !total.IsAllLTE(moduleBalance) {
			return fmt.Errorf("escrow ledger exceeds module account balance: ledger %s, module %s", total, moduleBalance)
		}
		return nil
	})
}

// InvariantRecipientSpendWithinLimit checks that no recipient's window total
// exceeds the agent's per-recipient cap, when one is set.
func InvariantRecipientSpendWithinLimit(k Keeper) uinv.Func {
	return uinv.AnyErrorIsBreaking(func(ctx sdk.Context) error {
		return k.recipientSpend.Walk(ctx, nil, func(key collections.Pair[string, []byte], rec types.RecipientSpendWindow) (bool, error) {
			agent, found := k.GetAgent(ctx, key.K1())
			if !found {
				return true, fmt.Errorf("recipient spend for unknown agent: %s", key.K1())
			}
			limit := agent.SpendLimitPerRecipientPerWindow
			if limit.IsNil() || limit.IsZero() {
				return false, nil
			}
			if rec.Spent.IsNil() || rec.Spent.IsNegative() || rec.Spent.GT(limit) {
				return true, fmt.Errorf("recipient spend out of bounds: agent %s recipient %s spent %s limit %s", key.K1(), sdk.AccAddress(key.K2()), rec.Spent, limit)
			}
			return false, nil
		})
	})
}
