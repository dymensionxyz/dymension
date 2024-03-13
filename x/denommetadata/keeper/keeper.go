package keeper

import (
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
)

// Keeper provides a way to manage streamer module storage.
type Keeper struct {
	bankKeeper types.BankKeeper
	hooks      types.MultiDenomMetadataHooks
}

// NewKeeper returns a new instance of the incentive module keeper struct.
func NewKeeper(bankKeeper types.BankKeeper) *Keeper {

	return &Keeper{
		bankKeeper: bankKeeper,
		hooks:      nil,
	}
}

func (k *Keeper) GetBankKeeper() types.BankKeeper {
	return k.bankKeeper
}

/* -------------------------------------------------------------------------- */
/*                                    Hooks                                   */
/* -------------------------------------------------------------------------- */

// Set the rollapp hooks
func (k *Keeper) SetHooks(sh types.MultiDenomMetadataHooks) {
	if k.hooks != nil {
		panic("cannot set rollapp hooks twice")
	}
	k.hooks = sh
}

func (k *Keeper) GetHooks() types.MultiDenomMetadataHooks {
	return k.hooks
}
