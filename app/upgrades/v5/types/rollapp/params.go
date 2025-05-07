package rollapp

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

const (
	ModuleName = "rollapp"
)

// Params defines the parameters for the rollapp module
type Params struct {
	DisputePeriodInBlocks  uint64   `json:"dispute_period_in_blocks" yaml:"dispute_period_in_blocks"`
	LivenessSlashBlocks    uint64   `json:"liveness_slash_blocks" yaml:"liveness_slash_blocks"`
	LivenessSlashInterval  uint64   `json:"liveness_slash_interval" yaml:"liveness_slash_interval"`
	AppRegistrationFee     sdk.Coin `json:"app_registration_fee" yaml:"app_registration_fee"`
	MinSequencerBondGlobal sdk.Coin `json:"min_sequencer_bond_global" yaml:"min_sequencer_bond_global"`
}

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	// KeyDisputePeriodInBlocks is store's key for DisputePeriodInBlocks Params
	KeyDisputePeriodInBlocks = []byte("DisputePeriodInBlocks")

	KeyLivenessSlashBlocks   = []byte("LivenessSlashBlocks")
	KeyLivenessSlashInterval = []byte("LivenessSlashInterval")

	// KeyAppRegistrationFee defines the key to store the cost of the app
	KeyAppRegistrationFee = []byte("AppRegistrationFee")

	KeyMinSequencerBondGlobal = []byte("KeyMinSequencerBondGlobal")
)

const (
	DefaultDisputePeriodInBlocks uint64 = 3
	// MinDisputePeriodInBlocks is the minimum number of blocks for dispute period
	MinDisputePeriodInBlocks uint64 = 1

	DefaultLivenessSlashBlocks   = uint64(7200) // 12 hours worth of blocks at 1 block per 6 seconds
	DefaultLivenessSlashInterval = uint64(600)  // 1 hour worth of blocks at 1 block per 6 seconds
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	disputePeriodInBlocks uint64,
	livenessSlashBlocks uint64,
	livenessSlashInterval uint64,
	appRegistrationFee sdk.Coin,
	minSequencerBondGlobal sdk.Coin,
) Params {
	return Params{
		DisputePeriodInBlocks:  disputePeriodInBlocks,
		LivenessSlashBlocks:    livenessSlashBlocks,
		LivenessSlashInterval:  livenessSlashInterval,
		AppRegistrationFee:     appRegistrationFee,
		MinSequencerBondGlobal: minSequencerBondGlobal,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	minBond, ok := math.NewIntFromString("100000000000000000000")
	if !ok {
		panic("failed to parse min sequencer bond amount")
	}
	return NewParams(
		DefaultDisputePeriodInBlocks,
		DefaultLivenessSlashBlocks,
		DefaultLivenessSlashInterval,
		sdk.NewCoin("adym", math.NewInt(1000000000000000000)), // 1 DYM
		sdk.NewCoin("adym", minBond),                          // 100 DYM
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDisputePeriodInBlocks, &p.DisputePeriodInBlocks, validateDisputePeriodInBlocks),
		paramtypes.NewParamSetPair(KeyLivenessSlashBlocks, &p.LivenessSlashBlocks, validateLivenessSlashBlocks),
		paramtypes.NewParamSetPair(KeyLivenessSlashInterval, &p.LivenessSlashInterval, validateLivenessSlashInterval),
		paramtypes.NewParamSetPair(KeyAppRegistrationFee, &p.AppRegistrationFee, validateAppRegistrationFee),
		paramtypes.NewParamSetPair(KeyMinSequencerBondGlobal, &p.MinSequencerBondGlobal, validateMinSequencerBondGlobal),
	}
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateDisputePeriodInBlocks(p.DisputePeriodInBlocks); err != nil {
		return err
	}
	if err := validateLivenessSlashBlocks(p.LivenessSlashBlocks); err != nil {
		return err
	}
	if err := validateLivenessSlashInterval(p.LivenessSlashInterval); err != nil {
		return err
	}
	if err := validateAppRegistrationFee(p.AppRegistrationFee); err != nil {
		return err
	}
	if err := validateMinSequencerBondGlobal(p.MinSequencerBondGlobal); err != nil {
		return err
	}
	return nil
}

func validateDisputePeriodInBlocks(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v < MinDisputePeriodInBlocks {
		return fmt.Errorf("dispute period in blocks must be at least %d", MinDisputePeriodInBlocks)
	}
	return nil
}

func validateLivenessSlashBlocks(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("liveness slash blocks cannot be zero")
	}
	return nil
}

func validateLivenessSlashInterval(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("liveness slash interval cannot be zero")
	}
	return nil
}

func validateAppRegistrationFee(i interface{}) error {
	v, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.Amount.IsNegative() {
		return fmt.Errorf("app registration fee cannot be negative")
	}
	return nil
}

func validateMinSequencerBondGlobal(i interface{}) error {
	v, ok := i.(sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.Amount.IsNegative() {
		return fmt.Errorf("min sequencer bond global cannot be negative")
	}
	return nil
}
