package types

type SimRollapp struct {
	// rollappId is the unique identifier of the rollapp chain.
	// The rollappId follows the same standard as cosmos chain_id.
	RollappId string
	// maxSequencers is the maximum number of sequencers.
	MaxSequencers uint64
	// permissionedAddresses is a bech32-encoded address list of the
	// sequencers that are allowed to serve this rollappId.
	// In the case of an empty list, the rollapp is considered permissionless.
	PermissionedAddresses []string
	// Sequencers is a list of indexes of sequencers in
	// GlobalSequencerAddressesList by joining order
	Sequencers []int
	// LastHeight is the last updated rollapp height
	LastHeight uint64
	// LastCreationHeight is the last block height that an update was created in
	LastCreationHeight uint64
}
