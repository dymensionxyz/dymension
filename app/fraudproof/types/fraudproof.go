package types

import (
	"bytes"
	"errors"
	fmt "fmt"

	ics23 "github.com/confio/ics23/go"
	db "github.com/tendermint/tm-db"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/iavl"
	abci "github.com/tendermint/tendermint/abci/types"
	tmcrypto "github.com/tendermint/tendermint/proto/tendermint/crypto"
)

var ErrMoreThanOneBlockTypeUsed = errors.New("fraudProof has more than one type of fradulent state transitions marked nil")

// Represents a single-round fraudProof
type FraudProof struct {
	// The block height to load state of
	BlockHeight int64

	// TODO: Add Proof that appHash is inside merklized ISRs in block header at block height

	PreStateAppHash      []byte
	ExpectedValidAppHash []byte
	// A map from module name to state witness
	StateWitness map[string]StateWitness

	// Fraudulent state transition has to be one of these
	// Only one have of these three can be non-nil
	fraudulentBeginBlock *abci.RequestBeginBlock
	fraudulentDeliverTx  *abci.RequestDeliverTx
	fraudulentEndBlock   *abci.RequestEndBlock

	// TODO: Add Proof that fraudulent state transition is inside merkelizied transactions in block header
}

// State witness with a list of all witness data
type StateWitness struct {
	// store level proof
	Proof    tmcrypto.ProofOp
	RootHash []byte
	// List of witness data
	WitnessData []*WitnessData
}

// Witness data represents a trace operation along with inclusion proofs required for said operation
type WitnessData struct {
	Operation iavl.Operation
	Key       []byte
	Value     []byte
	Proofs    []*tmcrypto.ProofOp
}

func convertToProofOps(existenceProofs []*ics23.ExistenceProof) []*tmcrypto.ProofOp {
	if existenceProofs == nil {
		return nil
	}
	proofOps := make([]*tmcrypto.ProofOp, 0)
	for _, existenceProof := range existenceProofs {
		proofOps = append(proofOps, getProofOp(existenceProof))
	}
	return proofOps
}

func getProofOp(exist *ics23.ExistenceProof) *tmcrypto.ProofOp {
	commitmentProof := &ics23.CommitmentProof{
		Proof: &ics23.CommitmentProof_Exist{
			Exist: exist,
		},
	}
	proofOp := storetypes.NewIavlCommitmentOp(exist.Key, commitmentProof).ProofOp()
	return &proofOp
}

func convertToExistenceProofs(proofs []*tmcrypto.ProofOp) ([]*ics23.ExistenceProof, error) {
	existenceProofs := make([]*ics23.ExistenceProof, 0)
	for _, proof := range proofs {
		_, existenceProof, err := GetExistenceProof(*proof)
		if err != nil {
			return nil, err
		}
		existenceProofs = append(existenceProofs, existenceProof)
	}
	return existenceProofs, nil
}

func GetExistenceProof(proofOp tmcrypto.ProofOp) (storetypes.CommitmentOp, *ics23.ExistenceProof, error) {
	op, err := storetypes.CommitmentOpDecoder(proofOp)
	if err != nil {
		return storetypes.CommitmentOp{}, nil, err
	}
	commitmentOp := op.(storetypes.CommitmentOp)

	commitmentProof := commitmentOp.GetProof()
	return commitmentOp, commitmentProof.GetExist(), nil
}

func (fraudProof *FraudProof) GetModules() []string {
	keys := make([]string, 0, len(fraudProof.StateWitness))
	for k := range fraudProof.StateWitness {
		keys = append(keys, k)
	}
	return keys
}

// Returns a map from storeKey to IAVL Deep Subtrees which have witness data and
// initial root hash initialized from fraud proof
func (fraudProof *FraudProof) GetDeepIAVLTrees() (map[string]*iavl.DeepSubTree, error) {
	storeKeyToIAVLTree := make(map[string]*iavl.DeepSubTree)
	for storeKey, stateWitness := range fraudProof.StateWitness {
		dst := iavl.NewDeepSubTree(db.NewMemDB(), 100, false, fraudProof.BlockHeight)
		iavlWitnessData := make([]iavl.WitnessData, 0)
		for _, witnessData := range stateWitness.WitnessData {
			existenceProofs, err := convertToExistenceProofs(witnessData.Proofs)
			if err != nil {
				return nil, err
			}
			iavlWitnessData = append(
				iavlWitnessData,
				iavl.WitnessData{
					Operation: witnessData.Operation,
					Key:       witnessData.Key,
					Value:     witnessData.Value,
					Proofs:    existenceProofs,
				},
			)
			dst.SetWitnessData(iavlWitnessData)
		}
		dst.SetInitialRootHash(stateWitness.RootHash)
		storeKeyToIAVLTree[storeKey] = dst
	}
	return storeKeyToIAVLTree, nil
}

// Returns true only if only one of the three pointers is nil
func (fraudProof *FraudProof) CheckFraudulentStateTransition() bool {
	if fraudProof.fraudulentBeginBlock != nil {
		return fraudProof.fraudulentDeliverTx == nil && fraudProof.fraudulentEndBlock == nil
	}
	if fraudProof.fraudulentDeliverTx != nil {
		return fraudProof.fraudulentEndBlock == nil
	}
	return fraudProof.fraudulentEndBlock != nil
}

// Performs fraud proof verification on a store and substore level
func (fraudProof *FraudProof) VerifyFraudProof() (bool, error) {
	if !fraudProof.CheckFraudulentStateTransition() {
		return false, ErrMoreThanOneBlockTypeUsed
	}
	for storeKey, stateWitness := range fraudProof.StateWitness {
		// Fraudproof verification on a store level
		proofOp := stateWitness.Proof
		proof, err := storetypes.CommitmentOpDecoder(proofOp)
		if err != nil {
			return false, err
		}
		if !bytes.Equal(proof.GetKey(), []byte(storeKey)) {
			return false, fmt.Errorf("got storeKey: %s, expected: %s", string(proof.GetKey()), storeKey)
		}
		appHash, err := proof.Run([][]byte{stateWitness.RootHash})
		if err != nil {
			return false, err
		}
		if !bytes.Equal(appHash[0], fraudProof.PreStateAppHash) {
			return false, fmt.Errorf("got appHash: %s, expected: %s", string(fraudProof.PreStateAppHash), string(fraudProof.PreStateAppHash))
		}

		// Fraudproof verification on a substore level
		// Note: We can only verify the first witness in this witnessData
		// with current root hash. Other proofs are verified in the IAVL tree.
		if len(stateWitness.WitnessData) > 0 {
			witness := stateWitness.WitnessData[0]
			for _, proofOp := range witness.Proofs {
				op, existenceProof, err := GetExistenceProof(*proofOp)
				if err != nil {
					return false, err
				}
				verified := ics23.VerifyMembership(op.Spec, stateWitness.RootHash, op.Proof, op.Key, existenceProof.Value)
				if !verified {
					return false, fmt.Errorf("existence proof verification failed, expected rootHash: %s, key: %s, value: %s for storeKey: %s", string(stateWitness.RootHash), string(op.Key), string(existenceProof.Value), storeKey)
				}
			}
		}
	}
	return true, nil
}
