package keeper

import (
	"fmt"

	"github.com/tendermint/tendermint/libs/log"

	sdkfraudtypes "github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	fraudtypes "github.com/dymensionxyz/dymension/v3/app/fraudproof"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type (
	Keeper struct {
		cdc      codec.BinaryCodec
		storeKey storetypes.StoreKey
		memKey   storetypes.StoreKey

		fraudProofVerifier FraudProofVerifier
		hooks              types.MultiRollappHooks
		paramstore         paramtypes.Subspace
	}
)

type FraudProofVerifier interface {
	Init(*sdkfraudtypes.FraudProof) error
	Verify(*sdkfraudtypes.FraudProof) error
}

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
		hooks:      nil,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// SetHooks sets the rollapp hooks
func (k *Keeper) SetHooks(sh types.MultiRollappHooks) {
	if k.hooks != nil {
		panic("cannot set rollapp hooks twice")
	}
	k.hooks = sh
}

func (k *Keeper) GetHooks() types.MultiRollappHooks {
	return k.hooks
}

func (k *Keeper) SetFraudProofVerifier(fpv fraudtypes.Verifier) {
	k.fraudProofVerifier = fpv
}
