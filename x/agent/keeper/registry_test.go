package keeper_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/agent/keeper"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/dymensionxyz/dymension/v3/x/common/tee"
)

type RegistryTestSuite struct {
	apptesting.KeeperTestHelper

	msgServer types.MsgServer
}

func TestRegistryTestSuite(t *testing.T) {
	suite.Run(t, new(RegistryTestSuite))
}

func (s *RegistryTestSuite) SetupTest() {
	app := apptesting.Setup(s.T())
	s.App = app
	s.Ctx = app.NewContext(false)
	s.msgServer = keeper.NewMsgServerImpl(*s.App.AgentKeeper)
}

// validPolicy returns a policy that passes ValidateBasic: a parseable self-signed
// cert plus non-empty rego fields.
func validPolicy(s *RegistryTestSuite) tee.Policy {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	s.Require().NoError(err)
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "test"}}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	s.Require().NoError(err)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	return tee.Policy{
		GcpRootCertPem:  string(certPEM),
		PolicyValues:    `{"k":"v"}`,
		PolicyQuery:     "data.x.allow",
		PolicyStructure: "package x\nallow = true",
	}
}

func (s *RegistryTestSuite) fundedOwner() sdk.AccAddress {
	_, _, addr := testdata.KeyTestPubAddr()
	fee, err := s.App.AgentKeeper.AgentRegistrationFee(s.Ctx)
	s.Require().NoError(err)
	s.FundAcc(addr, sdk.NewCoins(sdk.NewCoin(fee.Denom, fee.Amount.MulRaw(10))))
	return addr
}

func (s *RegistryTestSuite) TestRegisterAgent_HappyPathAndFeeBurned() {
	owner := s.fundedOwner()
	fee, err := s.App.AgentKeeper.AgentRegistrationFee(s.Ctx)
	s.Require().NoError(err)
	balBefore := s.App.BankKeeper.GetBalance(s.Ctx, owner, fee.Denom)
	supplyBefore := s.App.BankKeeper.GetSupply(s.Ctx, fee.Denom)

	_, err = s.msgServer.RegisterAgent(s.Ctx, types.NewMsgRegisterAgent(owner.String(), "agent-1", validPolicy(s)))
	s.Require().NoError(err)

	agent, found := s.App.AgentKeeper.GetAgent(s.Ctx, "agent-1")
	s.Require().True(found)
	s.Require().Equal(owner.String(), agent.Owner)
	s.Require().True(agent.Active)
	s.Require().Equal(uint64(0), agent.ActionSeq)

	// fee debited from owner and burned (total supply reduced)
	balAfter := s.App.BankKeeper.GetBalance(s.Ctx, owner, fee.Denom)
	s.Require().Equal(balBefore.Amount.Sub(fee.Amount), balAfter.Amount)
	supplyAfter := s.App.BankKeeper.GetSupply(s.Ctx, fee.Denom)
	s.Require().Equal(supplyBefore.Amount.Sub(fee.Amount), supplyAfter.Amount)
}

func (s *RegistryTestSuite) TestRegisterAgent_DuplicateRejected() {
	owner := s.fundedOwner()
	_, err := s.msgServer.RegisterAgent(s.Ctx, types.NewMsgRegisterAgent(owner.String(), "dup", validPolicy(s)))
	s.Require().NoError(err)

	_, err = s.msgServer.RegisterAgent(s.Ctx, types.NewMsgRegisterAgent(owner.String(), "dup", validPolicy(s)))
	s.Require().ErrorIs(err, types.ErrAgentExists)
}

func (s *RegistryTestSuite) TestDeactivateAgent_NonOwnerRejected() {
	owner := s.fundedOwner()
	_, err := s.msgServer.RegisterAgent(s.Ctx, types.NewMsgRegisterAgent(owner.String(), "agent-x", validPolicy(s)))
	s.Require().NoError(err)

	_, _, other := testdata.KeyTestPubAddr()
	_, err = s.msgServer.DeactivateAgent(s.Ctx, types.NewMsgDeactivateAgent(other.String(), "agent-x"))
	s.Require().ErrorIs(err, types.ErrUnauthorized)

	agent, found := s.App.AgentKeeper.GetAgent(s.Ctx, "agent-x")
	s.Require().True(found)
	s.Require().True(agent.Active)
}

func (s *RegistryTestSuite) TestDeactivateAgent_OwnerSucceeds() {
	owner := s.fundedOwner()
	_, err := s.msgServer.RegisterAgent(s.Ctx, types.NewMsgRegisterAgent(owner.String(), "agent-y", validPolicy(s)))
	s.Require().NoError(err)

	_, err = s.msgServer.DeactivateAgent(s.Ctx, types.NewMsgDeactivateAgent(owner.String(), "agent-y"))
	s.Require().NoError(err)

	agent, found := s.App.AgentKeeper.GetAgent(s.Ctx, "agent-y")
	s.Require().True(found)
	s.Require().False(agent.Active)
}
