package types

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/testutil/sample"
)

const (
	maxAppNameLength         = 32
	maxDescriptionLength     = 512
	maxDisplayNameLength     = 32
	maxTaglineLength         = 64
	maxURLLength             = 256
	maxGenesisChecksumLength = 64
	maxTags                  = 3
)

// RollappTags are hardcoded for now TODO: manage them through the store
var RollappTags = []string{
	"Meme",
	"AI",
	"DeFI",
	"NFT",
	"Gaming",
	"Betting",
	"Community",
	"Social",
	"DePIN",
	"Launchpad",
}

type AllowedDecimals uint32

const (
	Decimals18 AllowedDecimals = 18
)

func NewRollapp(creator, rollappId, initialSequencer string, minSequencerBond sdk.Coin, vmType Rollapp_VMType, metadata *RollappMetadata, genInfo GenesisInfo) Rollapp {
	return Rollapp{
		RollappId:        rollappId,
		Owner:            creator,
		InitialSequencer: initialSequencer,
		MinSequencerBond: sdk.Coins{minSequencerBond},
		VmType:           vmType,
		Metadata:         metadata,
		GenesisInfo:      genInfo,
		GenesisState:     RollappGenesisState{},
		Revisions: []Revision{{
			Number:      0,
			StartHeight: 0,
		}},
	}
}

func (r Rollapp) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(r.Owner)
	if err != nil {
		return errors.Join(ErrInvalidCreatorAddress, err)
	}

	// validate rollappId
	_, err = NewChainID(r.RollappId)
	if err != nil {
		return err
	}

	if err = ValidateBasicMinSeqBondCoins(r.MinSequencerBond); err != nil {
		return errorsmod.Wrap(err, "min sequencer bond")
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

func (r Rollapp) IsTransferEnabled() bool {
	return r.GenesisState.IsTransferEnabled()
}

func (s RollappGenesisState) IsTransferEnabled() bool {
	return s.TransferProofHeight != 0
}

func (r Rollapp) AllImmutableFieldsAreSet() bool {
	return r.InitialSequencer != "" && r.GenesisInfoFieldsAreSet() && ValidateBasicMinSeqBondCoins(r.MinSequencerBond) == nil
}

func (r Rollapp) GenesisInfoFieldsAreSet() bool {
	return r.GenesisInfo.AllSet()
}

func (r Rollapp) LatestRevision() Revision {
	if len(r.Revisions) == 0 {
		// Revision 0 if no revisions exist.
		// Should happen only in tests.
		return Revision{}
	}
	return r.Revisions[len(r.Revisions)-1]
}

// TODO: rollapp type method should be more robust https://github.com/dymensionxyz/dymension/issues/1596
func (r Rollapp) GetRevisionForHeight(h uint64) Revision {
	for i := len(r.Revisions) - 1; i >= 0; i-- {
		if r.Revisions[i].StartHeight <= h {
			return r.Revisions[i]
		}
	}
	return Revision{}
}

func (r Rollapp) IsRevisionStartHeight(revision, height uint64) bool {
	rev := r.GetRevisionForHeight(height)
	return rev.Number == revision && rev.StartHeight == height
}

func (r Rollapp) DidFork() bool {
	return 1 < len(r.Revisions)
}

func (r *Rollapp) BumpRevision(nextRevisionStartHeight uint64) {
	r.Revisions = append(r.Revisions, Revision{
		Number:      r.LatestRevision().Number + 1,
		StartHeight: nextRevisionStartHeight,
	})
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

func (dm DenomMetadata) IsSet() bool {
	return dm != DenomMetadata{}
}

func (dm DenomMetadata) Validate() error {
	if err := sdk.ValidateDenom(dm.Base); err != nil {
		return fmt.Errorf("invalid metadata base denom: %w", err)
	}

	if err := sdk.ValidateDenom(dm.Display); err != nil {
		return fmt.Errorf("invalid metadata display denom: %w", err)
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

	if len(md.Tags) > maxTags {
		return ErrTooManyTags
	}

	seen := make(map[string]struct{})

	for _, tag := range md.Tags {
		if !slices.Contains(RollappTags, tag) {
			return ErrInvalidTag
		}
		if _, ok := seen[tag]; ok {
			return ErrDuplicateTag
		}
		seen[tag] = struct{}{}
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
