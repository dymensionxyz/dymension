package inclusion_test

import (
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	inclusion "github.com/dymensionxyz/dymension/v3/app/dainclusionproofs"
	abcitypes "github.com/tendermint/tendermint/abci/types"
)

func TestNonInclusionProof(t *testing.T) {

	require := require.New(t)

	file, err := os.Open("./non_inclusion_proof.json")
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file.Close()

	// Decode the JSON-encoded data into your struct
	jsonDecoder := json.NewDecoder(file)
	proof := abcitypes.NonInclusionProof{}
	err = jsonDecoder.Decode(&proof)
	require.NoError(err)

	nonInclusionProof := &inclusion.NonInclusionProof{}

	nonInclusionProof.DataRoot = proof.GetDataroot()
	nonInclusionProof.RowProof = proof.GetRproofs()

	err = nonInclusionProof.VerifyNonInclusion(1500, 50, nonInclusionProof.DataRoot)
	require.NoError(err)

}
