package types

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/types"
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
	name     string
	eip155ID *big.Int
	revision uint64
}

func NewChainID(id string) (ChainID, error) {
	chainID := strings.TrimSpace(id)

	if chainID == "" {
		return ChainID{}, errorsmod.Wrapf(ErrInvalidRollappID, "empty")
	}

	if len(chainID) > types.MaxChainIDLen {
		return ChainID{}, errorsmod.Wrapf(ErrInvalidRollappID, "exceeds %d chars: %s: len: %d", types.MaxChainIDLen, chainID, len(chainID))
	}

	matches := ethermintChainID.FindStringSubmatch(chainID)

	if matches == nil || len(matches) != 4 || matches[1] == "" {
		return ChainID{}, ErrInvalidRollappID
	}
	// verify that the chain-id entered is a base 10 integer
	chainIDInt, ok := new(big.Int).SetString(matches[2], 10)
	if !ok {
		return ChainID{}, errorsmod.Wrapf(ErrInvalidRollappID, "EIP155 part %s must be base-10 integer format", matches[2])
	}

	revision, err := strconv.ParseUint(matches[3], 0, 64)
	if err != nil {
		return ChainID{}, errorsmod.Wrapf(ErrInvalidRollappID, "parse revision number: error: %v", err)
	}

	return ChainID{
		chainID:  chainID,
		eip155ID: chainIDInt,
		revision: revision,
		name:     matches[1],
	}, nil
}

func MustNewChainID(id string) ChainID {
	chainID, err := NewChainID(id)
	if err != nil {
		panic(err)
	}
	return chainID
}

func (c *ChainID) GetChainID() string {
	return c.chainID
}

func (c *ChainID) GetName() string {
	return c.name
}

func (c *ChainID) GetEIP155ID() uint64 {
	if c.eip155ID != nil {
		return c.eip155ID.Uint64()
	}
	return 0
}

func (c *ChainID) GetRevisionNumber() uint64 {
	return c.revision
}
