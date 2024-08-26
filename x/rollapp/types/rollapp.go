package types

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
)

func NewRollapp(
	creator,
	rollappId,
	initSequencer,
	bech32Prefix,
	genesisChecksum string,
	vmType Rollapp_VMType,
	metadata *RollappMetadata,
	transfersEnabled bool,
) Rollapp {
	return Rollapp{
		RollappId:        rollappId,
		Owner:            creator,
		InitialSequencer: initSequencer,
		GenesisChecksum:  genesisChecksum,
		Bech32Prefix:     bech32Prefix,
		VmType:           vmType,
		Metadata:         metadata,
		GenesisState: RollappGenesisState{
			TransfersEnabled: transfersEnabled,
		},
	}
}

const (
	maxDescriptionLength     = 512
	maxDisplayNameLength     = 32
	maxTaglineLength         = 64
	maxURLLength             = 256
	maxGenesisChecksumLength = 64
)

func (r Rollapp) LastStateUpdateHeightIsSet() bool {
	return r.LastStateUpdateHeight != 0
}

func (r Rollapp) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(r.Owner)
	if err != nil {
		return errorsmod.Wrap(ErrInvalidCreatorAddress, err.Error())
	}

	// validate rollappId
	_, err = NewChainID(r.RollappId)
	if err != nil {
		return err
	}

	if err = validateInitialSequencer(r.InitialSequencer); err != nil {
		return errorsmod.Wrap(ErrInvalidInitialSequencer, err.Error())
	}

	if r.Bech32Prefix != "" {
		if err = validateBech32Prefix(r.Bech32Prefix); err != nil {
			return gerrc.ErrInvalidArgument.Wrap("bech32")
		}
	}

	if len(r.GenesisChecksum) > maxGenesisChecksumLength {
		return errorsmod.Wrap(ErrInvalidGenesisChecksum, "GenesisChecksum")
	}

	if r.VmType == 0 {
		return ErrInvalidVMType
	}

	if err = validateMetadata(r.Metadata); err != nil {
		return errorsmod.Wrap(ErrInvalidMetadata, err.Error())
	}

	return nil
}

func (r Rollapp) AllImmutableFieldsAreSet() bool {
	return r.GenesisChecksum != "" && r.InitialSequencer != "" && r.Bech32Prefix != ""
}

func validateInitialSequencer(initialSequencer string) error {
	if initialSequencer == "" || initialSequencer == "*" {
		return nil
	}

	seen := make(map[string]struct{})
	addrs := strings.Split(initialSequencer, ",")

	for _, addr := range addrs {
		if _, ok := seen[addr]; ok {
			return ErrInvalidInitialSequencer
		}
		seen[addr] = struct{}{}
		_, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			return err
		}
	}

	return nil
}

func validateBech32Prefix(prefix string) error {
	bechAddr, err := sdk.Bech32ifyAddressBytes(prefix, sample.Acc())
	if err != nil {
		return errorsmod.Wrap(err, "bech32ify addr bytes")
	}

	bAddr, err := sdk.GetFromBech32(bechAddr, prefix)
	if err != nil {
		return errorsmod.Wrap(err, "get from bech 32")
	}

	if err = sdk.VerifyAddressFormat(bAddr); err != nil {
		return errorsmod.Wrap(err, "verify addr format")
	}
	return nil
}

func validateMetadata(metadata *RollappMetadata) error {
	if metadata == nil {
		return nil
	}

	if err := validateURL(metadata.Website); err != nil {
		return errorsmod.Wrap(ErrInvalidURL, err.Error())
	}

	if err := validateURL(metadata.X); err != nil {
		return errorsmod.Wrap(ErrInvalidURL, err.Error())
	}

	if err := validateURL(metadata.GenesisUrl); err != nil {
		return errorsmod.Wrap(errors.Join(ErrInvalidURL, err), "genesis url")
	}

	if err := validateURL(metadata.Telegram); err != nil {
		return errorsmod.Wrap(ErrInvalidURL, err.Error())
	}

	if len(metadata.Description) > maxDescriptionLength {
		return ErrInvalidDescription
	}

	if len(metadata.DisplayName) > maxDisplayNameLength {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "display name too long")
	}

	if len(metadata.Tagline) > maxTaglineLength {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "tagline too long")
	}

	if err := validateURL(metadata.LogoUrl); err != nil {
		return errorsmod.Wrap(ErrInvalidURL, err.Error())
	}

	if err := validateURL(metadata.ExplorerUrl); err != nil {
		return errorsmod.Wrap(ErrInvalidURL, err.Error())
	}

	return nil
}

func validateURL(urlStr string) error {
	if urlStr == "" {
		return nil
	}

	if len(urlStr) > maxURLLength {
		return fmt.Errorf("URL exceeds maximum length")
	}

	if _, err := url.Parse(urlStr); err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	return nil
}
