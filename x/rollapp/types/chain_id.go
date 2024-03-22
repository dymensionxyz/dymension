package types

import (
	"fmt"
	"math/big"
	"regexp"
	"strings"

	errorsmod "cosmossdk.io/errors"
)

var (
	regexChainID         = `[a-z]{1,}`
	regexEIP155Separator = `_{1}`
	regexEIP155          = `[1-9][0-9]*`
	regexEpochSeparator  = `-{1}`
	regexEpoch           = `[1-9][0-9]*`
	ethermintChainID     = regexp.MustCompile(fmt.Sprintf(`^(%s)%s(%s)%s(%s)$`,
		regexChainID,
		regexEIP155Separator,
		regexEIP155,
		regexEpochSeparator,
		regexEpoch))
)

type EIP155ChainID struct {
	ChainID   *big.Int
	ChainName string
}

func NewEIP155ChainID(chainID string) (EIP155ChainID, error) {
	chainID = strings.TrimSpace(chainID)

	eip155ID, err := createEIP155ChainID(chainID)
	if err != nil {
		return EIP155ChainID{}, err
	}
	return EIP155ChainID{
		ChainID:   eip155ID,
		ChainName: chainID,
	}, err
}

// GetValidEIP155ChainId parses a string chain identifier's epoch to an Ethereum-compatible
// chain-id in *big.Int format. The function returns an error if the chain-id has an invalid format
func createEIP155ChainID(chainID string) (*big.Int, error) {
	if chainID == "" {
		return nil, errorsmod.Wrapf(ErrInvalidRollappID, "chain-id cannot be empty")
	}

	if len(chainID) > 48 {
		return nil, errorsmod.Wrapf(ErrInvalidRollappID, "rollapp-id '%s' cannot exceed 48 chars", chainID)
	}

	matches := ethermintChainID.FindStringSubmatch(chainID)
	if matches == nil || len(matches) != 4 || matches[1] == "" {
		return nil, errorsmod.Wrapf(ErrInvalidRollappID, "wrong formatted EVM Chain ID")
	}

	// verify that the chain-id entered is a base 10 integer
	chainIDInt, ok := new(big.Int).SetString(matches[2], 10)
	if !ok {
		return nil, errorsmod.Wrapf(ErrInvalidRollappID, "epoch %s must be base-10 integer format", matches[2])
	}

	return chainIDInt, nil
}
