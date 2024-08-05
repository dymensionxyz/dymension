package types

type SimRollapp struct {
	// rollappId is the unique identifier of the rollapp chain.
	// The rollappId follows the same standard as cosmos chain_id.
	RollappId string
	// Sequencers is a list of indexes of sequencers in
	// GlobalSequencerAddressesList by joining order
	Sequencers []int
	// LastHeight is the last updated rollapp height
	LastHeight uint64
	// LastCreationHeight is the last block height that an update was created in
	LastCreationHeight uint64
}
