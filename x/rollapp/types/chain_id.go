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

// GetValidEIP155ChainId parses a string chain identifier's epoch to an Ethereum-compatible
// chain-id in *big.Int format. The function returns an error if the chain-id has an invalid format
func GetValidEIP155ChainId(chainID string) (string, *big.Int, error) {
	chainID = strings.TrimSpace(chainID)

	if chainID == "" {
		return chainID, nil, errorsmod.Wrapf(ErrInvalidRollappID, "chain-id cannot be empty")
	}

	if len(chainID) > 48 {
		return chainID, nil, errorsmod.Wrapf(ErrInvalidRollappID, "rollapp-id '%s' cannot exceed 48 chars", chainID)
	}

	matches := ethermintChainID.FindStringSubmatch(chainID)
	if matches == nil || len(matches) != 4 || matches[1] == "" {
		return chainID, nil, nil
	}

	// verify that the chain-id entered is a base 10 integer
	chainIDInt, ok := new(big.Int).SetString(matches[2], 10)
	if !ok {
		return chainID, nil, errorsmod.Wrapf(ErrInvalidRollappID, "epoch %s must be base-10 integer format", matches[2])
	}

	return chainID, chainIDInt, nil
}
