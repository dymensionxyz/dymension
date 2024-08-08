package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// x/rollapp module sentinel errors
var (
	ErrRollappExists                   = gerrc.ErrAlreadyExists.Wrap("rollapp")
	ErrInvalidInitialSequencer         = gerrc.ErrInvalidArgument.Wrap("initial sequencer")
	ErrInvalidCreatorAddress           = gerrc.ErrInvalidArgument.Wrap("creator address")
	ErrInvalidBech32Prefix             = gerrc.ErrInvalidArgument.Wrap("Bech32 prefix")
	ErrRollappFrozen                   = gerrc.ErrFailedPrecondition.Wrap("rollapp is frozen")
	ErrInvalidNumBlocks                = gerrc.ErrInvalidArgument.Wrap("number of blocks")
	ErrInvalidBlockSequence            = gerrc.ErrInvalidArgument.Wrap("block sequence")
	ErrUnknownRollappID                = gerrc.ErrNotFound.Wrap("rollapp does not exist")
	ErrWrongBlockHeight                = gerrc.ErrInvalidArgument.Wrap("start-height does not match rollapps state")
	ErrInvalidGenesisChecksum          = gerrc.ErrInvalidArgument.Wrap("genesis checksum")
	ErrInvalidStateRoot                = gerrc.ErrInvalidArgument.Wrap("blocks state root")
	ErrFeePayment                      = errorsmod.Register(ModuleName, 1013, "rollapp creation fee payment error")
	ErrStateNotExists                  = gerrc.ErrNotFound.Wrap("state of this height doesn't exist")
	ErrInvalidHeight                   = gerrc.ErrInvalidArgument.Wrap("rollapp height")
	ErrInvalidRollappID                = gerrc.ErrInvalidArgument.Wrap("rollapp-id")
	ErrNoFinalizedStateYetForRollapp   = gerrc.ErrNotFound.Wrap("no finalized state yet for rollapp")
	ErrInvalidClientState              = gerrc.ErrInvalidArgument.Wrap("client state")
	ErrRollappNotRegistered            = gerrc.ErrFailedPrecondition.Wrap("rollapp not registered")
	ErrUnknownRequest                  = gerrc.ErrFailedPrecondition.Wrap("unknown request")
	ErrNotFound                        = gerrc.ErrNotFound
	ErrLogic                           = gerrc.ErrInternal
	ErrInvalidAddress                  = gerrc.ErrInvalidArgument.Wrap("address")
	ErrInvalidAlias                    = errorsmod.Wrap(gerrc.ErrInvalidArgument, "alias")
	ErrInvalidURL                      = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid url")
	ErrInvalidDescription              = errorsmod.Wrap(gerrc.ErrInvalidArgument, "description")
	ErrInvalidLogoURI                  = errorsmod.Wrap(gerrc.ErrInvalidArgument, "logo uri")
	ErrInvalidTokenLogoURI             = errorsmod.Wrap(gerrc.ErrInvalidArgument, "token logo uri")
	ErrInvalidMetadata                 = errorsmod.Wrap(gerrc.ErrInvalidArgument, "metadata")
	ErrImmutableFieldUpdateAfterSealed = errorsmod.Wrap(gerrc.ErrInvalidArgument, "update immutable field after rollapp sealed")
	ErrSealWithImmutableFieldsNotSet   = errorsmod.Wrap(gerrc.ErrInvalidArgument, "seal with immutable fields not set")
	ErrUnauthorizedSigner              = errorsmod.Wrap(gerrc.ErrPermissionDenied, "unauthorized signer")
	ErrSameOwner                       = errorsmod.Wrap(gerrc.ErrInvalidArgument, "same owner")
	ErrInvalidRequest                  = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid request")
	ErrInvalidVMType                   = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid vm type")

	/* ------------------------------ fraud related ----------------------------- */

	ErrDisputeAlreadyFinalized = gerrc.ErrFailedPrecondition.Wrap("disputed height already finalized")
	ErrDisputeAlreadyReverted  = gerrc.ErrFailedPrecondition.Wrap("disputed height already reverted")
	ErrWrongClientId           = gerrc.ErrInvalidArgument.Wrap("client id does not match the rollapp")
	ErrWrongProposerAddr       = gerrc.ErrInvalidArgument.Wrap("wrong proposer address")
)
