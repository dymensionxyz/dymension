package tee_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/common/tee"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

var (
	//go:embed testdata/tee/confidential_space_root.pem
	gcpRootCertificate string

	//go:embed testdata/tee/secure_policy_values.json
	securePolicyValues string
	//go:embed testdata/tee/secure_query.rego
	securePolicyQuery string
	//go:embed testdata/tee/secure_policy.rego
	securePolicyStructure string

	//go:embed testdata/tee/insecure_policy_values.json
	insecurePolicyValues string
	//go:embed testdata/tee/insecure_query.rego
	insecurePolicyQuery string
	//go:embed testdata/tee/insecure_policy.rego
	insecurePolicyStructure string
)

// claims that satisfy the secure policy values fixture.
func goodClaims(nonce string) jwt.MapClaims {
	return jwt.MapClaims{
		"hwmodel":   "GCP_INTEL_TDX",
		"aud":       "dymension",
		"iss":       "https://confidentialcomputing.googleapis.com",
		"secboot":   true,
		"swname":    "CONFIDENTIAL_SPACE",
		"dbgstat":   "disabled-since-boot",
		"eat_nonce": nonce,
		"submods": map[string]any{
			"container": map[string]any{
				"image_digest":   "sha256:bc4c32cb2ca046ba07dcd964b07a320b7d0ca88a5cf8e979da15cae68a2103ee",
				"restart_policy": "Never",
			},
		},
	}
}

// TestEvaluateOPAPolicy exercises the OPA/rego evaluation layer over synthetic
// claims (no GCP signature). The same claims pass the secure policy values and
// fail the insecure ones (different allowed image digest).
func TestEvaluateOPAPolicy(t *testing.T) {
	ctx := sdk.Context{}.WithContext(context.Background())
	const nonce = "test-nonce-123"
	claims := goodClaims(nonce)

	securePolicy := tee.Policy{
		GcpRootCertPem:  gcpRootCertificate,
		PolicyValues:    securePolicyValues,
		PolicyQuery:     securePolicyQuery,
		PolicyStructure: securePolicyStructure,
	}
	ok, err := tee.EvaluateOPAPolicy(ctx, securePolicy, claims, nonce)
	require.NoError(t, err)
	require.True(t, ok, "claims should satisfy the secure policy")

	insecurePolicy := tee.Policy{
		GcpRootCertPem:  gcpRootCertificate,
		PolicyValues:    insecurePolicyValues,
		PolicyQuery:     insecurePolicyQuery,
		PolicyStructure: insecurePolicyStructure,
	}
	ok, err = tee.EvaluateOPAPolicy(ctx, insecurePolicy, claims, nonce)
	require.NoError(t, err)
	require.False(t, ok, "claims should not satisfy the insecure policy")
}

// TestEvaluateOPAPolicyNonceMismatch confirms the verifier formats the raw
// nonce into the policy structure: a wrong nonce fails the secure policy.
func TestEvaluateOPAPolicyNonceMismatch(t *testing.T) {
	ctx := sdk.Context{}.WithContext(context.Background())
	claims := goodClaims("test-nonce-123")

	securePolicy := tee.Policy{
		PolicyValues:    securePolicyValues,
		PolicyQuery:     securePolicyQuery,
		PolicyStructure: securePolicyStructure,
	}
	ok, err := tee.EvaluateOPAPolicy(ctx, securePolicy, claims, "different-nonce")
	require.NoError(t, err)
	require.False(t, ok, "nonce mismatch should fail the policy")
}

// exampleResponse holds a real GCP attestation response. There is no fixture in
// the repo and a Google-signed token cannot be synthesized locally, so the
// full-PKI path test below is skip-gated until one is added.
var exampleResponse string

type ExampleResponse struct {
	Result struct {
		Token string `json:"token"`
		Nonce string `json:"nonce"`
	} `json:"result"`
}

// TestVerify exercises the full PKI cert-chain + signature path. It requires a
// real GCP-signed token fixture, which the repo does not have, so it is skipped.
func TestVerify(t *testing.T) {
	t.Skip("Requires a real response from GCP")

	var res ExampleResponse
	require.NoError(t, json.Unmarshal([]byte(exampleResponse), &res))

	policy := tee.Policy{
		GcpRootCertPem:  gcpRootCertificate,
		PolicyValues:    securePolicyValues,
		PolicyQuery:     securePolicyQuery,
		PolicyStructure: securePolicyStructure,
	}

	ctx := sdk.Context{}.WithContext(context.Background()).
		WithBlockTime(time.Date(2025, 9, 18, 9, 47, 0, 0, time.UTC))
	err := tee.NewVerifier().Verify(ctx, policy, res.Result.Nonce, res.Result.Token)
	require.NoError(t, err)
}
