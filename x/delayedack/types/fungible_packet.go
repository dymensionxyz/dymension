package types

import (
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type TransferDataWithFinalization struct {
	rollapptypes.TransferData
	// Proof height is only be populated if and only if the rollappID is not empty
	ProofHeight uint64
	// Finalized is only be populated if and only if the rollappID is not empty
	Finalized bool
}
