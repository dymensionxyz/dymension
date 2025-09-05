package keeper

import (
	"context"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/storage/inmem"
)

// opaPolicy defines the OPA policy for validating TEE attestation claims
const opaPolicy = `
	package tee_attestation

	import rego.v1

	default allow := false
	default hw_verified := false
	default image_digest_verified := false
	default nonce_verified := false
	default issuer_verified := false
	default secboot_verified := false
	default sw_name_verified := false

	allow if {
		hw_verified
		image_digest_verified
		nonce_verified
		issuer_verified
		secboot_verified
		sw_name_verified
	}

	hw_verified if input.hwmodel in ["GCP-TDX", "TDX"]
	image_digest_verified if input.submods.container.image_digest in data.allowed_image_digests
	issuer_verified if input.iss == "https://confidentialcomputing.googleapis.com"
	secboot_verified if input.secboot == true
	sw_name_verified if input.swname == "CONFIDENTIAL_SPACE"
	nonce_verified if {
		input.eat_nonce == data.expected_nonce
	}
`

// FastFinalizeWithTEE handles TEE attestation-based fast finalization
func (k msgServer) FastFinalizeWithTEE(goCtx context.Context, msg *types.MsgFastFinalizeWithTEE) (*types.MsgFastFinalizeWithTEEResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Get TEE config
	params := k.GetParams(ctx)
	teeConfig := params.TeeConfig

	if !teeConfig.Enabled {
		return nil, gerrc.ErrFailedPrecondition.Wrap("TEE fast finalization is not enabled")
	}

	// Verify the creator is the active sequencer for the rollapp
	_, found := k.GetRollapp(ctx, msg.RollappId)
	if !found {
		return nil, gerrc.ErrNotFound.Wrapf("rollapp: %s", msg.RollappId)
	}

	// Get the active sequencer
	proposer := k.SequencerK.GetProposer(ctx, msg.RollappId)
	if proposer.Sentinel() {
		return nil, gerrc.ErrNotFound.Wrap("no active sequencer for rollapp")
	}

	if proposer.Address != msg.Creator {
		return nil, gerrc.ErrPermissionDenied.Wrapf("only active sequencer can submit TEE attestation: expected %s, got %s",
			proposer.Address, msg.Creator)
	}

	// Verify the PEM certificate SHA1 matches the configured value
	pemSHA1 := sha1.Sum(msg.PemCert)
	pemSHA1Hex := hex.EncodeToString(pemSHA1[:])

	if !strings.EqualFold(pemSHA1Hex, teeConfig.GcpRootCertSha1) {
		return nil, gerrc.ErrInvalidArgument.Wrapf("PEM cert SHA1 mismatch: expected %s, got %s",
			teeConfig.GcpRootCertSha1, pemSHA1Hex)
	}

	// Parse and validate the JWT token with certificate chain
	token, err := k.validatePKIToken(ctx, msg.AttestationToken, msg.PemCert)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to validate PKI token")
	}

	// Verify the state index matches what's in the nonce
	if msg.StateIndex != msg.Nonce.StateIndex {
		return nil, gerrc.ErrInvalidArgument.Wrap("state index mismatch between message and nonce")
	}

	// Calculate expected nonce
	expectedNonce := k.calculateTEENonce(msg.RollappId, msg.Nonce)

	// Validate claims against OPA policy
	err = k.validateClaimsWithOPA(ctx, *token, expectedNonce, teeConfig)
	if err != nil {
		return nil, errorsmod.Wrap(err, "claims validation failed")
	}

	// Fast finalize the states up to the given state index
	err = k.FastFinalizeRollappStatesUntilStateIndex(ctx, msg.RollappId, msg.StateIndex)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to fast finalize states")
	}

	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTEEFastFinalization,
			sdk.NewAttribute(types.AttributeKeyRollappId, msg.RollappId),
			sdk.NewAttribute(types.AttributeKeyStateIndex, fmt.Sprintf("%d", msg.StateIndex)),
		),
	)

	return &types.MsgFastFinalizeWithTEEResponse{}, nil
}

// validatePKIToken validates the PKI token returned from the attestation service
func (k msgServer) validatePKIToken(ctx sdk.Context, attestationToken string, pemCert []byte) (*jwt.Token, error) {
	// Parse the token without verification first to get the x5c header
	unverifiedToken, _, err := jwt.NewParser().ParseUnverified(attestationToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("parse unverified token: %w", err)
	}

	// Extract x5c certificate chain from headers
	x5cInterface, ok := unverifiedToken.Header["x5c"]
	if !ok {
		return nil, fmt.Errorf("x5c header not found in token")
	}

	x5c, ok := x5cInterface.([]interface{})
	if !ok || len(x5c) < 3 {
		return nil, fmt.Errorf("invalid x5c header format or insufficient certificates")
	}

	// Parse the certificate chain
	var certs []*x509.Certificate
	for i, certStr := range x5c {
		certDER, err := base64.StdEncoding.DecodeString(certStr.(string))
		if err != nil {
			return nil, fmt.Errorf("failed to decode certificate %d: %w", i, err)
		}

		cert, err := x509.ParseCertificate(certDER)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate %d: %w", i, err)
		}
		certs = append(certs, cert)
	}

	// Parse the PEM root certificate
	block, _ := pem.Decode(pemCert)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block")
	}

	rootCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse root certificate: %w", err)
	}

	// Verify the certificate chain
	rootPool := x509.NewCertPool()
	rootPool.AddCert(rootCert)

	intermediatePool := x509.NewCertPool()
	if len(certs) > 2 {
		for i := 1; i < len(certs)-1; i++ {
			intermediatePool.AddCert(certs[i])
		}
	}

	opts := x509.VerifyOptions{
		Roots:         rootPool,
		Intermediates: intermediatePool,
		CurrentTime:   ctx.BlockTime(),
	}

	if _, err := certs[0].Verify(opts); err != nil {
		return nil, fmt.Errorf("certificate chain verification failed: %w", err)
	}

	// Compare root certificate fingerprints
	providedRootFingerprint := sha256.Sum256(certs[len(certs)-1].Raw)
	expectedRootFingerprint := sha256.Sum256(rootCert.Raw)
	if providedRootFingerprint != expectedRootFingerprint {
		return nil, fmt.Errorf("root certificate fingerprint mismatch")
	}

	// Now parse and verify the JWT with the leaf certificate's public key
	token, err := jwt.Parse(attestationToken, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the leaf certificate's public key for JWT verification
		return certs[0].PublicKey.(*rsa.PublicKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse/validate JWT token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

	// Check token expiration
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"].(float64); ok {
			expTime := time.Unix(int64(exp), 0)
			if ctx.BlockTime().After(expTime) {
				return nil, fmt.Errorf("token has expired")
			}
		}
	}

	return token, nil
}

// validateClaimsWithOPA validates the claims using OPA policy
func (k msgServer) validateClaimsWithOPA(ctx sdk.Context, token jwt.Token, expectedNonce string, teeConfig types.TEEConfig) error {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("failed to extract JWT claims")
	}

	// Prepare OPA data store with allowed values
	policyData := map[string]interface{}{
		"allowed_image_digests": teeConfig.AllowedImageDigests,
		"expected_nonce":        expectedNonce,
	}
	store := inmem.NewFromObject(policyData)

	// Prepare and evaluate OPA query
	query, err := rego.New(
		rego.Query("data.tee_attestation.allow"),
		rego.Store(store),
		rego.Module("tee_attestation.rego", opaPolicy),
	).PrepareForEval(ctx.Context())
	if err != nil {
		return fmt.Errorf("error creating OPA query: %w", err)
	}

	results, err := query.Eval(ctx.Context(), rego.EvalInput(claims))
	if err != nil {
		return fmt.Errorf("error evaluating OPA policy: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("undefined result from OPA policy evaluation")
	}

	if allowed, ok := results[0].Expressions[0].Value.(bool); !ok || !allowed {
		return fmt.Errorf("TEE attestation claims failed policy validation")
	}

	return nil
}

// calculateTEENonce calculates the expected nonce hash
func (k msgServer) calculateTEENonce(rollappID string, nonce types.TEENonce) string {
	// Create a deterministic string from the nonce data
	nonceData := fmt.Sprintf("%d:%s:%x", nonce.StateIndex, rollappID, nonce.LastBlockHash)

	// Calculate SHA256 hash
	hash := sha256.Sum256([]byte(nonceData))
	return hex.EncodeToString(hash[:])
}
