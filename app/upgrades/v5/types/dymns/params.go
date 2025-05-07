package dymns

import (
	"fmt"
	"time"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

const (
	ModuleName = "dymns"
)

// Params defines the parameters for the dymns module
type Params struct {
	Price  PriceParams  `json:"price" yaml:"price"`
	Chains ChainsParams `json:"chains" yaml:"chains"`
	Misc   MiscParams   `json:"misc" yaml:"misc"`
}

// PriceParams defines the price parameters
type PriceParams struct {
	PriceDenom string `json:"price_denom" yaml:"price_denom"`
}

// ChainsParams defines the chains parameters
type ChainsParams struct {
	AliasesOfChainIds []AliasesOfChainId `json:"aliases_of_chain_ids" yaml:"aliases_of_chain_ids"`
}

// AliasesOfChainId defines the aliases for a chain ID
type AliasesOfChainId struct {
	ChainId string   `json:"chain_id" yaml:"chain_id"`
	Aliases []string `json:"aliases" yaml:"aliases"`
}

// MiscParams defines the miscellaneous parameters
type MiscParams struct {
	EndEpochHookIdentifier string        `json:"end_epoch_hook_identifier" yaml:"end_epoch_hook_identifier"`
	GracePeriodDuration    time.Duration `json:"grace_period_duration" yaml:"grace_period_duration"`
	SellOrderDuration      time.Duration `json:"sell_order_duration" yaml:"sell_order_duration"`
}

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	// KeyPriceParams is the key for the price params
	KeyPriceParams = []byte("PriceParams")

	// KeyChainsParams is the key for the chains params
	KeyChainsParams = []byte("ChainsParams")

	// KeyMiscParams is the key for the misc params
	KeyMiscParams = []byte("MiscParams")
)

const (
	defaultEndEpochHookIdentifier = "hour"
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(price PriceParams, chains ChainsParams, misc MiscParams) Params {
	return Params{
		Price:  price,
		Chains: chains,
		Misc:   misc,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultPriceParams(),
		DefaultChainsParams(),
		DefaultMiscParams(),
	)
}

// DefaultPriceParams returns a default set of price parameters
func DefaultPriceParams() PriceParams {
	return PriceParams{
		PriceDenom: "adym",
	}
}

// DefaultChainsParams returns a default set of chains parameters
func DefaultChainsParams() ChainsParams {
	return ChainsParams{
		AliasesOfChainIds: []AliasesOfChainId{},
	}
}

// DefaultMiscParams returns a default set of miscellaneous parameters
func DefaultMiscParams() MiscParams {
	return MiscParams{
		EndEpochHookIdentifier: defaultEndEpochHookIdentifier,
		GracePeriodDuration:    24 * time.Hour,
		SellOrderDuration:      24 * time.Hour,
	}
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyPriceParams, &p.Price, validatePriceParams),
		paramtypes.NewParamSetPair(KeyChainsParams, &p.Chains, validateChainsParams),
		paramtypes.NewParamSetPair(KeyMiscParams, &p.Misc, validateMiscParams),
	}
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validatePriceParams(p.Price); err != nil {
		return err
	}
	if err := validateChainsParams(p.Chains); err != nil {
		return err
	}
	if err := validateMiscParams(p.Misc); err != nil {
		return err
	}
	return nil
}

func validatePriceParams(i interface{}) error {
	v, ok := i.(PriceParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.PriceDenom == "" {
		return fmt.Errorf("price denom cannot be empty")
	}
	return nil
}

func validateChainsParams(i interface{}) error {
	v, ok := i.(ChainsParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	for _, chain := range v.AliasesOfChainIds {
		if chain.ChainId == "" {
			return fmt.Errorf("chain ID cannot be empty")
		}
		for _, alias := range chain.Aliases {
			if alias == "" {
				return fmt.Errorf("alias cannot be empty")
			}
		}
	}
	return nil
}

func validateMiscParams(i interface{}) error {
	v, ok := i.(MiscParams)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.EndEpochHookIdentifier == "" {
		return fmt.Errorf("end epoch hook identifier cannot be empty")
	}
	if v.GracePeriodDuration <= 0 {
		return fmt.Errorf("grace period duration must be positive")
	}
	if v.SellOrderDuration <= 0 {
		return fmt.Errorf("sell order duration must be positive")
	}
	return nil
}
