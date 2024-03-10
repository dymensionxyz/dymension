package types

// DONTCOVER

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// x/rollapp module sentinel errors
var (
	ErrRollappExists                       = sdkerrors.Register(ModuleName, 1000, "rollapp already exist for this rollapp-id; must use new rollapp-id")
	ErrInvalidMaxSequencers                = sdkerrors.Register(ModuleName, 1001, "invalid max sequencers")
	ErrInvalidPermissionedAddress          = sdkerrors.Register(ModuleName, 1003, "invalid permissioned address")
	ErrPermissionedAddressesDuplicate      = sdkerrors.Register(ModuleName, 1004, "permissioned-address has duplicates")
	ErrInvalidNumBlocks                    = sdkerrors.Register(ModuleName, 1005, "invalid number of blocks")
	ErrInvalidBlockSequence                = sdkerrors.Register(ModuleName, 1006, "invalid block sequence")
	ErrUnknownRollappID                    = sdkerrors.Register(ModuleName, 1007, "rollapp does not exist")
	ErrVersionMismatch                     = sdkerrors.Register(ModuleName, 1008, "rollapp version mismatch")
	ErrWrongBlockHeight                    = sdkerrors.Register(ModuleName, 1009, "start-height does not match rollapps state")
	ErrMultiUpdateStateInBlock             = sdkerrors.Register(ModuleName, 1010, "only one state update can take place per block")
	ErrInvalidStateRoot                    = sdkerrors.Register(ModuleName, 1011, "invalid blocks state root")
	ErrInvalidIntermediateStatesRoot       = sdkerrors.Register(ModuleName, 1012, "invalid blocks intermediate states root")
	ErrUnauthorizedRollappCreator          = sdkerrors.Register(ModuleName, 1013, "rollapp creator not register in whitelist")
	ErrInvalidClientType                   = sdkerrors.Register(ModuleName, 1014, "client type of the rollapp isn't dymint")
	ErrHeightStateNotFinalized             = sdkerrors.Register(ModuleName, 1015, "rollapp block on this height was not finalized yet")
	ErrInvalidAppHash                      = sdkerrors.Register(ModuleName, 1016, "the app hash is different from the finalized state root")
	ErrStateNotExists                      = sdkerrors.Register(ModuleName, 1017, "state of this height doesn't exist")
	ErrInvalidHeight                       = sdkerrors.Register(ModuleName, 1018, "invalid rollapp height")
	ErrRollappCreatorExceedMaximumRollapps = sdkerrors.Register(ModuleName, 1019, "rollapp creator exceed maximum allowed rollapps as register in whitelist")
	ErrInvalidRollappID                    = sdkerrors.Register(ModuleName, 1020, "invalid rollapp-id")
	ErrEIP155Exists                        = sdkerrors.Register(ModuleName, 1021, "EIP155 already exist; must use unique EIP155 identifier")
	ErrRollappsDisabled                    = sdkerrors.Register(ModuleName, 1022, "rollapps are disabled")
	ErrInvalidTokenMetadata                = sdkerrors.Register(ModuleName, 1023, "invalid token metadata")
	ErrNoFinalizedStateYetForRollapp       = sdkerrors.Register(ModuleName, 1024, "no finalized state yet for rollapp")
	ErrInvalidClientState                  = sdkerrors.Register(ModuleName, 1025, "invalid client state")
	ErrInvalidSequencer                    = sdkerrors.Register(ModuleName, 1026, "invalid sequencer")
	ErrInvalidGenesisChannelId             = sdkerrors.Register(ModuleName, 1027, "invalid genesis channel id")
	ErrGenesisEventNotDefined              = sdkerrors.Register(ModuleName, 1028, "genesis event not defined")
	ErrGenesisEventAlreadyTriggered        = sdkerrors.Register(ModuleName, 1029, "genesis event already triggered")
)
