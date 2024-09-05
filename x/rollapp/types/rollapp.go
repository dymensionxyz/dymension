package types

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/testutil/sample"
)

func NewRollapp(
	creator,
	rollappId,
	initialSequencer string,
	vmType Rollapp_VMType,
	metadata *RollappMetadata,
	genInfo GenesisInfo,
	transfersEnabled bool,
) Rollapp {
	return Rollapp{
		RollappId:        rollappId,
		Owner:            creator,
		InitialSequencer: initialSequencer,
		VmType:           vmType,
		Metadata:         metadata,
		GenesisInfo:      genInfo,
		GenesisState: RollappGenesisState{
			TransfersEnabled: transfersEnabled,
		},
	}
}

const (
	maxAppNameLength         = 32
	maxDescriptionLength     = 512
	maxDisplayNameLength     = 32
	maxTaglineLength         = 64
	maxURLLength             = 256
	maxGenesisChecksumLength = 64
	maxDenomBaseLength       = 128
	maxDenomDisplayLength    = 128
)

type AllowedDecimals uint32

const (
	Decimals18 AllowedDecimals = 18
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

	if err = r.GenesisInfo.Validate(); err != nil {
		return err
	}

	if r.VmType == 0 {
		return ErrInvalidVMType
	}

	if r.Metadata != nil {
		if err = r.Metadata.Validate(); err != nil {
			return errors.Join(ErrInvalidMetadata, err)
		}
	}

	// if rollapp is started, genesis info must be sealed
	if r.Launched && !r.GenesisInfo.Sealed {
		return fmt.Errorf("genesis info needs to be sealed if rollapp is started")
	}

	return nil
}

func (r Rollapp) AllImmutableFieldsAreSet() bool {
	return r.InitialSequencer != ""
}

func (r Rollapp) GenesisInfoFieldsAreSet() bool {
	return r.GenesisInfo.GenesisChecksum != "" &&
		r.GenesisInfo.NativeDenom.Validate() == nil &&
		r.GenesisInfo.Bech32Prefix != "" &&
		!r.GenesisInfo.InitialSupply.IsNil()
}

func (r GenesisInfo) Validate() error {
	if r.Bech32Prefix != "" {
		if err := validateBech32Prefix(r.Bech32Prefix); err != nil {
			return errors.Join(ErrInvalidBech32Prefix, err)
		}
	}

	if len(r.GenesisChecksum) > maxGenesisChecksumLength {
		return ErrInvalidGenesisChecksum
	}

	if err := r.NativeDenom.Validate(); err != nil {
		return errors.Join(ErrInvalidNativeDenom, err)
	}

	if r.InitialSupply.IsNil() {
		return ErrInvalidInitialSupply
	}

	return nil
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
		return err
	}

	bAddr, err := sdk.GetFromBech32(bechAddr, prefix)
	if err != nil {
		return err
	}

	if err = sdk.VerifyAddressFormat(bAddr); err != nil {
		return err
	}
	return nil
}

func (dm DenomMetadata) Validate() error {
	if l := len(dm.Base); l == 0 || l > maxDenomBaseLength {
		return fmt.Errorf("base denom")
	}

	if l := len(dm.Display); l == 0 || l > maxDenomDisplayLength {
		return fmt.Errorf("display denom")
	}

	// validate exponent
	if AllowedDecimals(dm.Exponent) != Decimals18 {
		return fmt.Errorf("invalid exponent")
	}

	return nil
}

func (md *RollappMetadata) Validate() error {
	if err := validateURL(md.Website); err != nil {
		return errorsmod.Wrap(err, "website URL")
	}

	if err := validateURL(md.X); err != nil {
		return errorsmod.Wrap(err, "X URL")
	}

	if err := validateURL(md.GenesisUrl); err != nil {
		return errorsmod.Wrap(err, "genesis URL")
	}

	if err := validateURL(md.Telegram); err != nil {
		return errorsmod.Wrap(err, "telegram URL")
	}

	if len(md.Description) > maxDescriptionLength {
		return fmt.Errorf("description too long")
	}

	if len(md.DisplayName) > maxDisplayNameLength {
		return fmt.Errorf("display name too long")
	}

	if len(md.Tagline) > maxTaglineLength {
		return fmt.Errorf("tagline too long")
	}

	if err := validateURL(md.LogoUrl); err != nil {
		return errors.Join(ErrInvalidURL, err)
	}

	if err := validateURL(md.ExplorerUrl); err != nil {
		return errors.Join(ErrInvalidURL, err)
	}

	if md.FeeDenom != nil {
		if err := md.FeeDenom.Validate(); err != nil {
			return errors.Join(ErrInvalidFeeDenom, err)
		}
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
