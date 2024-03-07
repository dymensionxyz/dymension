package keeper_test

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	extendedevmkeeper "github.com/dymensionxyz/dymension/v3/x/evm/keeper"
	"github.com/ethereum/go-ethereum/common"
	"github.com/stretchr/testify/suite"
)

var _ = suite.TestingSuite(nil)

func (suite *KeeperTestSuite) TestHookOperation_BeforeEpochStart() {
	suite.SetupTest()

	denomAdym := createDenomMetadata("adym", "DYM", 18)
	suite.Require().Nil(denomAdym.Validate(), "bad setup")
	suite.App.BankKeeper.SetDenomMetaData(suite.Ctx, denomAdym)

	denomAphoton := createDenomMetadata("aphoton", "PHOTON", 18)
	suite.Require().Nil(denomAphoton.Validate(), "bad setup")
	suite.App.BankKeeper.SetDenomMetaData(suite.Ctx, denomAphoton)

	denomIbcAtom := createDenomMetadata("ibc/uatom", "ATOM", 6)
	suite.Require().Nil(denomIbcAtom.Validate(), "bad setup")
	suite.App.BankKeeper.SetDenomMetaData(suite.Ctx, denomIbcAtom)

	denomIbcOsmo := createDenomMetadata("ibc/uosmo", "OSMO", 6)
	suite.Require().Nil(denomIbcOsmo.Validate(), "bad setup")
	suite.App.BankKeeper.SetDenomMetaData(suite.Ctx, denomIbcOsmo)

	const epochIdentifier = "day"

	hooks := extendedevmkeeper.NewEvmEpochHooks(*suite.App.EvmKeeper, suite.App.BankKeeper)

	suite.Ctx = suite.Ctx.WithBlockHeight(0) // ignore invoking x/evm exec for contract deployment

	/* ------------------------- call in-correct epoch hook ------------------------- */
	err := hooks.BeforeEpochStart(suite.Ctx, "month", 0)
	suite.Require().NoError(err)

	_, foundAdym := suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomAdym.Base)
	suite.Require().False(foundAdym)
	_, foundAphoton := suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomAphoton.Base)
	suite.Require().False(foundAphoton)
	_, foundIbcAtom := suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomIbcAtom.Base)
	suite.Require().False(foundIbcAtom)
	_, foundIbcOsmo := suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomIbcOsmo.Base)
	suite.Require().False(foundIbcOsmo)

	/* ------------------------- call corresponding epoch hook ------------------------- */
	err = hooks.BeforeEpochStart(suite.Ctx, epochIdentifier, 0)
	suite.Require().NoError(err)

	_, found := suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomAdym.Base)
	suite.Require().False(found, "only deploy VFBC for IBC denom")

	_, found = suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomAphoton.Base)
	suite.Require().False(found, "only deploy VFBC for IBC denom")

	originalVfbcIbcAtomAddr, found := suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomIbcAtom.Base)
	suite.Require().True(found)
	suite.Require().NotEqual(common.Address{}, originalVfbcIbcAtomAddr)

	originalVfbcIbcOsmoAddr, found := suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomIbcOsmo.Base)
	suite.Require().True(found)
	suite.Require().NotEqual(common.Address{}, originalVfbcIbcOsmoAddr)

	originalVfbcIbcAtom := suite.App.EvmKeeper.GetVirtualFrontierContract(suite.Ctx, originalVfbcIbcAtomAddr)
	suite.Require().NotNil(originalVfbcIbcAtom)

	originalVfbcIbcOsmo := suite.App.EvmKeeper.GetVirtualFrontierContract(suite.Ctx, originalVfbcIbcOsmoAddr)
	suite.Require().NotNil(originalVfbcIbcOsmo)

	suite.Require().NotEqual(originalVfbcIbcAtomAddr, originalVfbcIbcOsmoAddr, "contract addresses must be different")

	/* ------------------------- call epoch hook again ------------------------- */

	err = hooks.BeforeEpochStart(suite.Ctx, epochIdentifier, 0)
	suite.Require().NoError(err)

	_, found = suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomAdym.Base)
	suite.Require().False(found, "only deploy VFBC for IBC denom")

	_, found = suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomAphoton.Base)
	suite.Require().False(found, "only deploy VFBC for IBC denom")

	vfbcIbcAtomAddr, found := suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomIbcAtom.Base)
	suite.Require().True(found)
	suite.Require().Equal(originalVfbcIbcAtomAddr, vfbcIbcAtomAddr, "address must not be changed")

	vfbcIbcOsmoAddr, found := suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomIbcOsmo.Base)
	suite.Require().True(found)
	suite.Require().Equal(originalVfbcIbcOsmoAddr, vfbcIbcOsmoAddr, "address must not be changed")

	vfbcIbcAtom := suite.App.EvmKeeper.GetVirtualFrontierContract(suite.Ctx, originalVfbcIbcAtomAddr)
	suite.Require().NotNil(vfbcIbcAtom)
	suite.Require().Equal(originalVfbcIbcAtom, vfbcIbcAtom, "content must not be changed")

	vfbcIbcOsmo := suite.App.EvmKeeper.GetVirtualFrontierContract(suite.Ctx, originalVfbcIbcOsmoAddr)
	suite.Require().NotNil(vfbcIbcOsmo)
	suite.Require().Equal(originalVfbcIbcOsmo, vfbcIbcOsmo, "content must not be changed")

	/* ------------------------- call epoch hook again with different epoch identifier ------------------------- */

	err = hooks.BeforeEpochStart(suite.Ctx, "month", 0)
	suite.Require().NoError(err)

	_, found = suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomAdym.Base)
	suite.Require().False(found, "only deploy VFBC for IBC denom")

	_, found = suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomAphoton.Base)
	suite.Require().False(found, "only deploy VFBC for IBC denom")

	vfbcIbcAtomAddr, found = suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomIbcAtom.Base)
	suite.Require().True(found)
	suite.Require().Equal(originalVfbcIbcAtomAddr, vfbcIbcAtomAddr, "address must not be changed")

	vfbcIbcOsmoAddr, found = suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomIbcOsmo.Base)
	suite.Require().True(found)
	suite.Require().Equal(originalVfbcIbcOsmoAddr, vfbcIbcOsmoAddr, "address must not be changed")

	vfbcIbcAtom = suite.App.EvmKeeper.GetVirtualFrontierContract(suite.Ctx, originalVfbcIbcAtomAddr)
	suite.Require().NotNil(vfbcIbcAtom)
	suite.Require().Equal(originalVfbcIbcAtom, vfbcIbcAtom, "content must not be changed")

	vfbcIbcOsmo = suite.App.EvmKeeper.GetVirtualFrontierContract(suite.Ctx, originalVfbcIbcOsmoAddr)
	suite.Require().NotNil(vfbcIbcOsmo)
	suite.Require().Equal(originalVfbcIbcOsmo, vfbcIbcOsmo, "content must not be changed")

	/* ------------------------- deploy new contract ------------------------- */

	denomIbcEvmos := createDenomMetadata("ibc/aevmos", "EVMOS", 18)
	suite.Require().Nil(denomIbcEvmos.Validate(), "bad setup")
	suite.App.BankKeeper.SetDenomMetaData(suite.Ctx, denomIbcEvmos)

	err = hooks.BeforeEpochStart(suite.Ctx, epochIdentifier, 0)
	suite.Require().NoError(err)

	vfbcIbcAtomAddr, found = suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomIbcAtom.Base)
	suite.Require().True(found)

	vfbcIbcOsmoAddr, found = suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomIbcOsmo.Base)
	suite.Require().True(found)

	vfbcIbcEvmosAddr, found := suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(suite.Ctx, denomIbcEvmos.Base)
	suite.Require().True(found)

	vfbcIbcAtom = suite.App.EvmKeeper.GetVirtualFrontierContract(suite.Ctx, originalVfbcIbcAtomAddr)
	suite.Require().NotNil(vfbcIbcAtom)
	suite.Require().Equal(originalVfbcIbcAtom, vfbcIbcAtom, "content must not be changed")

	vfbcIbcOsmo = suite.App.EvmKeeper.GetVirtualFrontierContract(suite.Ctx, originalVfbcIbcOsmoAddr)
	suite.Require().NotNil(vfbcIbcOsmo)
	suite.Require().Equal(originalVfbcIbcOsmo, vfbcIbcOsmo, "content must not be changed")

	vfbcIbcEvmos := suite.App.EvmKeeper.GetVirtualFrontierContract(suite.Ctx, vfbcIbcEvmosAddr)
	suite.Require().NotNil(vfbcIbcEvmos)

	suite.Require().NotEqual(vfbcIbcAtomAddr, vfbcIbcEvmosAddr, "contract addresses must be different")
	suite.Require().NotEqual(vfbcIbcOsmoAddr, vfbcIbcEvmosAddr, "contract addresses must be different")

	/* ------------------------- ensure contract content ------------------------- */

	accountContractIbcAtom := suite.App.EvmKeeper.GetAccount(suite.Ctx, vfbcIbcAtomAddr)
	suite.Require().NotNil(accountContractIbcAtom)
	suite.Require().NotEmpty(accountContractIbcAtom.CodeHash)
	suite.Equal(uint64(1), accountContractIbcAtom.Nonce)

	accountContractIbcOsmo := suite.App.EvmKeeper.GetAccount(suite.Ctx, vfbcIbcOsmoAddr)
	suite.Require().NotNil(accountContractIbcOsmo)
	suite.Require().NotEmpty(accountContractIbcOsmo.CodeHash)
	suite.Equal(uint64(1), accountContractIbcOsmo.Nonce)

	accountContractIbcEvmos := suite.App.EvmKeeper.GetAccount(suite.Ctx, vfbcIbcEvmosAddr)
	suite.Require().NotNil(accountContractIbcEvmos)
	suite.Require().NotEmpty(accountContractIbcEvmos.CodeHash)
	suite.Equal(uint64(1), accountContractIbcEvmos.Nonce)

	if suite.Equal(accountContractIbcAtom.CodeHash, accountContractIbcOsmo.CodeHash) &&
		suite.Equal(accountContractIbcAtom.CodeHash, accountContractIbcEvmos.CodeHash) {
		// Currently DYM does not use EthAccount is the default proto account so the following test won't work.
		// TODO: In the future, if enable contract creation for DYM, need to migrate all the VFBC to use EthAccount and set the corresponding code hash and code, see https://github.com/VictorTrustyDev/fork-dym-ethermint/blob/0baa0fd7c351cf91356788450acb20e42ebd6366/x/evm/keeper/virtual_frontier_contract.go#L289-L290
		/*
			suite.NotEmpty(suite.App.EvmKeeper.GetCode(suite.Ctx, common.BytesToHash(accountContractIbcAtom.CodeHash)))
			suite.NotEmpty(suite.App.EvmKeeper.GetCode(suite.Ctx, common.BytesToHash(accountContractIbcOsmo.CodeHash)))
			suite.NotEmpty(suite.App.EvmKeeper.GetCode(suite.Ctx, common.BytesToHash(accountContractIbcEvmos.CodeHash)))

			expectedCodeHash := []byte{242, 212, 238, 130, 251, 7, 113, 156, 203, 245, 205, 60, 34, 40, 221, 251, 24, 166, 103, 255, 59, 173, 19, 34, 43, 190, 198, 94, 101, 40, 4, 154}
			if !suite.Equal(expectedCodeHash, accountContractIbcAtom.CodeHash) {
				suite.NotEqual(crypto.Keccak256(nil), accountContractIbcAtom.CodeHash, "code hash must not be empty")
			}
		*/
	}
}

func createDenomMetadata(base, display string, decimals uint32) banktypes.Metadata {
	return banktypes.Metadata{
		Description: display,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    base,
				Exponent: 0,
			},
			{
				Denom:    display,
				Exponent: decimals,
			},
		},
		Base:    base,
		Display: display,
		Name:    display,
		Symbol:  display,
	}
}
