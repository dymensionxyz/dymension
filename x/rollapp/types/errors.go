package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/rollapp module sentinel errors
var (
	ErrRollappExists                       = errorsmod.Register(ModuleName, 1000, "rollapp already exist for this rollapp-id; must use new rollapp-id")
	ErrInvalidMaxSequencers                = errorsmod.Register(ModuleName, 1001, "invalid max sequencers")
	ErrInvalidCreatorAddress               = errorsmod.Register(ModuleName, 1002, "invalid creator address")
	ErrInvalidPermissionedAddress          = errorsmod.Register(ModuleName, 1003, "invalid permissioned address")
	ErrPermissionedAddressesDuplicate      = errorsmod.Register(ModuleName, 1004, "permissioned-address has duplicates")
	ErrInvalidNumBlocks                    = errorsmod.Register(ModuleName, 1005, "invalid number of blocks")
	ErrInvalidBlockSequence                = errorsmod.Register(ModuleName, 1006, "invalid block sequence")
	ErrUnknownRollappID                    = errorsmod.Register(ModuleName, 1007, "rollapp does not exist")
	ErrVersionMismatch                     = errorsmod.Register(ModuleName, 1008, "rollapp version mismatch")
	ErrWrongBlockHeight                    = errorsmod.Register(ModuleName, 1009, "start-height does not match rollapps state")
	ErrMultiUpdateStateInBlock             = errorsmod.Register(ModuleName, 1010, "only one state update can take place per block")
	ErrInvalidStateRoot                    = errorsmod.Register(ModuleName, 1011, "invalid blocks state root")
	ErrInvalidIntermediateStatesRoot       = errorsmod.Register(ModuleName, 1012, "invalid blocks intermediate states root")
	ErrUnauthorizedRollappCreator          = errorsmod.Register(ModuleName, 1013, "rollapp creator not register in whitelist")
	ErrInvalidClientType                   = errorsmod.Register(ModuleName, 1014, "client type of the rollapp isn't dymint")
	ErrHeightStateNotFinalized             = errorsmod.Register(ModuleName, 1015, "rollapp block on this height was not finalized yet")
	ErrInvalidAppHash                      = errorsmod.Register(ModuleName, 1016, "the app hash is different from the finalized state root")
	ErrStateNotExists                      = errorsmod.Register(ModuleName, 1017, "state of this height doesn't exist")
	ErrInvalidHeight                       = errorsmod.Register(ModuleName, 1018, "invalid rollapp height")
	ErrRollappCreatorExceedMaximumRollapps = errorsmod.Register(ModuleName, 1019, "rollapp creator exceed maximum allowed rollapps as register in whitelist")
	ErrInvalidRollappID                    = errorsmod.Register(ModuleName, 1020, "invalid rollapp-id")
	ErrEIP155Exists                        = errorsmod.Register(ModuleName, 1021, "EIP155 already exist; must use unique EIP155 identifier")
	ErrRollappsDisabled                    = errorsmod.Register(ModuleName, 1022, "rollapps are disabled")
	ErrInvalidTokenMetadata                = errorsmod.Register(ModuleName, 1023, "invalid token metadata")
	ErrNoFinalizedStateYetForRollapp       = errorsmod.Register(ModuleName, 1024, "no finalized state yet for rollapp")
	ErrInvalidClientState                  = errorsmod.Register(ModuleName, 1025, "invalid client state")
	ErrInvalidSequencer                    = errorsmod.Register(ModuleName, 1026, "invalid sequencer")
	ErrInvalidGenesisChannelId             = errorsmod.Register(ModuleName, 1027, "invalid genesis channel id")
	ErrGenesisEventNotDefined              = errorsmod.Register(ModuleName, 1028, "genesis event not defined")
	ErrGenesisEventAlreadyTriggered        = errorsmod.Register(ModuleName, 1029, "genesis event already triggered")
	ErrInvalidGenesisAccount               = errorsmod.Register(ModuleName, 1030, "invalid genesis account")
	/* ------------------------------ fraud related ----------------------------- */
	ErrDisputeAlreadyFinalized = errorsmod.Register(ModuleName, 2000, "disputed height already finalized")
	ErrDisputeAlreadyReverted  = errorsmod.Register(ModuleName, 2001, "disputed height already reverted")
	ErrWrongClientId           = errorsmod.Register(ModuleName, 2002, "client id does not match the rollapp")
	ErrWrongProposerAddr       = errorsmod.Register(ModuleName, 2003, "wrong proposer address")
)
