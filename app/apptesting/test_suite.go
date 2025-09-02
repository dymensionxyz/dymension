package apptesting

import (
	"strings"
	"time"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/rand"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/dymensionxyz/sdk-utils/utils/urand"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app"
	"github.com/dymensionxyz/dymension/v3/app/params"
	delayedackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencerkeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

var Alice = "dym1wg8p6j0pxpnsvhkwfu54ql62cnrumf0v634mft"

func init() {
	config := sdk.GetConfig()
	params.SetAddressPrefixes(config)
	config.Seal()
}

type KeeperTestHelper struct {
	suite.Suite
	App *app.App
	Ctx sdk.Context
}

func (s *KeeperTestHelper) CreateDefaultRollappAndProposer() (string, string) {
	rollappId := s.CreateDefaultRollapp()
	proposer := s.CreateDefaultSequencer(s.Ctx, rollappId)
	return rollappId, proposer
}

func (s *KeeperTestHelper) CreateDefaultRollapp() string {
	rollappId := urand.RollappID()
	s.CreateRollappByName(rollappId)
	return rollappId
}

func (s *KeeperTestHelper) CreateFairLaunchRollapp() string {
	iroParams := s.App.IROKeeper.GetParams(s.Ctx)

	rollappId := urand.RollappID()
	s.CreateRollappByName(rollappId)

	rollapp := s.App.RollappKeeper.MustGetRollapp(s.Ctx, rollappId)
	rollapp.GenesisInfo.InitialSupply = iroParams.FairLaunch.AllocationAmount
	rollapp.GenesisInfo.GenesisAccounts = &rollapptypes.GenesisAccounts{
		Accounts: []rollapptypes.GenesisAccount{
			{
				Address: s.App.IROKeeper.GetModuleAccountAddress(),
				Amount:  iroParams.FairLaunch.AllocationAmount,
			},
		},
	}
	s.App.RollappKeeper.SetRollapp(s.Ctx, rollapp)

	return rollappId
}

func (s *KeeperTestHelper) CreateRollappByName(name string) {
	msgCreateRollapp := rollapptypes.MsgCreateRollapp{
		Creator:          Alice,
		RollappId:        name,
		InitialSequencer: "*",
		MinSequencerBond: rollapptypes.DefaultMinSequencerBondGlobalCoin,

		Alias:  strings.ToLower(rand.Str(7)),
		VmType: rollapptypes.Rollapp_EVM,
		GenesisInfo: &rollapptypes.GenesisInfo{
			Bech32Prefix:    strings.ToLower(rand.Str(3)),
			GenesisChecksum: "1234567890abcdefg",
			InitialSupply:   math.NewInt(1000),
			NativeDenom: rollapptypes.DenomMetadata{
				Display:  "DEN",
				Base:     "aden",
				Exponent: 18,
			},
		},
		Metadata: &rollapptypes.RollappMetadata{
			Website:     "https://dymension.xyz",
			Description: "Sample description",
			LogoUrl:     "https://dymension.xyz/logo.png",
			Telegram:    "https://t.me/rolly",
			X:           "https://x.dymension.xyz",
		},
	}

	s.FundForAliasRegistration(msgCreateRollapp)

	msgServer := rollappkeeper.NewMsgServerImpl(s.App.RollappKeeper)
	_, err := msgServer.CreateRollapp(s.Ctx, &msgCreateRollapp)
	s.Require().NoError(err)
}

func (s *KeeperTestHelper) CreateDefaultSequencer(ctx sdk.Context, rollappId string) string {
	pubkey := ed25519.GenPrivKey().PubKey()
	err := s.CreateSequencerByPubkey(ctx, rollappId, pubkey)
	s.Require().NoError(err)
	return sdk.AccAddress(pubkey.Address()).String()
}

func (s *KeeperTestHelper) CreateSequencerByPubkey(ctx sdk.Context, rollappId string, pubKey types.PubKey) error {
	addr := sdk.AccAddress(pubKey.Address())
	// fund account
	FundAccount(s.App, ctx, addr, sdk.NewCoins(rollapptypes.DefaultMinSequencerBondGlobalCoin))

	pkAny, err := codectypes.NewAnyWithValue(pubKey)
	s.Require().Nil(err)

	sequencerMsg1 := sequencertypes.MsgCreateSequencer{
		Creator:      addr.String(),
		DymintPubKey: pkAny,
		Bond:         rollapptypes.DefaultMinSequencerBondGlobalCoin,
		RollappId:    rollappId,
		Metadata: sequencertypes.SequencerMetadata{
			Rpcs:    []string{"https://rpc.wpd.evm.rollapp.noisnemyd.xyz:443"},
			EvmRpcs: []string{"https://rpc.evm.rollapp.noisnemyd.xyz:443"},
		},
	}

	msgServer := sequencerkeeper.NewMsgServerImpl(s.App.SequencerKeeper)
	_, err = msgServer.CreateSequencer(ctx, &sequencerMsg1)
	return err
}

func (s *KeeperTestHelper) PostStateUpdate(ctx sdk.Context, rollappId, seqAddr string, startHeight, numOfBlocks uint64) (lastHeight uint64, err error) {
	return s.PostStateUpdateWithOptions(ctx, rollappId, seqAddr, startHeight, numOfBlocks, 0, 1)
}

func (s *KeeperTestHelper) PostStateUpdateWithDRSVersion(ctx sdk.Context, rollappId, seqAddr string, startHeight, numOfBlocks uint64, drsVersion uint32) (lastHeight uint64, err error) {
	return s.PostStateUpdateWithOptions(ctx, rollappId, seqAddr, startHeight, numOfBlocks, 0, drsVersion)
}

func (s *KeeperTestHelper) PostStateUpdateWithRevision(ctx sdk.Context, rollappId, seqAddr string, startHeight, numOfBlocks, revision uint64) (lastHeight uint64, err error) {
	return s.PostStateUpdateWithOptions(ctx, rollappId, seqAddr, startHeight, numOfBlocks, revision, 1)
}

func (s *KeeperTestHelper) PostStateUpdateWithOptions(ctx sdk.Context, rollappId, seqAddr string, startHeight, numOfBlocks, revision uint64, drsVersion uint32) (lastHeight uint64, err error) {
	var bds rollapptypes.BlockDescriptors
	bds.BD = make([]rollapptypes.BlockDescriptor, numOfBlocks)
	for k := uint64(0); k < numOfBlocks; k++ {
		bds.BD[k] = rollapptypes.BlockDescriptor{Height: startHeight + k, Timestamp: time.Now().UTC(), DrsVersion: drsVersion}
	}

	updateState := rollapptypes.MsgUpdateState{
		Creator:         seqAddr,
		RollappId:       rollappId,
		StartHeight:     startHeight,
		NumBlocks:       numOfBlocks,
		DAPath:          "",
		BDs:             bds,
		RollappRevision: revision,
		Last:            false,
	}
	msgServer := rollappkeeper.NewMsgServerImpl(s.App.RollappKeeper)
	_, err = msgServer.UpdateState(ctx, &updateState)
	return startHeight + numOfBlocks, err
}

// FundAcc funds target address with specified amount.
func (s *KeeperTestHelper) FundAcc(acc sdk.AccAddress, amounts sdk.Coins) {
	err := bankutil.FundAccount(s.Ctx, s.App.BankKeeper, acc, amounts)
	s.Require().NoError(err)
}

// FundModuleAcc funds target modules with specified amount.
func (s *KeeperTestHelper) FundModuleAcc(moduleName string, amounts sdk.Coins) {
	err := bankutil.FundModuleAccount(s.Ctx, s.App.BankKeeper, moduleName, amounts)
	s.Require().NoError(err)
}

func (s *KeeperTestHelper) FundForAliasRegistration(msgCreateRollApp rollapptypes.MsgCreateRollapp) {
	FundForAliasRegistration(s.App, s.Ctx, msgCreateRollApp.Alias, msgCreateRollApp.Creator)
}

func (s *KeeperTestHelper) FinalizeAllPendingPackets(address string) int {
	s.T().Helper()
	// Query all pending packets by address
	querier := delayedackkeeper.NewQuerier(s.App.DelayedAckKeeper)
	packets, err := querier.Keeper.GetPendingPacketsByAddress(s.Ctx, address)
	s.Require().NoError(err)
	// Finalize all packets and return the num of finalized
	for _, packet := range packets {
		handler := s.App.MsgServiceRouter().Handler(new(delayedacktypes.MsgFinalizePacket))
		resp, err := handler(s.Ctx, &delayedacktypes.MsgFinalizePacket{
			Sender:            authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			RollappId:         packet.RollappId,
			PacketProofHeight: packet.ProofHeight,
			PacketType:        packet.Type,
			PacketSrcChannel:  packet.Packet.SourceChannel,
			PacketSequence:    packet.Packet.Sequence,
		})
		s.Require().NoError(err)
		s.Require().NotNil(resp)
	}
	return len(packets)
}

// StateNotAltered validates that app state is not altered. Fails if it is.
func (s *KeeperTestHelper) StateNotAltered() {
	oldState := s.App.ExportState(s.Ctx)
	_, err := s.App.Commit()
	s.Require().NoError(err)
	newState := s.App.ExportState(s.Ctx)
	s.Require().Equal(oldState, newState)
}
