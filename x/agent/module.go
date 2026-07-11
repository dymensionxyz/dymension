package agent

import (
	"context"
	"encoding/json"
	"fmt"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/dymension/v3/x/agent/client/cli"
	"github.com/dymensionxyz/dymension/v3/x/agent/keeper"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.HasGenesis     = AppModule{}
	_ module.HasServices    = AppModule{}

	_ appmodule.AppModule = AppModule{}
)

// ----------------------------------------------------------------------------
// AppModuleBasic
// ----------------------------------------------------------------------------

type AppModuleBasic struct {
	cdc codec.BinaryCodec
}

func NewAppModuleBasic(cdc codec.BinaryCodec) AppModuleBasic {
	return AppModuleBasic{cdc: cdc}
}

func (AppModuleBasic) Name() string {
	return types.ModuleName
}

func (AppModuleBasic) RegisterCodec(cdc *codec.LegacyAmino) {
	types.RegisterCodec(cdc)
}

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterCodec(cdc)
}

func (a AppModuleBasic) RegisterInterfaces(reg cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(reg)
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, config client.TxEncodingConfig, bz json.RawMessage) error {
	var genState types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &genState); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return genState.Validate()
}

func (AppModuleBasic) RegisterRESTRoutes(clientCtx client.Context, rtr *mux.Router) {
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// nolint: errcheck, gosec
	types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

// ----------------------------------------------------------------------------
// AppModule
// ----------------------------------------------------------------------------

type AppModule struct {
	AppModuleBasic

	keeper *keeper.Keeper
}

func NewAppModule(
	cdc codec.Codec,
	keeper *keeper.Keeper,
) AppModule {
	return AppModule{
		AppModuleBasic: NewAppModuleBasic(cdc),
		keeper:         keeper,
	}
}

func (am AppModule) IsAppModule() {}

func (am AppModule) IsOnePerModuleType() {}

func (am AppModule) Name() string {
	return am.AppModuleBasic.Name()
}

func (am AppModule) RegisterServices(cfg module.Configurator) {
	types.RegisterMsgServer(cfg.MsgServer(), keeper.NewMsgServerImpl(*am.keeper))
	types.RegisterQueryServer(cfg.QueryServer(), am.keeper)
}

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, gs json.RawMessage) {
	var genState types.GenesisState
	cdc.MustUnmarshalJSON(gs, &genState)
	keeper.InitGenesis(ctx, am.keeper, genState)
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	genState := keeper.ExportGenesis(ctx, am.keeper)
	return cdc.MustMarshalJSON(genState)
}

func (AppModule) ConsensusVersion() uint64 { return 1 }

func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: "dymensionxyz.dymension.agent.Query",
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod:      "Agent",
					Use:            "show-agent [agent-id]",
					Short:          "Show a registered agent by id",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "agent_id"}},
				},
				{
					RpcMethod: "Agents",
					Use:       "list-agents",
					Short:     "List all registered agents",
				},
				{
					RpcMethod:      "AgentActions",
					Use:            "agent-actions [agent-id]",
					Short:          "List an agent's attested action log",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "agent_id"}},
				},
				{
					RpcMethod:      "AgentAction",
					Use:            "agent-action [agent-id] [seq]",
					Short:          "Show a single attested action log entry",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "agent_id"}, {ProtoField: "seq"}},
				},
				{
					RpcMethod:      "AgentReputation",
					Use:            "agent-reputation [agent-id]",
					Short:          "Show an agent's aggregate reputation and average score",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "agent_id"}},
				},
				{
					RpcMethod:      "AgentFeedback",
					Use:            "agent-feedback [agent-id] [client]",
					Short:          "Show a single feedback record by agent and client",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "agent_id"}, {ProtoField: "client"}},
				},
				{
					RpcMethod:      "AgentFeedbacks",
					Use:            "agent-feedbacks [agent-id]",
					Short:          "List all feedback records for an agent",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "agent_id"}},
				},
			},
		},
	}
}

func (am AppModule) GetTxCmd() *cobra.Command {
	return cli.GetTxCmd()
}
