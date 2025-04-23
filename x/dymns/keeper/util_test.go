package keeper_test

import (
	"github.com/ethereum/go-ethereum/common"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

var dymNsModuleAccAddr = authtypes.NewModuleAddress(dymnstypes.ModuleName)

// ta stands for test-address, a simple wrapper for generating account for testing purpose.
// Usage is short, memorable, easy to type.
// The generated address is predictable, deterministic, supports output multiple formats.
type ta struct {
	bz []byte
}

// testAddr creates a general 20-bytes address from seed.
func testAddr(no uint64) ta {
	bz1 := sdk.Uint64ToBigEndian(no)
	bz2 := make([]byte, 20)
	copy(bz2, bz1)
	return ta{bz: bz2}
}

// testICAddr creates a 32-bytes address of Interchain Account from seed.
func testICAddr(no uint64) ta {
	bz1 := sdk.Uint64ToBigEndian(no)
	bz2 := make([]byte, 32)
	copy(bz2, bz1)
	return ta{bz: bz2}
}

func (a ta) bytes() []byte {
	return a.bz
}

func (a ta) bech32() string {
	return a.bech32C(params.AccountAddressPrefix)
}

func (a ta) bech32Valoper() string {
	return a.bech32C(params.AccountAddressPrefix + "valoper")
}

func (a ta) bech32C(customHrp string) string {
	return sdk.MustBech32ifyAddressBytes(customHrp, a.bz)
}

func (a ta) fallback() dymnstypes.FallbackAddress {
	return a.bz
}

func (a ta) hexStr() string {
	return dymnsutils.GetHexAddressFromBytes(a.bz)
}

func (a ta) checksumHex() string {
	if len(a.bz) != 20 {
		panic("invalid call")
	}
	return common.BytesToAddress(a.bz).Hex()
}
