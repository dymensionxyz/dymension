package keeper

import (
	"crypto/sha256"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/golang-jwt/jwt/v5"
	"github.com/open-policy-agent/opa/v1/rego"
	"github.com/open-policy-agent/opa/v1/storage/inmem"
	"github.com/open-policy-agent/opa/v1/util"
)

//go:embed asset/tee_policy.rego
var opaPolicy string

/*
Validation logic is from https://github.com/GoogleCloudPlatform/confidential-space/blob/b6ade09bb9d3c7f39bb6af482ba71c7156184fd0/codelabs/health_data_analysis_codelab/src/uwear/workload.go#L1-L380 (https://codelabs.developers.google.com/confidential-space-pki?hl=en#0)
*/

func (k msgServer) validateAttestation(ctx sdk.Context, nonce, token string) error {
	// make sure the token really came from GCP
	jwt, err := k.validateAttestationAuthenticity(ctx, token)
	if err != nil {
		return errorsmod.Wrap(err, "validate PKI token")
	}

	// make the sure token actually certifies the non-tampered with computation
	err = k.validateAttestationIntegrity(ctx, jwt, nonce)
	if err != nil {
		return errorsmod.Wrap(err, "claims validation")
	}
	return nil
}

func (k Keeper) pemCert(ctx sdk.Context) (*x509.Certificate, error) {
	block, _ := pem.Decode(k.GetParams(ctx).TeeConfig.GcpRootCertPem)
	if block == nil {
		return nil, gerrc.ErrInvalidArgument.Wrap("parse pem block")
	}
	return x509.ParseCertificate(block.Bytes)
}

func (k msgServer) validateAttestationIntegrity(ctx sdk.Context, token jwt.Token, nonce string) error {
	policyData := k.GetParams(ctx).TeeConfig.PolicyData
	policyQuery := k.GetParams(ctx).TeeConfig.PolicyQuery

	authorized, err := evaluateOPAPolicy(ctx, token, nonce, policyData, policyQuery)
	if err != nil {
		return errorsmod.Wrap(err, "evaluate opa policy")
	}
	if !authorized {
		return gerrc.ErrFailedPrecondition.Wrap("tee policy not authorized")
	}
	return nil
}

// evaluateOPAPolicy returns boolean indicating if OPA policy is satisfied or not, or error if occurred
func evaluateOPAPolicy(ctx sdk.Context, token jwt.Token, nonce string, policyData string, policyQuery string) (bool, error) {
	var claims jwt.MapClaims
	var ok bool
	if claims, ok = token.Claims.(jwt.MapClaims); !ok {
		return false, gerrc.ErrInvalidArgument.Wrap("get claims from jwt")
	}

	module := fmt.Sprintf(opaPolicy, nonce)

	var json map[string]any
	err := util.UnmarshalJSON([]byte(policyData), &json)
	if err != nil {
		return false, errorsmod.Wrap(err, "unmarshal json")
	}
	store := inmem.NewFromObject(json)

	// Bind 'allow' to the value of the policy decision
	// Bind 'hw_verified', 'image_verified', 'audience_verified, 'nonce_verified' to their respective policy evaluations
	query, err := rego.New(
		rego.Query(policyQuery),                        // Argument 1 (Query string)
		rego.Store(store),                              // Argument 2 (Data store)
		rego.Module("confidential_space.rego", module), // Argument 3 (Policy module)
	).PrepareForEval(ctx)
	if err != nil {
		return false, errorsmod.Wrap(err, "create opa query")
	}

	results, err := query.Eval(ctx, rego.EvalInput(claims))

	if err != nil {
		return false, errorsmod.Wrap(err, "evaluate opa policy")
	} else if len(results) == 0 {
		return false, gerrc.ErrInvalidArgument.Wrap("undefined result from opa policy evaluation")
	} else if _, ok := results[0].Bindings["allow"].(bool); !ok {
		return false, gerrc.ErrInvalidArgument.Wrap("unexpected result type from opa policy")
	}

	result := results[0].Bindings["allow"]
	if result == true {
		return true, nil
	}
	return false, nil
}

// verifyCertificateChain verifies the certificate chain from leaf to root.
// It also checks that all certificate lifetimes are valid.
func verifyCertificateChain(certificates CertificateChain, now time.Time) error {
	// Additional check: Verify that all certificates in the cert chain are valid.
	// Note: The *x509.Certificate Verify method in golang already validates this but for other coding
	// languages it is important to make sure the certificate lifetimes are checked.
	if !isCertificateLifetimeValid(certificates.LeafCert, now) {
		return gerrc.ErrInvalidArgument.Wrap("leaf certificate lifetime not valid")
	}

	if !isCertificateLifetimeValid(certificates.IntermediateCert, now) {
		return gerrc.ErrInvalidArgument.Wrap("intermediate certificate lifetime not valid")
	}
	interPool := x509.NewCertPool()
	interPool.AddCert(certificates.IntermediateCert)

	if !isCertificateLifetimeValid(certificates.RootCert, now) {
		return gerrc.ErrInvalidArgument.Wrap("root certificate lifetime not valid")
	}
	rootPool := x509.NewCertPool()
	rootPool.AddCert(certificates.RootCert)

	_, err := certificates.LeafCert.Verify(x509.VerifyOptions{
		Intermediates: interPool,
		Roots:         rootPool,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	})
	if err != nil {
		return errorsmod.Wrap(err, "verify certificate chain")
	}

	return nil
}

func isCertificateLifetimeValid(certificate *x509.Certificate, now time.Time) bool {
	return now.After(certificate.NotBefore) && now.Before(certificate.NotAfter)
}

// compareCertificates compares two certificate fingerprints.
func compareCertificates(cert1 x509.Certificate, cert2 x509.Certificate) error {
	fingerprint1 := sha256.Sum256(cert1.Raw)
	fingerprint2 := sha256.Sum256(cert2.Raw)
	if fingerprint1 != fingerprint2 {
		return gerrc.ErrInvalidArgument.Wrap("certificate fingerprint mismatch")
	}
	return nil
}

// CertificateChain contains the certificates extracted from the x5c header.
type CertificateChain struct {
	LeafCert         *x509.Certificate
	IntermediateCert *x509.Certificate
	RootCert         *x509.Certificate
}

// extractCertificatesFromX5CHeader extracts the certificates from the given x5c header.
func extractCertificatesFromX5CHeader(x5cHeaders []any) (CertificateChain, error) {
	if x5cHeaders == nil {
		return CertificateChain{}, gerrc.ErrInvalidArgument.Wrap("x5c header not set")
	}

	x5c := []string{}
	for _, header := range x5cHeaders {
		asString, ok := header.(string)
		if !ok {
			return CertificateChain{}, gerrc.ErrInvalidArgument.Wrapf("x5c header is not a string")
		}
		x5c = append(x5c, asString)
	}

	// x5c header should have at least 3 certificates - leaf, intermediate and root
	if len(x5c) < 3 {
		return CertificateChain{}, gerrc.ErrInvalidArgument.Wrapf("not enough certificates in x5c header expected 3 got %v", len(x5c))
	}

	leafCert, err := decodeAndParseDERCertificate(x5c[0])
	if err != nil {
		return CertificateChain{}, errorsmod.Wrap(err, "parse leaf certificate")
	}

	intermediateCert, err := decodeAndParseDERCertificate(x5c[1])
	if err != nil {
		return CertificateChain{}, errorsmod.Wrap(err, "parse intermediate certificate")
	}

	rootCert, err := decodeAndParseDERCertificate(x5c[2])
	if err != nil {
		return CertificateChain{}, errorsmod.Wrap(err, "parse root certificate")
	}

	certificates := CertificateChain{
		LeafCert:         leafCert,
		IntermediateCert: intermediateCert,
		RootCert:         rootCert,
	}
	return certificates, nil
}

// decodeAndParseDERCertificate decodes the given DER certificate string and parses it into an x509 certificate.
func decodeAndParseDERCertificate(certificate string) (*x509.Certificate, error) {
	bytes, _ := base64.StdEncoding.DecodeString(certificate)

	cert, err := x509.ParseCertificate(bytes)
	if err != nil {
		return nil, errorsmod.Wrap(err, "parse certificate")
	}

	return cert, nil
}

// validatePKIToken validates the PKI token returned from the attestation service.
// It verifies the token the certificate chain and that the token is signed by Google
// Returns a jwt.Token or returns an error if invalid.
func (k Keeper) validateAttestationAuthenticity(ctx sdk.Context, attestationToken string) (jwt.Token, error) {
	// IMPORTANT: The attestation token should be considered untrusted until the certificate chain and
	// the signature is verified.
	storedRootCert, err := k.pemCert(ctx)
	if err != nil {
		return jwt.Token{}, errorsmod.Wrap(err, "decode and parse root certificate")
	}

	jwtHeaders, err := extractJWTHeaders(attestationToken)
	if err != nil {
		return jwt.Token{}, errorsmod.Wrap(err, "extract jwt headers")
	}

	if jwtHeaders["alg"] != "RS256" {
		return jwt.Token{}, gerrc.ErrInvalidArgument.Wrapf("got alg %v want rs256", jwtHeaders["alg"])
	}

	// Additional Check: Validate the ALG in the header matches the certificate SPKI.
	// https://datatracker.ietf.org/doc/html/rfc5280#section-4.1.2.7
	// This is included in golangs jwt.Parse function

	x5cHeaders, ok := jwtHeaders["x5c"].([]any)
	if !ok {
		return jwt.Token{}, gerrc.ErrInvalidArgument.Wrapf("x5c header not set")
	}
	certificates, err := extractCertificatesFromX5CHeader(x5cHeaders)
	if err != nil {
		return jwt.Token{}, errorsmod.Wrap(err, "extract certificates from x5c header")
	}

	// Verify the leaf certificate signature algorithm is an RSA key
	if certificates.LeafCert.SignatureAlgorithm != x509.SHA256WithRSA {
		return jwt.Token{}, gerrc.ErrInvalidArgument.Wrap("leaf certificate signature algorithm is not sha256withrsa")
	}

	// Verify the leaf certificate public key algorithm is RSA
	if certificates.LeafCert.PublicKeyAlgorithm != x509.RSA {
		return jwt.Token{}, gerrc.ErrInvalidArgument.Wrap("leaf certificate public key algorithm is not rsa")
	}

	// Verify the storedRootCertificate is the same as the root certificate returned in the token
	// storedRootCertificate is downloaded from the confidential computing well known endpoint
	// https://confidentialcomputing.googleapis.com/.well-known/attestation-pki-root
	err = compareCertificates(*storedRootCert, *certificates.RootCert)
	if err != nil {
		return jwt.Token{}, errorsmod.Wrap(err, "verify certificate chain")
	}

	err = verifyCertificateChain(certificates, ctx.BlockTime())
	if err != nil {
		return jwt.Token{}, errorsmod.Wrap(err, "verify x5c chain")
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
		return nil, errorsmod.Wrap(err, "parse claims token")
	}

	return parsedToken.Header, nil
}
