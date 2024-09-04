package apptesting

import (
	"strings"
	"time"

	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"

	"github.com/dymensionxyz/dymension/v3/app/params"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"

	"github.com/cometbft/cometbft/libs/rand"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/dymensionxyz/sdk-utils/utils/urand"
	"github.com/stretchr/testify/suite"

	"github.com/dymensionxyz/dymension/v3/app"
	rollappkeeper "github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	sequencerkeeper "github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	sequencertypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

var (
	alice = "dym1wg8p6j0pxpnsvhkwfu54ql62cnrumf0v634mft"
	bond  = sequencertypes.DefaultParams().MinBond
)

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

// creates a rollapp and return its rollappID
func (s *KeeperTestHelper) CreateDefaultRollapp() string {
	rollappId := urand.RollappID()
	s.CreateRollappByName(rollappId)
	return rollappId
}

func (s *KeeperTestHelper) CreateRollappByName(name string) {
	msgCreateRollapp := rollapptypes.MsgCreateRollapp{
		Creator:          alice,
		RollappId:        name,
		InitialSequencer: "*",
		Alias:            strings.ToLower(rand.Str(7)),
		VmType:           rollapptypes.Rollapp_EVM,
		GenesisInfo: rollapptypes.GenesisInfo{
			Bech32Prefix:    strings.ToLower(rand.Str(3)),
			GenesisChecksum: "1234567890abcdefg",
			InitialSupply:   sdk.NewInt(1000),
			NativeDenom: &rollapptypes.DenomMetadata{
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
			TokenSymbol: "ROL",
		},
	}

	s.FundForAliasRegistration(msgCreateRollapp)

	msgServer := rollappkeeper.NewMsgServerImpl(*s.App.RollappKeeper)
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
	err := bankutil.FundAccount(s.App.BankKeeper, ctx, addr, sdk.NewCoins(bond))
	s.Require().Nil(err)

	pkAny, err := codectypes.NewAnyWithValue(pubKey)
	s.Require().Nil(err)

	sequencerMsg1 := sequencertypes.MsgCreateSequencer{
		Creator:      addr.String(),
		DymintPubKey: pkAny,
		Bond:         bond,
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
	var bds rollapptypes.BlockDescriptors
	bds.BD = make([]rollapptypes.BlockDescriptor, numOfBlocks)
	for k := 0; k < int(numOfBlocks); k++ {
		bds.BD[k] = rollapptypes.BlockDescriptor{Height: startHeight + uint64(k), Timestamp: time.Now().UTC()}
	}

	updateState := rollapptypes.MsgUpdateState{
		Creator:     seqAddr,
		RollappId:   rollappId,
		StartHeight: startHeight,
		NumBlocks:   numOfBlocks,
		DAPath:      "",
		BDs:         bds,
	}
	msgServer := rollappkeeper.NewMsgServerImpl(*s.App.RollappKeeper)
	_, err = msgServer.UpdateState(ctx, &updateState)
	return startHeight + numOfBlocks, err
}

// FundAcc funds target address with specified amount.
func (s *KeeperTestHelper) FundAcc(acc sdk.AccAddress, amounts sdk.Coins) {
	err := bankutil.FundAccount(s.App.BankKeeper, s.Ctx, acc, amounts)
	s.Require().NoError(err)
}

// FundModuleAcc funds target modules with specified amount.
func (s *KeeperTestHelper) FundModuleAcc(moduleName string, amounts sdk.Coins) {
	err := bankutil.FundModuleAccount(s.App.BankKeeper, s.Ctx, moduleName, amounts)
	s.Require().NoError(err)
}

// StateNotAltered validates that app state is not altered. Fails if it is.
func (s *KeeperTestHelper) StateNotAltered() {
	oldState := s.App.ExportState(s.Ctx)
	s.App.Commit()
	newState := s.App.ExportState(s.Ctx)
	s.Require().Equal(oldState, newState)
}

func (s *KeeperTestHelper) FundForAliasRegistration(msgCreateRollApp rollapptypes.MsgCreateRollapp) {
	err := FundForAliasRegistration(s.Ctx, s.App.BankKeeper, msgCreateRollApp)
	s.Require().NoError(err)
}

func FundForAliasRegistration(
	ctx sdk.Context,
	bankKeeper bankkeeper.Keeper,
	msgCreateRollApp rollapptypes.MsgCreateRollapp,
) error {
	if msgCreateRollApp.Alias == "" {
		return nil
	}
	dymNsParams := dymnstypes.DefaultPriceParams()
	aliasRegistrationCost := sdk.NewCoins(sdk.NewCoin(
		params.BaseDenom, dymNsParams.GetAliasPrice(msgCreateRollApp.Alias),
	))
	return bankutil.FundAccount(
		bankKeeper, ctx, sdk.MustAccAddressFromBech32(msgCreateRollApp.Creator), aliasRegistrationCost,
	)
}
