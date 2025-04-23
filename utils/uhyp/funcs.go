package uhyp

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	hyperutil "github.com/dymensionxyz/hyperlane-cosmos/util"
	ismkeeper "github.com/dymensionxyz/hyperlane-cosmos/x/core/01_interchain_security/keeper"
	ismtypes "github.com/dymensionxyz/hyperlane-cosmos/x/core/01_interchain_security/types"
	pdkeeper "github.com/dymensionxyz/hyperlane-cosmos/x/core/02_post_dispatch/keeper"
	pdtypes "github.com/dymensionxyz/hyperlane-cosmos/x/core/02_post_dispatch/types"
	corekeeper "github.com/dymensionxyz/hyperlane-cosmos/x/core/keeper"
	types "github.com/dymensionxyz/hyperlane-cosmos/x/core/types"
	warpkeeper "github.com/dymensionxyz/hyperlane-cosmos/x/warp/keeper"
	warptypes "github.com/dymensionxyz/hyperlane-cosmos/x/warp/types"
)

/*
This function exposes some functions publicly (via copy paste) for easier integration testing.
*/

var LocalDomain = uint32(1)
var RemoteDomain = uint32(2)

func MustDecodeHexAddress(s string) hyperutil.HexAddress {
	addr, err := hyperutil.DecodeHexAddress(s)
	if err != nil {
		panic(err)
	}
	return addr
}

type Server struct {
	coreK *corekeeper.Keeper
	pdK   *pdkeeper.Keeper
	ismK  *ismkeeper.Keeper
	warpK warpkeeper.Keeper
}

func NewServer(coreK *corekeeper.Keeper, pdK *pdkeeper.Keeper, ismK *ismkeeper.Keeper, warpK warpkeeper.Keeper) *Server {
	return &Server{
		coreK: coreK,
		pdK:   pdK,
		ismK:  ismK,
		warpK: warpK,
	}
}

func (s *Server) coreServer() types.MsgServer {
	return corekeeper.NewMsgServerImpl(s.coreK)
}

func (s *Server) pdServer() pdtypes.MsgServer {
	return pdkeeper.NewMsgServerImpl(s.pdK)
}

func (s *Server) ismServer() ismtypes.MsgServer {
	return ismkeeper.NewMsgServerImpl(s.ismK)
}

func (s *Server) warpServer() warptypes.MsgServer {
	return warpkeeper.NewMsgServerImpl(s.warpK)
}

// creates a mailbox with IGP but no merkle tree.
// That means the post dispatch will not create a merkle tree.
// Also, any warp routes which don't specify an override ISM will go to the default ISM (via approuter) which is noop.
// So inbound warp route messages are not verified.
func (s *Server) CreateDefaultMailbox(ctx sdk.Context, owner string, denom string) (hyperutil.HexAddress, error) {
	ism, err := s.CreateNoopIsm(ctx, owner)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	// igp, err := s.CreateIGP(ctx, owner, denom)
	// if err != nil {
	// 	return hyperutil.HexAddress{}, err
	// }
	noopHook, err := s.CreateNoopHook(ctx, owner)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	// remoteDomain := uint32(1)
	// err = s.SetDestinationGasConfig(ctx,
	// 	owner,
	// 	igp.String(),
	// 	remoteDomain, // should this really by?
	// 	math.NewInt(1e10),
	// 	math.NewInt(1),
	// 	math.NewInt(200000),
	// )
	// if err != nil {
	// return hyperutil.HexAddress{}, err
	// }

	mailboxId, err := s.CreateMailbox(ctx, CreateMailboxArgs{
		Owner:        owner,
		LocalDomain:  LocalDomain,
		Ism:          ism,
		RequiredHook: noopHook, // would typically be IGP (sure?)
		DefaultHook:  noopHook, // would typically be merkle hook (sure?)
	})
	if err != nil {
		return hyperutil.HexAddress{}, err
	}

	return mailboxId, nil
}

/////////////////////////
// Core
/////////////////////////

type CreateMailboxArgs struct {
	Owner        string
	LocalDomain  uint32
	Ism          hyperutil.HexAddress
	RequiredHook hyperutil.HexAddress // Runs first.
	DefaultHook  hyperutil.HexAddress // Runs second.
}

func (s *Server) CreateMailbox(ctx sdk.Context,
	args CreateMailboxArgs,
) (hyperutil.HexAddress, error) {

	msg := &types.MsgCreateMailbox{
		Owner:        args.Owner,
		LocalDomain:  args.LocalDomain,
		DefaultIsm:   args.Ism,
		DefaultHook:  &args.DefaultHook,
		RequiredHook: &args.RequiredHook,
	}
	res, err := s.coreServer().CreateMailbox(ctx, msg)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	return res.Id, nil

}

func (s *Server) SetMailbox(ctx sdk.Context,
	owner string,
	mailboxId hyperutil.HexAddress,
	localDomain uint32,
	ismId, igpHookId, noopHookId hyperutil.HexAddress,
	newOwner string,
) error {
	msg := &types.MsgSetMailbox{
		Owner:        owner,
		MailboxId:    mailboxId,
		DefaultIsm:   &ismId,
		DefaultHook:  &noopHookId,
		RequiredHook: &igpHookId,
		NewOwner:     newOwner,
	}
	_, err := s.coreServer().SetMailbox(ctx, msg)
	return err
}

func (s *Server) ProcessMessage(ctx sdk.Context,
	mailboxId hyperutil.HexAddress,
	relayer string,
	metadata string,
	message string,
) error {
	msg := &types.MsgProcessMessage{
		MailboxId: mailboxId,
		Relayer:   relayer,
		Metadata:  metadata,
		Message:   message,
	}
	_, err := s.coreServer().ProcessMessage(ctx, msg)
	return err
}

/////////////////////////
// Post Dispatch
/////////////////////////

func (s *Server) CreateIGP(ctx sdk.Context, owner string, denom string) (hyperutil.HexAddress, error) {
	msg := &pdtypes.MsgCreateIgp{
		Owner: owner,
		Denom: denom,
	}
	res, err := s.pdServer().CreateIgp(ctx, msg)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	return res.Id, nil
}

func (s *Server) SetIgpOwner(ctx sdk.Context, owner string, igpId hyperutil.HexAddress, // TODO: hex address?
	newOwner string) error {
	msg := &pdtypes.MsgSetIgpOwner{
		Owner:    owner,
		IgpId:    igpId,
		NewOwner: newOwner,
	}
	_, err := s.pdServer().SetIgpOwner(ctx, msg)
	return err
}

func (s *Server) SetDestinationGasConfig(
	ctx sdk.Context,
	owner string,
	igpId hyperutil.HexAddress,
	remoteDomain uint32,
	tokenExchangeRate math.Int,
	gasPrice math.Int,
	gasOverhead math.Int,
) error {
	cfg := &pdtypes.DestinationGasConfig{
		RemoteDomain: remoteDomain,
		GasOracle: &pdtypes.GasOracle{
			TokenExchangeRate: tokenExchangeRate,
			GasPrice:          gasPrice,
		},
		GasOverhead: gasOverhead,
	}
	msg := &pdtypes.MsgSetDestinationGasConfig{
		Owner:                owner,
		IgpId:                igpId,
		DestinationGasConfig: cfg,
	}
	_, err := s.pdServer().SetDestinationGasConfig(ctx, msg)
	return err
}

func (s *Server) PayforGas(
	ctx sdk.Context,
	sender string,
	igpId hyperutil.HexAddress,
	messageId hyperutil.HexAddress,
	destinationDomain uint32,
	amount sdk.Coin,
	gasLimit math.Int,
) error {
	msg := &pdtypes.MsgPayForGas{
		Sender:            sender,
		IgpId:             igpId,
		MessageId:         messageId,
		DestinationDomain: destinationDomain,
		Amount:            amount,
		GasLimit:          gasLimit,
	}
	_, err := s.pdServer().PayForGas(ctx, msg)
	return err
}

func (s *Server) Claim(
	ctx sdk.Context,
	sender string,
	igpId hyperutil.HexAddress,
) error {
	msg := &pdtypes.MsgClaim{
		Sender: sender,
		IgpId:  igpId,
	}
	_, err := s.pdServer().Claim(ctx, msg)
	return err
}

func (s *Server) CreateMerkleTreeHook(ctx sdk.Context, owner string, mailboxId hyperutil.HexAddress) (hyperutil.HexAddress, error) {
	msg := &pdtypes.MsgCreateMerkleTreeHook{
		Owner:     owner,
		MailboxId: mailboxId,
	}
	res, err := s.pdServer().CreateMerkleTreeHook(ctx, msg)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	return res.Id, nil
}

func (s *Server) CreateNoopHook(ctx sdk.Context, owner string) (hyperutil.HexAddress, error) {
	msg := &pdtypes.MsgCreateNoopHook{
		Owner: owner,
	}
	res, err := s.pdServer().CreateNoopHook(ctx, msg)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	return res.Id, nil
}

/////////////////////////
// Interchain Security Modules
/////////////////////////

func (s *Server) CreateMessageIdMultisigIsm(ctx sdk.Context, creator string) (hyperutil.HexAddress, error) {
	msg := &ismtypes.MsgCreateMessageIdMultisigIsm{
		Creator: creator,
		Validators: []string{
			"0xa05b6a0aa112b61a7aa16c19cac27d970692995e",
			"0xb05b6a0aa112b61a7aa16c19cac27d970692995e",
			"0xd05b6a0aa112b61a7aa16c19cac27d970692995e",
		},
		Threshold: 2,
	}
	res, err := s.ismServer().CreateMessageIdMultisigIsm(ctx, msg)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	return res.Id, nil
}

// (merkle root based)
func (s *Server) CreateMultisigIsm(ctx sdk.Context, creator string) (hyperutil.HexAddress, error) {
	msg := &ismtypes.MsgCreateMerkleRootMultisigIsm{
		Creator: creator,
		Validators: []string{
			"0xa05b6a0aa112b61a7aa16c19cac27d970692995e",
			"0xb05b6a0aa112b61a7aa16c19cac27d970692995e",
			"0xd05b6a0aa112b61a7aa16c19cac27d970692995e",
		},
		Threshold: 2,
	}
	res, err := s.ismServer().CreateMerkleRootMultisigIsm(ctx, msg)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	return res.Id, nil
}

func (s *Server) CreateNoopIsm(ctx sdk.Context, creator string) (hyperutil.HexAddress, error) {
	msg := &ismtypes.MsgCreateNoopIsm{
		Creator: creator,
	}
	res, err := s.ismServer().CreateNoopIsm(ctx, msg)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	return res.Id, nil
}

func (s *Server) AnnounceValidator(ctx sdk.Context,
	validator string,
	storageLocation string,
	signature string,
	mailboxId hyperutil.HexAddress,
	creator string,
) error {
	msg := &ismtypes.MsgAnnounceValidator{
		Validator:       validator,
		StorageLocation: storageLocation,
		Signature:       signature,
		MailboxId:       mailboxId,
		Creator:         creator,
	}
	_, err := s.ismServer().AnnounceValidator(ctx, msg)
	return err
}

/////////////////////////
// Warp Routes
/////////////////////////

// returns tokens id
func (s *Server) CreateSyntheticToken(ctx sdk.Context, owner string, originMailbox hyperutil.HexAddress) (hyperutil.HexAddress, error) {
	msg := &warptypes.MsgCreateSyntheticToken{
		Owner:         owner,
		OriginMailbox: originMailbox,
	}
	res, err := s.warpServer().CreateSyntheticToken(ctx, msg)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	return res.Id, nil
}

// returns tokens id
func (s *Server) CreateCollateralToken(ctx sdk.Context, owner string, originMailbox hyperutil.HexAddress, originDenom string) (hyperutil.HexAddress, error) {
	msg := &warptypes.MsgCreateCollateralToken{
		Owner:         owner,
		OriginMailbox: originMailbox,
		OriginDenom:   originDenom,
	}
	res, err := s.warpServer().CreateCollateralToken(ctx, msg)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	return res.Id, nil
}

func (s *Server) SetToken(ctx sdk.Context, owner string, tokenId hyperutil.HexAddress, newOwner string, ismId *hyperutil.HexAddress) error {
	msg := &warptypes.MsgSetToken{
		Owner:    owner,
		TokenId:  tokenId,
		NewOwner: newOwner,
		IsmId:    ismId,
	}
	_, err := s.warpServer().SetToken(ctx, msg)
	return err
}

func (s *Server) EnrollRemoteRouter(ctx sdk.Context, owner string, tokenId hyperutil.HexAddress, remoteRouter warptypes.RemoteRouter) error {
	msg := &warptypes.MsgEnrollRemoteRouter{
		Owner:        owner,
		TokenId:      tokenId,
		RemoteRouter: &remoteRouter,
	}
	_, err := s.warpServer().EnrollRemoteRouter(ctx, msg)
	return err
}

func (s *Server) UnrollRemoteRouter(ctx sdk.Context, owner string, tokenId hyperutil.HexAddress, receiverDomain uint32) error {
	msg := &warptypes.MsgUnrollRemoteRouter{
		Owner:          owner,
		TokenId:        tokenId,
		ReceiverDomain: receiverDomain,
	}
	_, err := s.warpServer().UnrollRemoteRouter(ctx, msg)
	return err
}

// returns message id
func (s *Server) RemoteTransfer(
	ctx sdk.Context,
	owner string,
	tokenId hyperutil.HexAddress,
	destinationDomain uint32,
	recipient hyperutil.HexAddress,
	customHookId *hyperutil.HexAddress,
	maxFee sdk.Coin,
	customHookMetadata string,
	amt math.Int,
	gasLimit math.Int,
) (hyperutil.HexAddress, error) {
	msg := &warptypes.MsgRemoteTransfer{
		Sender:             owner,
		TokenId:            tokenId,
		DestinationDomain:  destinationDomain,
		Recipient:          recipient,
		Amount:             amt,
		CustomHookId:       customHookId,
		GasLimit:           gasLimit,
		MaxFee:             maxFee,
		CustomHookMetadata: customHookMetadata,
	}
	res, err := s.warpServer().RemoteTransfer(ctx, msg)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	return res.MessageId, nil
}
