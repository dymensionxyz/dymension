package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// x/rollapp module sentinel errors
var (
	ErrRollappIDExists                 = errorsmod.Register(ModuleName, 1000, "rollapp already exists for this rollapp-id; must use new rollapp-id")
	ErrInvalidInitialSequencer         = errorsmod.Register(ModuleName, 1001, "empty initial sequencer")
	ErrInvalidCreatorAddress           = errorsmod.Register(ModuleName, 1002, "invalid creator address")
	ErrInvalidBech32Prefix             = errorsmod.Register(ModuleName, 1003, "invalid Bech32 prefix")
	ErrRollappFrozen                   = errorsmod.Register(ModuleName, 1004, "rollapp is frozen")
	ErrInvalidNumBlocks                = errorsmod.Register(ModuleName, 1005, "invalid number of blocks")
	ErrInvalidBlockSequence            = errorsmod.Register(ModuleName, 1006, "invalid block sequence")
	ErrUnknownRollappID                = errorsmod.Register(ModuleName, 1007, "rollapp does not exist")
	ErrWrongBlockHeight                = errorsmod.Register(ModuleName, 1009, "start-height does not match rollapps state")
	ErrInvalidGenesisChecksum          = errorsmod.Register(ModuleName, 1010, "invalid genesis checksum")
	ErrInvalidStateRoot                = errorsmod.Register(ModuleName, 1011, "invalid blocks state root")
	ErrFeePayment                      = errorsmod.Register(ModuleName, 1013, "rollapp creation fee payment error")
	ErrStateNotExists                  = errorsmod.Register(ModuleName, 1017, "state of this height doesn't exist")
	ErrInvalidHeight                   = errorsmod.Register(ModuleName, 1018, "invalid rollapp height")
	ErrInvalidRollappID                = errorsmod.Register(ModuleName, 1020, "invalid rollapp-id")
	ErrNoFinalizedStateYetForRollapp   = errorsmod.Register(ModuleName, 1024, "no finalized state yet for rollapp")
	ErrInvalidClientState              = errorsmod.Register(ModuleName, 1025, "invalid client state")
	ErrRollappNotRegistered            = errorsmod.Register(ModuleName, 1035, "rollapp not registered")
	ErrUnknownRequest                  = errorsmod.Register(ModuleName, 1036, "unknown request")
	ErrNotFound                        = errorsmod.Register(ModuleName, 1037, "not found")
	ErrLogic                           = errorsmod.Register(ModuleName, 1038, "internal logic error")
	ErrInvalidAddress                  = errorsmod.Register(ModuleName, 1040, "invalid address")
	ErrInvalidDAClientType            = errorsmod.Register(ModuleName, 1041, "invalid data availability client")
	ErrDAClientValidationFailed       = errorsmod.Register(ModuleName, 1042, "data availability client validation failed")
	ErrInvalidAlias                    = errorsmod.Wrap(gerrc.ErrInvalidArgument, "alias")
	ErrInvalidURL                      = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid url")
	ErrInvalidDescription              = errorsmod.Wrap(gerrc.ErrInvalidArgument, "description")
	ErrInvalidLogoURI                  = errorsmod.Wrap(gerrc.ErrInvalidArgument, "logo uri")
	ErrInvalidTokenLogoURI             = errorsmod.Wrap(gerrc.ErrInvalidArgument, "token logo uri")
	ErrInvalidMetadata                 = errorsmod.Wrap(gerrc.ErrInvalidArgument, "metadata")
	ErrImmutableFieldUpdateAfterSealed = errorsmod.Wrap(gerrc.ErrInvalidArgument, "update immutable field after rollapp sealed")
	ErrSealWithImmutableFieldsNotSet   = errorsmod.Wrap(gerrc.ErrInvalidArgument, "seal with immutable fields not set")
	ErrInvalidHandle                   = errorsmod.Wrap(gerrc.ErrInvalidArgument, "handle")
	ErrInvalidRequest                  = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid request")
	ErrRollappAliasExists              = errorsmod.Wrap(gerrc.ErrInvalidArgument, "rollapp already exists for this alias; must use new alias")

	/* ------------------------------ fraud related ----------------------------- */
	ErrDisputeAlreadyFinalized = errorsmod.Register(ModuleName, 2000, "disputed height already finalized")
	ErrDisputeAlreadyReverted  = errorsmod.Register(ModuleName, 2001, "disputed height already reverted")
	ErrWrongClientId           = errorsmod.Register(ModuleName, 2002, "client id does not match the rollapp")
	ErrWrongProposerAddr       = errorsmod.Register(ModuleName, 2003, "wrong proposer address")
)
