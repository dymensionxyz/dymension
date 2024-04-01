package ante

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	ethermint "github.com/evmos/ethermint/types"
)

var Registry = codectypes.NewInterfaceRegistry()

func init() {
	Registry.RegisterImplementations(
		(*tx.TxExtensionOptionI)(nil),
		&ethermint.ExtensionOptionsWeb3Tx{},
	)
}
