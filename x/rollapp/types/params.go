package types

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"

	appparams "github.com/dymensionxyz/dymension/v3/app/params"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	// KeyAliasFeeTable is store's key for AliasFeeTable Params
	KeyAliasFeeTable = []byte("AliasFeeTable")
	// DYM is the integer representation of 1 DYM
	DYM = sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
	// DefaultAliasFeeTable is the default registration fee
	// only string keys allowed for params map
	DefaultAliasFeeTable = map[string]sdk.Coin{
		"7": sdk.NewCoin(appparams.BaseDenom, DYM.MulRaw(5)),
		"6": sdk.NewCoin(appparams.BaseDenom, DYM.MulRaw(10)),
		"5": sdk.NewCoin(appparams.BaseDenom, DYM.MulRaw(25)),
		"4": sdk.NewCoin(appparams.BaseDenom, DYM.MulRaw(100)),
		"3": sdk.NewCoin(appparams.BaseDenom, DYM.MulRaw(250)),
		"2": sdk.NewCoin(appparams.BaseDenom, DYM.MulRaw(1000)),
		"1": sdk.NewCoin(appparams.BaseDenom, DYM.MulRaw(5000)),
	}
	// KeyDisputePeriodInBlocks is store's key for DisputePeriodInBlocks Params
	KeyDisputePeriodInBlocks            = []byte("DisputePeriodInBlocks")
	DefaultDisputePeriodInBlocks uint64 = 3
	// MinDisputePeriodInBlocks is the minimum number of blocks for dispute period
	MinDisputePeriodInBlocks uint64 = 1
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	disputePeriodInBlocks uint64,
	aliasFeePricingTable map[string]sdk.Coin,
) Params {
	return Params{
		DisputePeriodInBlocks: disputePeriodInBlocks,
		AliasFeeTable:         aliasFeePricingTable,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultDisputePeriodInBlocks, DefaultAliasFeeTable)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDisputePeriodInBlocks, &p.DisputePeriodInBlocks, validateDisputePeriodInBlocks),
		paramtypes.NewParamSetPair(KeyAliasFeeTable, &p.AliasFeeTable, validateAliasFeeTable),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateDisputePeriodInBlocks(p.DisputePeriodInBlocks); err != nil {
		return err
	}

	return validateAliasFeeTable(p.AliasFeeTable)
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

// validateAliasFeeTable validates the AliasFeeTable param
func validateAliasFeeTable(v interface{}) error {
	aliasPricingTable, ok := v.(map[string]sdk.Coin)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if err := checkIfConsecutiveAliasLengths(aliasPricingTable); err != nil {
		return err
	}

	for aliasLengthStr, fee := range aliasPricingTable {
		aliasLength, err := strconv.ParseInt(aliasLengthStr, 10, 32)
		if err != nil {
			return fmt.Errorf("invalid alias length: %s", aliasLengthStr)
		}

		if aliasLength < 1 {
			return errors.New("alias length must be at least 1")
		}

		if !fee.IsValid() {
			return errors.New("invalid fee")
		}

		if fee.Denom != appparams.BaseDenom {
			return errors.New("fee denom must be DYM")
		}
	}

	return nil
}

func checkIfConsecutiveAliasLengths(aliasPricingTable map[string]sdk.Coin) error {
	for i := 1; i <= len(aliasPricingTable); i++ {
		if _, ok := aliasPricingTable[fmt.Sprint(i)]; !ok {
			return fmt.Errorf("missing alias length: %d", i)
		}
	}
	return nil
}
