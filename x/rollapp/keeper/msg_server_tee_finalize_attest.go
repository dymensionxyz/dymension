package keeper

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/storage/inmem"
)

//go:embed asset/tee_attestation_policy.rego
var opaPolicy string

// FastFinalizeWithTEE handles TEE attestation-based fast finalization
func (k msgServer) FastFinalizeWithTEE(goCtx context.Context, msg *types.MsgFastFinalizeWithTEE) (*types.MsgFastFinalizeWithTEEResponse, error) {
	if err := msg.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	params := k.GetParams(ctx)
	teeConfig := params.TeeConfig

	if !teeConfig.Enabled {
		return nil, gerrc.ErrFailedPrecondition.Wrap("TEE fast finalization is not enabled")
	}

	rollapp := msg.Nonce.RollappId
	ix := msg.StateIndex

	_, found := k.GetRollapp(ctx, rollapp)
	if !found {
		return nil, gerrc.ErrNotFound.Wrapf("rollapp: %s", rollapp)
	}

	proposer := k.SequencerK.GetProposer(ctx, rollapp)

	if proposer.Address != msg.Creator {
		return nil, gerrc.ErrPermissionDenied.Wrapf("only active sequencer can submit TEE attestation: expected %s, got %s",
			proposer.Address, msg.Creator)
	}

	if k.IsFinalizedIndex(ctx, rollapp, ix) {
		return nil, gerrc.ErrOutOfRange.Wrapf("state index is already finalized")
	}

	info, found := k.GetStateInfo(ctx, rollapp, ix)
	if !found {
		return nil, gerrc.ErrNotFound.Wrapf("state info for rollapp: %s", rollapp)
	}

	if info.GetLatestHeight() != uint64(msg.Nonce.Height) {
		return nil, gerrc.ErrInvalidArgument.Wrapf("height index mismatch")
	}

	bd, _ := info.LastBlockDescriptor()

	if !bytes.Equal(bd.StateRoot, msg.Nonce.StateRoot) {
		return nil, gerrc.ErrInvalidArgument.Wrapf("state root mismatch")
	}

	err := k.validateAttestation(ctx, teeConfig.GcpRootCertPem, msg.Nonce.Hash(), msg.AttestationToken)
	if err != nil {
		return nil, errorsmod.Wrap(err, "validate attestation")
	}

	err = k.FastFinalizeRollappStatesUntilStateIndex(ctx, rollapp, ix)
	if err != nil {
		return nil, errorsmod.Wrap(err, "fast finalize states")
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTEEFastFinalization,
			sdk.NewAttribute(types.AttributeKeyRollappId, rollapp),
			sdk.NewAttribute(types.AttributeKeyStateIndex, fmt.Sprintf("%d", ix)),
		),
	)

	return &types.MsgFastFinalizeWithTEEResponse{}, nil
}

func (k msgServer) validateAttestation(ctx sdk.Context, nonce, token string) error {
	// make sure the token really came from GCP
	jwt, err := k.validateAttestationAuthenticity(ctx, token)
	if err != nil {
		return errorsmod.Wrap(err, "validate PKI token")
	}

	// make the sure token actually certifies the non-tampered with computation
	err = k.validateAttestationIntegrity(ctx, *jwt, nonce)
	if err != nil {
		return errorsmod.Wrap(err, "claims validation")
	}
	return nil
}

// validateAttestationAuthenticity validates the PKI token returned from the attestation service
func (k msgServer) validateAttestationAuthenticity(ctx sdk.Context, attestationToken string) (*jwt.Token, error) {
	// Parse the token without verification first to get the x5c header
	unverifiedToken, _, err := jwt.NewParser().ParseUnverified(attestationToken, jwt.MapClaims{})
	if err != nil {
		return nil, fmt.Errorf("parse unverified token: %w", err)
	}

	if unverifiedToken.Header["alg"] != "RS256" {
		return nil, fmt.Errorf("unexpected signing method: %v", unverifiedToken.Header["alg"])
	}

	x5cInterface, ok := unverifiedToken.Header["x5c"]
	if !ok {
		return nil, fmt.Errorf("x5c header not found in token")
	}

	x5c, ok := x5cInterface.([]interface{})
	if !ok || len(x5c) < 3 {
		return nil, fmt.Errorf("invalid x5c header format or insufficient certificates")
	}

	var certs []*x509.Certificate
	for i, certStr := range x5c {
		certDER, err := base64.StdEncoding.DecodeString(certStr.(string))
		if err != nil {
			return nil, fmt.Errorf("decode certificate %d: %w", i, err)
		}

		cert, err := x509.ParseCertificate(certDER)
		if err != nil {
			return nil, fmt.Errorf("parse certificate %d: %w", i, err)
		}
		certs = append(certs, cert)
	}

	block, _ := pem.Decode(pemCert)
	if block == nil {
		return nil, fmt.Errorf("parse PEM block")
	}

	rootCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse root certificate: %w", err)
	}

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

	providedRootFingerprint := sha256.Sum256(certs[len(certs)-1].Raw)
	expectedRootFingerprint := sha256.Sum256(rootCert.Raw)
	if providedRootFingerprint != expectedRootFingerprint {
		return nil, fmt.Errorf("root certificate fingerprint mismatch")
	}

	token, err := jwt.Parse(attestationToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		// Return the leaf certificate's public key for JWT verification
		return certs[0].PublicKey.(*rsa.PublicKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse/validate JWT token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid JWT token")
	}

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

func validatePKIToken(attestationToken string) (jwt.Token, error) {
	jwtHeaders, err := extractJWTHeaders(attestationToken)
	if err != nil {
		return jwt.Token{}, fmt.Errorf("ExtractJWTHeaders(token) - failed to extract JWT headers: %w", err)
	}

	if jwtHeaders["alg"] != "RS256" {
		return jwt.Token{}, fmt.Errorf("ValidatePKIToken(attestationToken, ekm) - got Alg: %v, want: %v", jwtHeaders["alg"], "RS256")
	}

	// Additional Check: Validate the ALG in the header matches the certificate SPKI.
	// https://datatracker.ietf.org/doc/html/rfc5280#section-4.1.2.7
	// This is included in golangs jwt.Parse function

	x5cHeaders := jwtHeaders["x5c"].([]any)
	certificates, err := extractCertificatesFromX5CHeader(x5cHeaders)
	if err != nil {
		return jwt.Token{}, fmt.Errorf("ExtractCertificatesFromX5CHeader(x5cHeaders) returned error: %w", err)
	}

	// Verify the leaf certificate signature algorithm is an RSA key
	if certificates.LeafCert.SignatureAlgorithm != x509.SHA256WithRSA {
		return jwt.Token{}, fmt.Errorf("leaf certificate signature algorithm is not SHA256WithRSA")
	}

	// Verify the leaf certificate public key algorithm is RSA
	if certificates.LeafCert.PublicKeyAlgorithm != x509.RSA {
		return jwt.Token{}, fmt.Errorf("leaf certificate public key algorithm is not RSA")
	}

	// Verify the storedRootCertificate is the same as the root certificate returned in the token
	// storedRootCertificate is downloaded from the confidential computing well known endpoint
	// https://confidentialcomputing.googleapis.com/.well-known/attestation-pki-root
	err = compareCertificates(*storedRootCert, *certificates.RootCert)
	if err != nil {
		return jwt.Token{}, fmt.Errorf("failed to verify certificate chain: %w", err)
	}

	err = verifyCertificateChain(certificates)
	if err != nil {
		return jwt.Token{}, fmt.Errorf("VerifyCertificateChain(CertificateChain) - error verifying x5c chain: %v", err)
	}

	keyFunc := func(token *jwt.Token) (any, error) {
		return certificates.LeafCert.PublicKey, nil
	}

	verifiedJWT, err := jwt.Parse(attestationToken, keyFunc)
	return *verifiedJWT, err
}

// extractJWTHeaders parses the JWT and returns the headers.
func extractJWTHeaders(token string) (map[string]any, error) {
	parser := &jwt.Parser{}
	// The claims returned from the token are unverified at this point
	// Do not use the claims until the algorithm, certificate chain verification and root certificate
	// comparison is successful
	unverifiedClaims := &jwt.MapClaims{}
	parsedToken, _, err := parser.ParseUnverified(token, unverifiedClaims)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse claims token: %v", err)
	}

	return parsedToken.Header, nil
}

// validateAttestationIntegrity validates the claims using OPA policy
func (k msgServer) validateAttestationIntegrity(ctx sdk.Context, token jwt.Token, expectedNonce string) error {
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return fmt.Errorf("extract JWT claims")
	}

	policyData := map[string]interface{}{
		"allowed_image_digests": "",
		"expected_nonce":        expectedNonce,
	}
	store := inmem.NewFromObject(policyData)

	query, err := rego.New(
		rego.Query("data.tee_attestation.allow"),
		rego.Store(store),
		rego.Module("tee_attestation.rego", opaPolicy),
	).PrepareForEval(ctx.Context())
	if err != nil {
		return fmt.Errorf("creating OPA query: %w", err)
	}

	results, err := query.Eval(ctx.Context(), rego.EvalInput(claims))
	if err != nil {
		return fmt.Errorf("evaluating OPA policy: %w", err)
	}

	if len(results) == 0 {
		return fmt.Errorf("undefined result from OPA policy evaluation")
	}

	if allowed, ok := results[0].Expressions[0].Value.(bool); !ok || !allowed {
		return fmt.Errorf("TEE attestation claims failed policy validation")
	}

	return nil
}

func (k Keeper) pemCert(ctx sdk.Context) (*x509.Certificate, error) {
	block, _ := pem.Decode(k.GetParams(ctx).TeeConfig.GcpRootCertPem)
	if block == nil {
		return nil, fmt.Errorf("parse PEM block")
	}
	return x509.ParseCertificate(block.Bytes)
}
