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

type ChainID struct {
	chainID  string
	eip155ID *big.Int
}

func NewChainID(id string) (ChainID, error) {
	chainID := strings.TrimSpace(id)

	if chainID == "" {
		return ChainID{}, errorsmod.Wrapf(ErrInvalidRollappID, "empty")
	}

	if len(chainID) > 48 {
		return ChainID{}, errorsmod.Wrapf(ErrInvalidRollappID, "exceeds 48 chars: %s: len: %d", chainID, len(chainID))
	}

	eip155, err := getEIP155ID(chainID)
	if err != nil {
		return ChainID{}, err
	}
	return ChainID{
		chainID:  chainID,
		eip155ID: eip155,
	}, nil
}

// getEIP155ID parses a string chain identifier's epoch to an Ethereum-compatible
// chain-id in *big.Int format. The function returns an error if the chain-id has an invalid format
func getEIP155ID(chainID string) (*big.Int, error) {
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

func (c *ChainID) IsEIP155() bool {
	return c.eip155ID != nil
}

func (c *ChainID) GetChainID() string {
	return c.chainID
}

func (c *ChainID) GetEIP155ID() uint64 {
	if c.eip155ID != nil {
		return c.eip155ID.Uint64()
	}
	return 0
}
