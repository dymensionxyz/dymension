package types

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
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

	KeyLivenessSlashBlocks   = []byte("LivenessSlashBlocks")
	KeyLivenessSlashInterval = []byte("LivenessSlashInterval")
	KeyLivenessJailBlocks    = []byte("LivenessJailBlocks")
)

const (
	DefaultDisputePeriodInBlocks uint64 = 3
	// MinDisputePeriodInBlocks is the minimum number of blocks for dispute period
	MinDisputePeriodInBlocks uint64 = 1

	DefaultLivenessSlashBlocks   = uint64(7200)  // 12 hours at 6 blocks per second
	DefaultLivenessSlashInterval = uint64(3600)  // 1 hour at 6 blocks per second
	DefaultLivenessJailBlocks    = uint64(28800) // 48 hours at 6 blocks per second
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	enabled bool,
	disputePeriodInBlocks uint64,
	deployerWhitelist []DeployerParams,
	livenessSlashBlocks uint64,
	livenessSlashInterval uint64,
	livenessJailBlocks uint64,
) Params {
	return Params{
		DisputePeriodInBlocks: disputePeriodInBlocks,
		DeployerWhitelist:     deployerWhitelist,
		RollappsEnabled:       enabled,
		LivenessSlashBlocks:   livenessSlashBlocks,
		LivenessSlashInterval: livenessSlashInterval,
		LivenessJailBlocks:    livenessJailBlocks,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		true,
		DefaultDisputePeriodInBlocks,
		[]DeployerParams{},
		DefaultLivenessSlashBlocks,
		DefaultLivenessSlashInterval,
		DefaultLivenessJailBlocks,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDisputePeriodInBlocks, &p.DisputePeriodInBlocks, validateDisputePeriodInBlocks),
		paramtypes.NewParamSetPair(KeyDeployerWhitelist, &p.DeployerWhitelist, validateDeployerWhitelist),
		paramtypes.NewParamSetPair(KeyRollappsEnabled, &p.RollappsEnabled, func(_ interface{}) error { return nil }),
		paramtypes.NewParamSetPair(KeyLivenessSlashBlocks, &p.LivenessSlashBlocks, validateLivenessSlashBlocks),
		paramtypes.NewParamSetPair(KeyLivenessSlashInterval, &p.LivenessSlashInterval, validateLivenessSlashInterval),
		paramtypes.NewParamSetPair(KeyLivenessJailBlocks, &p.LivenessJailBlocks, validateLivenessJailBlocks),
	}
}

func (p Params) WithDeployerWhitelist(l []DeployerParams) Params {
	p.DeployerWhitelist = l
	return p
}

func (p Params) WithDisputePeriodInBlocks(x uint64) Params {
	p.DisputePeriodInBlocks = x
	return p
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateDisputePeriodInBlocks(p.DisputePeriodInBlocks); err != nil {
		return errorsmod.Wrap(err, "dispute period")
	}

	if err := validateLivenessSlashBlocks(p.LivenessSlashBlocks); err != nil {
		return errorsmod.Wrap(err, "liveness slash blocks")
	}
	if err := validateLivenessSlashInterval(p.LivenessSlashInterval); err != nil {
		return errorsmod.Wrap(err, "liveness slash interval")
	}
	if err := validateLivenessJailBlocks(p.LivenessJailBlocks); err != nil {
		return errorsmod.Wrap(err, "liveness jail blocks")
	}

	return validateDeployerWhitelist(p.DeployerWhitelist)
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateLivenessSlashBlocks(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v <= 0 {
		return fmt.Errorf("must be positive: %d", v)
	}
	return nil
}

func validateLivenessSlashInterval(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v <= 0 {
		return fmt.Errorf("must be positive: %d", v)
	}
	return nil
}

func validateLivenessJailBlocks(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v <= 0 {
		return fmt.Errorf("must be positive: %d", v)
	}
	return nil
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
