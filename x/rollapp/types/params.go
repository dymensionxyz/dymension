package types

import (
	"errors"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	// KeyRollappsEnabled is store's key for RollappsEnabled Params
	KeyRollappsEnabled = []byte("RollappsEnabled")
	// KeyDeployerWhitelist is store's key for DeployerWhitelist Params
	KeyDeployerWhitelist = []byte("DeployerWhitelist")
	// KeyDisputePeriodInBlocks is store's key for DisputePeriodInBlocks Params
	KeyDisputePeriodInBlocks = []byte("DisputePeriodInBlocks")
	// KeyDisputePeriodTransferGenesisInBlocks is store's key for DisputePeriodInBlocks Params
	KeyDisputePeriodTransferGenesisInBlocks            = []byte("DisputePeriodTransferGenesisInBlocks")
	DefaultDisputePeriodInBlocks                uint64 = 3
	DefaultDisputePeriodInBlocksTransferGenesis uint64 = 20 // TODO:
	// MinDisputePeriodInBlocks is the minimum number of blocks for dispute period
	MinDisputePeriodInBlocks uint64 = 1
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	enabled bool,
	disputePeriodInBlocks uint64,
	disputePeriodInBlocksTransferGenesis uint64,
	deployerWhitelist []DeployerParams,
) Params {
	return Params{
		DisputePeriodInBlocks:                disputePeriodInBlocks,
		TransferGenesisDisputePeriodInBlocks: disputePeriodInBlocksTransferGenesis,
		DeployerWhitelist:                    deployerWhitelist,
		RollappsEnabled:                      enabled,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		true, DefaultDisputePeriodInBlocks, DefaultDisputePeriodInBlocksTransferGenesis, []DeployerParams{},
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDisputePeriodInBlocks, &p.DisputePeriodInBlocks, validateDisputePeriodInBlocks),
		paramtypes.NewParamSetPair(KeyDisputePeriodTransferGenesisInBlocks, &p.TransferGenesisDisputePeriodInBlocks, validateDisputePeriodTransferGenesisInBlocks),
		paramtypes.NewParamSetPair(KeyDeployerWhitelist, &p.DeployerWhitelist, validateDeployerWhitelist),
		paramtypes.NewParamSetPair(KeyRollappsEnabled, &p.RollappsEnabled, func(_ interface{}) error { return nil }),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateDisputePeriodInBlocks(p.DisputePeriodInBlocks); err != nil {
		return err
	}
	if err := validateDisputePeriodTransferGenesisInBlocks(p.DisputePeriodInBlocks); err != nil {
		return err
	}

	return validateDeployerWhitelist(p.DeployerWhitelist)
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// validateDisputePeriodInBlocks validates the DisputePeriodInBlocks param
func validateDisputePeriodInBlocks(v interface{}) error {
	disputePeriodInBlocks, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if disputePeriodInBlocks < MinDisputePeriodInBlocks {
		return errors.New("dispute period cannot be lower than 1 block")
	}

	return nil
}

func validateDisputePeriodTransferGenesisInBlocks(v interface{}) error {
	_, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("wrong type: %T", v)
	}

	return nil
}

// validateDeployerWhitelist validates the DeployerWhitelist param
func validateDeployerWhitelist(v interface{}) error {
	deployerWhitelist, ok := v.([]DeployerParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// Check for duplicated index in deployer address
	rollappDeployerIndexMap := make(map[string]struct{})

	for i, item := range deployerWhitelist {
		// check Bech32 format
		if _, err := sdk.AccAddressFromBech32(item.Address); err != nil {
			return fmt.Errorf("deployerWhitelist[%d] format error: %s", i, err.Error())
		}

		// check duplicate
		if _, ok := rollappDeployerIndexMap[item.Address]; ok {
			return errors.New("duplicated deployer address in deployerWhitelist")
		}
		rollappDeployerIndexMap[item.Address] = struct{}{}
	}

	return nil
}
