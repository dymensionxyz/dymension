/*
NOTE: Usage of x/params to manage parameters is deprecated in favor of x/gov
controlled execution of MsgUpdateParams messages. These types remains solely
for migration purposes and will be removed in a future release.
*/
package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/dymensionxyz/sdk-utils/utils/uparam"
)

var _ paramtypes.ParamSet = &Params{}

var (
	KeyMinBond                    = []byte("MinBond")
	KeyKickThreshold              = []byte("KickThreshold")
	KeyNoticePeriod               = []byte("NoticePeriod")
	KeyLivenessSlashMinMultiplier = []byte("LivenessSlashMultiplier")
	KeyLivenessSlashMinAbsolute   = []byte("LivenessSlashMinAbsolute")
)

// Deprecated: ParamKeyTable for module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// ParamSetPairs implements the ParamSet interface and returns all the key/value pairs
// pairs of module's parameters.
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMinBond, &p.MinBond, validateMinBond),
		paramtypes.NewParamSetPair(KeyKickThreshold, &p.KickThreshold, uparam.ValidateCoin),
		paramtypes.NewParamSetPair(KeyNoticePeriod, &p.NoticePeriod, validateTime),
		paramtypes.NewParamSetPair(KeyLivenessSlashMinMultiplier, &p.LivenessSlashMinMultiplier, validateLivenessSlashMultiplier),
		paramtypes.NewParamSetPair(KeyLivenessSlashMinAbsolute, &p.LivenessSlashMinAbsolute, uparam.ValidateCoin),
	}
}
