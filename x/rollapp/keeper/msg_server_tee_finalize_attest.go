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
		return nil, fmt.Errorf("parse PEM block")
	}
	return x509.ParseCertificate(block.Bytes)
}

func (k msgServer) validateAttestationIntegrity(ctx sdk.Context, token jwt.Token, nonce string) error {
	policyData := k.GetParams(ctx).TeeConfig.PolicyData
	policyQuery := k.GetParams(ctx).TeeConfig.PolicyQuery

	authorized, err := evaluateOPAPolicy(ctx, token, nonce, policyData, policyQuery)
	if err != nil {
		return fmt.Errorf("evaluate OPA policy: %w", err)
	}
	if !authorized {
		return fmt.Errorf("tee policy not authorized")
	}
	return nil
}

// evaluateOPAPolicy returns boolean indicating if OPA policy is satisfied or not, or error if occurred
func evaluateOPAPolicy(ctx sdk.Context, token jwt.Token, nonce string, policyData string, policyQuery string) (bool, error) {
	var claims jwt.MapClaims
	var ok bool
	if claims, ok = token.Claims.(jwt.MapClaims); !ok {
		return false, fmt.Errorf(" get the claims from the JWT")
	}

	module := fmt.Sprintf(opaPolicy, nonce)

	var json map[string]any
	err := util.UnmarshalJSON([]byte(policyData), &json)
	if err != nil {
		return false, fmt.Errorf("unmarshal JSON: %w", err)
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
		fmt.Printf("Error creating query: %v\n", err)
		return false, err
	}

	fmt.Println("Performing OPA query evaluation...")
	results, err := query.Eval(ctx, rego.EvalInput(claims))

	if err != nil {
		fmt.Printf("Error evaluating OPA policy: %v\n", err)
		return false, err
	} else if len(results) == 0 {
		fmt.Println("Undefined result from evaluating OPA policy")
		return false, err
	} else if result, ok := results[0].Bindings["allow"].(bool); !ok {
		fmt.Printf("Unexpected result type: %v\n", ok)
		fmt.Printf("Result: %+v\n", result)
		return false, err
	}

	fmt.Println("OPA policy evaluation completed.")

	fmt.Println("OPA policy result values:")
	for key, value := range results[0].Bindings {
		fmt.Printf("[ %s ]: %v\n", key, value)
	}
	result := results[0].Bindings["allow"]
	if result == true {
		fmt.Println("Policy check PASSED")
		return true, nil
	}
	fmt.Println("Policy check FAILED")
	return false, nil
}

// verifyCertificateChain verifies the certificate chain from leaf to root.
// It also checks that all certificate lifetimes are valid.
func verifyCertificateChain(certificates CertificateChain, now time.Time) error {
	// Additional check: Verify that all certificates in the cert chain are valid.
	// Note: The *x509.Certificate Verify method in golang already validates this but for other coding
	// languages it is important to make sure the certificate lifetimes are checked.
	if isCertificateLifetimeValid(certificates.LeafCert, now) {
		return fmt.Errorf("leaf certificate is not valid")
	}

	if isCertificateLifetimeValid(certificates.IntermediateCert, now) {
		return fmt.Errorf("intermediate certificate is not valid")
	}
	interPool := x509.NewCertPool()
	interPool.AddCert(certificates.IntermediateCert)

	if isCertificateLifetimeValid(certificates.RootCert, now) {
		return fmt.Errorf("root certificate is not valid")
	}
	rootPool := x509.NewCertPool()
	rootPool.AddCert(certificates.RootCert)

	_, err := certificates.LeafCert.Verify(x509.VerifyOptions{
		Intermediates: interPool,
		Roots:         rootPool,
		KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
	})
	if err != nil {
		return fmt.Errorf(" verify certificate chain: %v", err)
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
		return fmt.Errorf("certificate fingerprint mismatch")
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
		return CertificateChain{}, fmt.Errorf("x5c header not set")
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
		return CertificateChain{}, fmt.Errorf("not enough certificates in x5c header, expected 3 certificates, but got %v", len(x5c))
	}

	leafCert, err := decodeAndParseDERCertificate(x5c[0])
	if err != nil {
		return CertificateChain{}, fmt.Errorf("cannot parse leaf certificate: %v", err)
	}

	intermediateCert, err := decodeAndParseDERCertificate(x5c[1])
	if err != nil {
		return CertificateChain{}, fmt.Errorf("cannot parse intermediate certificate: %v", err)
	}

	rootCert, err := decodeAndParseDERCertificate(x5c[2])
	if err != nil {
		return CertificateChain{}, fmt.Errorf("cannot parse root certificate: %v", err)
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
		return nil, fmt.Errorf("cannot parse certificate: %v", err)
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
		return jwt.Token{}, fmt.Errorf("DecodeAndParsePEMCertificate(string) -  decode and parse root certificate: %w", err)
	}

	jwtHeaders, err := extractJWTHeaders(attestationToken)
	if err != nil {
		return jwt.Token{}, fmt.Errorf("ExtractJWTHeaders(token) -  extract JWT headers: %w", err)
	}

	if jwtHeaders["alg"] != "RS256" {
		return jwt.Token{}, fmt.Errorf("ValidatePKIToken(attestationToken, ekm) - got Alg: %v, want: %v", jwtHeaders["alg"], "RS256")
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
		return jwt.Token{}, fmt.Errorf(" verify certificate chain: %w", err)
	}

	err = verifyCertificateChain(certificates, ctx.BlockTime())
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
		return nil, fmt.Errorf(" parse claims token: %v", err)
	}

	return parsedToken.Header, nil
}
