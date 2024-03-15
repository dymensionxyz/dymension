package keeper_test

import (
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	denommetadatamodulekeeper "github.com/dymensionxyz/dymension/v3/x/denommetadata/keeper"
	"github.com/ethereum/go-ethereum/common"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

func (suite *KeeperTestSuite) TestHookOperation_AfterDenomMetadataCreation() {
	suite.SetupTest()

	denomAdym := createDenomMetadata("adym", "DYM", 18)
	suite.Require().Nil(denomAdym.Validate(), "bad setup")

	denomWei := createDenomMetadata("wei", "ETH", 18)
	suite.Require().Nil(denomWei.Validate(), "bad setup")

	denomIbcAtom := createDenomMetadata("ibc/uatom", "ATOM", 6)
	suite.Require().Nil(denomIbcAtom.Validate(), "bad setup")

	denomIbcOsmo := createDenomMetadata("ibc/uosmo", "OSMO", 6)
	suite.Require().Nil(denomIbcOsmo.Validate(), "bad setup")

	denomMixedCaseIbcTia := createDenomMetadata("IbC/utIa", "TiA", 6)
	suite.Require().Nil(denomMixedCaseIbcTia.Validate(), "bad setup")

	suite.App.BankKeeper.SetDenomMetaData(suite.Ctx, denomAdym)
	suite.App.BankKeeper.SetDenomMetaData(suite.Ctx, denomWei)
	suite.App.BankKeeper.SetDenomMetaData(suite.Ctx, denomIbcAtom)
	suite.App.BankKeeper.SetDenomMetaData(suite.Ctx, denomIbcOsmo)
	suite.App.BankKeeper.SetDenomMetaData(suite.Ctx, denomMixedCaseIbcTia)

	denomIbcNotExists := createDenomMetadata("ibc/not_exists", "NE", 9)

	hooks := denommetadatamodulekeeper.NewVirtualFrontierBankContractRegistrationHook(*suite.App.EvmKeeper)

	suite.Ctx = suite.Ctx.WithBlockHeight(0) // ignore invoking x/evm exec for contract deployment

	tests := []struct {
		name          string
		denomMetadata banktypes.Metadata
		wantFound     []string
		wantNotFound  []string
	}{
		{
			name:          "ignored - Deploy for non-IBC",
			denomMetadata: denomAdym,
			wantNotFound:  []string{denomAdym.Base, denomWei.Base, denomIbcAtom.Base, denomIbcOsmo.Base, denomMixedCaseIbcTia.Base},
		},
		{
			name:          "ignored - Only deploy the denom which passed into the hook (case non-IBC)",
			denomMetadata: denomWei,
			wantNotFound:  []string{denomAdym.Base, denomWei.Base, denomIbcAtom.Base, denomIbcOsmo.Base, denomMixedCaseIbcTia.Base},
		},
		{
			name:          "accepted - Only deploy the denom which passed into the hook (case IBC)",
			denomMetadata: denomIbcAtom,
			wantFound:     []string{denomIbcAtom.Base},
			wantNotFound:  []string{denomAdym.Base, denomWei.Base, denomIbcOsmo.Base, denomMixedCaseIbcTia.Base},
		},
		{
			name:          "accepted - Only deploy the denom which passed into the hook (case IBC)",
			denomMetadata: denomIbcOsmo,
			wantFound:     []string{denomIbcOsmo.Base},
			wantNotFound:  []string{denomAdym.Base, denomWei.Base, denomIbcAtom.Base, denomMixedCaseIbcTia.Base},
		},
		{
			name:          "accepted - Only deploy the denom which passed into the hook (case IBC)",
			denomMetadata: denomMixedCaseIbcTia,
			wantFound:     []string{denomMixedCaseIbcTia.Base},
			wantNotFound:  []string{denomAdym.Base, denomWei.Base, denomIbcAtom.Base, denomIbcOsmo.Base},
		},
		{
			name:          "ignored - Do not deploy the metadata which does not exists in bank",
			denomMetadata: denomIbcNotExists,
			wantNotFound:  []string{denomIbcNotExists.Base},
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			workingCtx, _ := suite.Ctx.CacheContext() // clone the ctx so no need to reset the ctx after each test case

			err := hooks.AfterDenomMetadataCreation(workingCtx, tt.denomMetadata)
			suite.Require().NoError(err, "should not be error in any case")

			if len(tt.wantFound) > 0 {
				for _, wantFound := range tt.wantFound {
					_, found := suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(workingCtx, wantFound)
					suite.Truef(found, "VFBC for %s must be found", wantFound)
				}
			}

			if len(tt.wantNotFound) > 0 {
				for _, wantNotFound := range tt.wantNotFound {
					_, found := suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(workingCtx, wantNotFound)
					suite.Falsef(found, "VFBC for %s must not be found", wantNotFound)
				}
			}

			newContractAddr1, found := suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(workingCtx, tt.denomMetadata.Base)
			if found {
				// perform double deployment check
				err := hooks.AfterDenomMetadataCreation(workingCtx, tt.denomMetadata)
				suite.Require().NoError(err, "should not be error in any case")

				// ensure the contract address is not changed
				newContractAddr2, found := suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(workingCtx, tt.denomMetadata.Base)
				suite.Require().True(found, "contract disappeared?")
				suite.Equal(newContractAddr1, newContractAddr2, "contract address must not be changed")
			}
		})
	}

	suite.Run("Double deployment call does not effective previous contract", func() {
		workingCtx, _ := suite.Ctx.CacheContext() // clone the ctx so no need to reset the ctx after each test case

		_, found := suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(workingCtx, denomIbcAtom.Base)
		suite.Require().False(found)

		err := hooks.AfterDenomMetadataCreation(workingCtx, denomIbcAtom)
		suite.Require().NoError(err, "should not be error in any case")

		newContractAddr1, found := suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(workingCtx, denomIbcAtom.Base)
		suite.Require().True(found, "contract must be deployed")

		// re-deploy
		err = hooks.AfterDenomMetadataCreation(workingCtx, denomIbcAtom)
		suite.Require().NoError(err, "should not be error in any case")

		// ensure the contract address is not changed
		newContractAddr2, found := suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(workingCtx, denomIbcAtom.Base)
		suite.Require().True(found, "contract disappeared?")
		suite.Equal(newContractAddr1, newContractAddr2, "contract address must not be changed")
	})

	suite.Run("Contract state must be set", func() {
		workingCtx, _ := suite.Ctx.CacheContext() // clone the ctx so no need to reset the ctx after each test case

		for _, metadata := range []banktypes.Metadata{denomIbcAtom, denomIbcOsmo} {
			contractAddr, found := suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(workingCtx, metadata.Base)
			suite.Require().False(found, "contract must not be exists")

			err := hooks.AfterDenomMetadataCreation(workingCtx, metadata)
			suite.Require().NoError(err, "should not be error in any case")

			contractAddr, found = suite.App.EvmKeeper.GetVirtualFrontierBankContractAddressByDenom(workingCtx, metadata.Base)
			suite.Require().True(found, "contract must be exists")

			accountContract := suite.App.EvmKeeper.GetAccount(workingCtx, contractAddr)
			suite.Require().NotNil(accountContract)
			suite.Require().NotEmpty(accountContract.CodeHash)
			suite.Equal(uint64(1), accountContract.Nonce)
			suite.Equal(evmtypes.VFBCCodeHash, accountContract.CodeHash)
			suite.Equal(evmtypes.VFBCCode, suite.App.EvmKeeper.GetCode(workingCtx, common.BytesToHash(accountContract.CodeHash)))
		}
	})
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
