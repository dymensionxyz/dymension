package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// x/sequencer module sentinel errors
var (
	ErrSequencerExists          = errorsmod.Register(ModuleName, 1000, "sequencer already exist for this address; must use new sequencer address")
	ErrUnknownRollappID         = errorsmod.Register(ModuleName, 1002, "rollapp does not exist")
	ErrUnknownSequencer         = errorsmod.Register(ModuleName, 1005, "sequencer was not registered")
	ErrSequencerRollappMismatch = errorsmod.Register(ModuleName, 1006, "sequencer was not registered for this rollapp")
	ErrNotActiveSequencer       = errorsmod.Register(ModuleName, 1007, "sequencer is not active")
	ErrInvalidSequencerStatus   = errorsmod.Register(ModuleName, 1008, "invalid sequencer status")
	ErrInvalidCoinDenom         = errorsmod.Register(ModuleName, 1010, "invalid coin denomination")
	ErrInsufficientBond         = errorsmod.Register(ModuleName, 1011, "insufficient bond")
	ErrRollappFrozen            = errorsmod.Register(ModuleName, 1012, "rollapp is frozen")
	ErrInvalidAddress           = errorsmod.Register(ModuleName, 1013, "invalid address")
	ErrInvalidPubKey            = errorsmod.Register(ModuleName, 1014, "invalid pubkey")
	ErrInvalidCoins             = errorsmod.Register(ModuleName, 1015, "invalid coins")
	ErrInvalidType              = errorsmod.Register(ModuleName, 1016, "invalid type")
	ErrUnknownRequest           = errorsmod.Register(ModuleName, 1017, "unknown request")
	ErrInvalidRequest           = errorsmod.Register(ModuleName, 1018, "invalid request")
	ErrSequencerJailed          = errorsmod.Register(ModuleName, 1019, "sequencer is jailed")
	ErrRotationInProgress       = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "sequencer rotation in progress")
	ErrBeforePreLaunchTime      = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "before pre-launch time")
	ErrNoProposer               = errorsmod.Wrap(gerrc.ErrNotFound, "proposer")
	ErrNotInitialSequencer      = errorsmod.Wrap(gerrc.ErrFailedPrecondition, "not the initial sequencer")
	ErrInvalidURL               = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid url")
	ErrInvalidMetadata          = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid metadata")
	ErrInvalidVMTypeUpdate      = errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid vm type update")
	ErrUnknownBondReduction     = errorsmod.Wrap(gerrc.ErrNotFound, "unknown bond reduction")
)
