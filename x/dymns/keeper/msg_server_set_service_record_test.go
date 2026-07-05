package keeper_test

import (
	"strings"
	"time"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// setupServiceRecordDymName persists a basic non-expired Dym-Name owned and controlled
// by the same account, optionally carrying the given configs.
func (s *KeeperTestSuite) setupServiceRecordDymName(name string, configs ...dymnstypes.DymNameConfig) dymnstypes.DymName {
	owner := testAddr(1).bech32()
	dymName := dymnstypes.DymName{
		Name:       name,
		Owner:      owner,
		Controller: owner,
		ExpireAt:   s.now.Add(time.Hour).Unix(),
		Configs:    configs,
	}
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymName))
	s.Require().NoError(s.dymNsKeeper.AfterDymNameOwnerChanged(s.ctx, name))
	s.Require().NoError(s.dymNsKeeper.AfterDymNameConfigChanged(s.ctx, name))
	return dymName
}

func (s *KeeperTestSuite) Test_msgServer_SetServiceRecord_AddUpdateDelete() {
	s.RefreshContext()

	const name = "agent"
	dymName := s.setupServiceRecordDymName(name)
	msgServer := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper)

	// add
	_, err := msgServer.SetServiceRecord(s.ctx, &dymnstypes.MsgSetServiceRecord{
		Name:       name,
		Controller: dymName.Controller,
		ServiceKey: "mcp",
		Value:      "https://mcp.example.com",
	})
	s.Require().NoError(err)
	s.Require().Equal("https://mcp.example.com", s.dymNsKeeper.GetServiceRecord(s.ctx, name, "mcp"))

	// update
	_, err = msgServer.SetServiceRecord(s.ctx, &dymnstypes.MsgSetServiceRecord{
		Name:       name,
		Controller: dymName.Controller,
		ServiceKey: "mcp",
		Value:      "https://mcp2.example.com",
	})
	s.Require().NoError(err)
	s.Require().Equal("https://mcp2.example.com", s.dymNsKeeper.GetServiceRecord(s.ctx, name, "mcp"))
	s.Require().Len(s.dymNsKeeper.GetServiceRecords(s.ctx, name), 1, "update must not duplicate")

	// delete (empty value)
	_, err = msgServer.SetServiceRecord(s.ctx, &dymnstypes.MsgSetServiceRecord{
		Name:       name,
		Controller: dymName.Controller,
		ServiceKey: "mcp",
		Value:      "",
	})
	s.Require().NoError(err)
	s.Require().Empty(s.dymNsKeeper.GetServiceRecord(s.ctx, name, "mcp"))
	s.Require().Empty(s.dymNsKeeper.GetServiceRecords(s.ctx, name))
}

func (s *KeeperTestSuite) Test_msgServer_SetServiceRecord_Rejections() {
	const name = "agent"

	tests := []struct {
		name        string
		msg         *dymnstypes.MsgSetServiceRecord
		wantErrPart string
	}{
		{
			name: "reject non-controller signer",
			msg: &dymnstypes.MsgSetServiceRecord{
				Name:       name,
				Controller: testAddr(9).bech32(),
				ServiceKey: "mcp",
				Value:      "https://mcp.example.com",
			},
			wantErrPart: "permission denied",
		},
		{
			name: "reject bad service key",
			msg: &dymnstypes.MsgSetServiceRecord{
				Name:       name,
				Controller: testAddr(1).bech32(),
				ServiceKey: "MCP_Bad",
				Value:      "https://mcp.example.com",
			},
			wantErrPart: "service config key is not valid",
		},
		{
			name: "reject value over 256 bytes",
			msg: &dymnstypes.MsgSetServiceRecord{
				Name:       name,
				Controller: testAddr(1).bech32(),
				ServiceKey: "mcp",
				Value:      strings.Repeat("a", dymnstypes.MaxServiceValueLength+1),
			},
			wantErrPart: "service config value is too long",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.RefreshContext()
			s.setupServiceRecordDymName(name)

			_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).SetServiceRecord(s.ctx, tt.msg)
			s.Require().Error(err)
			s.Require().Contains(err.Error(), tt.wantErrPart)
		})
	}
}

func (s *KeeperTestSuite) Test_msgServer_SetServiceRecord_NonExistentName() {
	s.RefreshContext()

	_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).SetServiceRecord(s.ctx, &dymnstypes.MsgSetServiceRecord{
		Name:       "ghost",
		Controller: testAddr(1).bech32(),
		ServiceKey: "mcp",
		Value:      "https://mcp.example.com",
	})
	s.Require().Error(err)
	s.Require().ErrorIs(err, gerrc.ErrNotFound)
}

func (s *KeeperTestSuite) Test_msgServer_SetServiceRecord_ExpiredName() {
	s.RefreshContext()

	const name = "agent"
	owner := testAddr(1).bech32()
	s.Require().NoError(s.dymNsKeeper.SetDymName(s.ctx, dymnstypes.DymName{
		Name:       name,
		Owner:      owner,
		Controller: owner,
		ExpireAt:   s.now.Add(-time.Hour).Unix(),
	}))

	_, err := dymnskeeper.NewMsgServerImpl(s.dymNsKeeper).SetServiceRecord(s.ctx, &dymnstypes.MsgSetServiceRecord{
		Name:       name,
		Controller: owner,
		ServiceKey: "mcp",
		Value:      "https://mcp.example.com",
	})
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "expired")
}

// Test_ServiceRecord_NotAnAddress verifies a service record is invisible to address
// resolution: it never contributes a resolution result, the endpoint string never leaks
// as an output address, and it never enters reverse-address lookups. The host-chain
// owner fallback (a separate, pre-existing mechanism) is intentionally not affected.
func (s *KeeperTestSuite) Test_ServiceRecord_NotAnAddress() {
	s.RefreshContext()

	const name = "agent"
	const endpoint = "https://mcp.example.com"
	const otherChainId = "another" // non-host, non-RollApp: no owner fallback applies
	dymName := s.setupServiceRecordDymName(name, dymnstypes.DymNameConfig{
		Type:  dymnstypes.DymNameConfigType_DCT_SERVICE,
		Path:  "mcp",
		Value: endpoint,
	})

	// On a non-host chain there is no fallback, so a service-only Dym-Name does not resolve.
	_, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name+"@"+otherChainId)
	s.Require().Error(err, "service record must not act as a resolution source")
	s.Require().ErrorIs(err, gerrc.ErrNotFound)

	// On the host chain the owner fallback resolves to the owner, never to the endpoint.
	resolved, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name+"@"+s.chainId)
	s.Require().NoError(err)
	s.Require().Equal(dymName.Owner, resolved)
	s.Require().NotEqual(endpoint, resolved, "endpoint string must never be returned as an address")

	// The endpoint string must never enter the configured-address reverse index.
	indexed, err := s.dymNsKeeper.GetDymNamesContainsConfiguredAddress(s.ctx, endpoint)
	s.Require().NoError(err)
	s.Require().Empty(indexed, "endpoint string must not be in the reverse-address index")
}

// Test_ServiceRecord_CoexistsWithNameConfig verifies a Dym-Name with both a DCT_NAME
// and a DCT_SERVICE record resolves its address exactly as before (regression) and
// returns the service record via the keeper getter.
func (s *KeeperTestSuite) Test_ServiceRecord_CoexistsWithNameConfig() {
	s.RefreshContext()

	const name = "agent"
	const endpoint = "https://mcp.example.com"
	owner := testAddr(1).bech32()

	s.setupServiceRecordDymName(name,
		dymnstypes.DymNameConfig{
			Type:  dymnstypes.DymNameConfigType_DCT_NAME,
			Value: owner,
		},
		dymnstypes.DymNameConfig{
			Type:  dymnstypes.DymNameConfigType_DCT_SERVICE,
			Path:  "mcp",
			Value: endpoint,
		},
	)

	resolved, err := s.dymNsKeeper.ResolveByDymNameAddress(s.ctx, name+"@"+s.chainId)
	s.Require().NoError(err)
	s.Require().Equal(owner, resolved)

	s.Require().Equal(endpoint, s.dymNsKeeper.GetServiceRecord(s.ctx, name, "mcp"))
}

func (s *KeeperTestSuite) Test_QueryServer_DymNameServices() {
	s.RefreshContext()

	const name = "agent"
	const endpoint = "https://mcp.example.com"
	s.setupServiceRecordDymName(name, dymnstypes.DymNameConfig{
		Type:  dymnstypes.DymNameConfigType_DCT_SERVICE,
		Path:  "mcp",
		Value: endpoint,
	})

	queryServer := dymnskeeper.NewQueryServerImpl(s.dymNsKeeper)

	servicesResp, err := queryServer.DymNameServices(s.ctx, &dymnstypes.QueryDymNameServicesRequest{Name: name})
	s.Require().NoError(err)
	s.Require().Equal([]dymnstypes.ServiceRecord{{ServiceKey: "mcp", Value: endpoint}}, servicesResp.Services)

	serviceResp, err := queryServer.DymNameService(s.ctx, &dymnstypes.QueryDymNameServiceRequest{Name: name, ServiceKey: "mcp"})
	s.Require().NoError(err)
	s.Require().Equal(endpoint, serviceResp.Value)

	// querying a missing key returns empty, not an error
	missingResp, err := queryServer.DymNameService(s.ctx, &dymnstypes.QueryDymNameServiceRequest{Name: name, ServiceKey: "a2a"})
	s.Require().NoError(err)
	s.Require().Empty(missingResp.Value)

	// querying a name with no service records returns empty list, not an error
	s.setupServiceRecordDymName("plain")
	emptyResp, err := queryServer.DymNameServices(s.ctx, &dymnstypes.QueryDymNameServicesRequest{Name: "plain"})
	s.Require().NoError(err)
	s.Require().Empty(emptyResp.Services)
}
