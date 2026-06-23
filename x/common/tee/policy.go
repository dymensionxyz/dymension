package tee

import (
	"crypto/x509"
	"encoding/pem"

	errorsmod "cosmossdk.io/errors"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	opastorage "github.com/open-policy-agent/opa/v1/storage"
	"github.com/open-policy-agent/opa/v1/storage/inmem"
	"github.com/open-policy-agent/opa/v1/util"
)

func (p Policy) PemCert() (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(p.GcpRootCertPem))
	if block == nil {
		return nil, gerrc.ErrInvalidArgument.Wrap("parse pem block")
	}
	return x509.ParseCertificate(block.Bytes)
}

func (p Policy) PolicyValuesStore() (opastorage.Store, error) {
	var json map[string]any
	err := util.UnmarshalJSON([]byte(p.PolicyValues), &json)
	if err != nil {
		return nil, errorsmod.Wrap(err, "unmarshal json")
	}
	return inmem.NewFromObject(json), nil
}
