package uhyp

import (
	"cosmossdk.io/math"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	ismkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/keeper"
	ismtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"
	pdkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/keeper"
	pdtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/types"
	corekeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/keeper"
	types "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	warpkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/warp/keeper"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

func (s *Server) CreateDefaultMailbox(ctx sdk.Context, creator string) (hyperutil.HexAddress, error) {
	igp, err := s.CreateIGP(ctx, creator)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	noopHook, err := s.CreateNoopHook(ctx, creator)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	ism, err := s.CreateNoopIsm(ctx, creator)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	localDomain := uint32(0)

	mailboxId, err := s.CreateMailbox(ctx, creator, localDomain, ism, igp, noopHook)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}

	return mailboxId, nil
}

/////////////////////////
// Core
/////////////////////////

func (s *Server) CreateMailbox(ctx sdk.Context, creator string, localDomain uint32, ismId, igpHookId, noopHookId hyperutil.HexAddress) (hyperutil.HexAddress, error) {

	msg := &types.MsgCreateMailbox{
		Owner:        creator,
		LocalDomain:  localDomain,
		DefaultIsm:   ismId,
		DefaultHook:  &noopHookId,
		RequiredHook: &igpHookId,
	}
	res, err := s.coreServer().CreateMailbox(ctx, msg)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	ret, err := hyperutil.DecodeHexAddress(res.Id)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}

	return ret, nil
}

func (s *Server) SetMailbox(ctx sdk.Context,
	creator string,
	mailboxId hyperutil.HexAddress,
	localDomain uint32,
	ismId, igpHookId, noopHookId hyperutil.HexAddress,
	newOwner string,
) error {
	msg := &types.MsgSetMailbox{
		Owner:        creator,
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

func (s *Server) CreateIGP(ctx sdk.Context, creator string) (hyperutil.HexAddress, error) {
	msg := &pdtypes.MsgCreateIgp{
		Owner: creator,
		Denom: "acoin",
	}
	res, err := s.pdServer().CreateIgp(ctx, msg)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	ret, err := hyperutil.DecodeHexAddress(res.Id)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	return ret, nil
}

func (s *Server) CreateNoopHook(ctx sdk.Context, creator string) (hyperutil.HexAddress, error) {
	msg := &pdtypes.MsgCreateNoopHook{
		Owner: creator,
	}
	res, err := s.pdServer().CreateNoopHook(ctx, msg)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	ret, err := hyperutil.DecodeHexAddress(res.Id)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	return ret, nil
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
	mailboxId string, // hex address?
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

func (s *Server) CreateSyntheticToken(ctx sdk.Context, creator string, originMailbox hyperutil.HexAddress) (hyperutil.HexAddress, error) {
	msg := &warptypes.MsgCreateSyntheticToken{
		Owner:         creator,
		OriginMailbox: originMailbox,
	}
	res, err := s.warpServer().CreateSyntheticToken(ctx, msg)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	ret, err := hyperutil.DecodeHexAddress(res.Id)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	return ret, nil
}

func (s *Server) CreateCollateralToken(ctx sdk.Context, creator string, originMailbox hyperutil.HexAddress, originDenom string) (hyperutil.HexAddress, error) {
	msg := &warptypes.MsgCreateCollateralToken{
		Owner:         creator,
		OriginMailbox: originMailbox,
		OriginDenom:   originDenom,
	}
	res, err := s.warpServer().CreateCollateralToken(ctx, msg)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	ret, err := hyperutil.DecodeHexAddress(res.Id)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	return ret, nil
}

func (s *Server) SetToken(ctx sdk.Context, creator string, tokenId hyperutil.HexAddress, newOwner string, ismId *hyperutil.HexAddress) error {
	msg := &warptypes.MsgSetToken{
		Owner:    creator,
		TokenId:  tokenId,
		NewOwner: newOwner,
		IsmId:    ismId,
	}
	_, err := s.warpServer().SetToken(ctx, msg)
	return err
}

func (s *Server) EnrollRemoteRouter(ctx sdk.Context, creator string, tokenId hyperutil.HexAddress, remoteRouter *warptypes.RemoteRouter) error {
	msg := &warptypes.MsgEnrollRemoteRouter{
		Owner:        creator,
		TokenId:      tokenId,
		RemoteRouter: remoteRouter,
	}
	_, err := s.warpServer().EnrollRemoteRouter(ctx, msg)
	return err
}

func (s *Server) UnrollRemoteRouter(ctx sdk.Context, creator string, tokenId hyperutil.HexAddress, receiverDomain uint32) error {
	msg := &warptypes.MsgUnrollRemoteRouter{
		Owner:          creator,
		TokenId:        tokenId,
		ReceiverDomain: receiverDomain,
	}
	_, err := s.warpServer().UnrollRemoteRouter(ctx, msg)
	return err
}

func (s *Server) RemoteTransfer(
	ctx sdk.Context,
	creator string,
	tokenId hyperutil.HexAddress,
	destinationDomain uint32,
	recipient hyperutil.HexAddress,
	customHookId hyperutil.HexAddress,
	maxFee sdk.Coin,
	customHookMetadata string,
	amt math.Int,
	gasLimit math.Int,
) (string, error) {
	msg := &warptypes.MsgRemoteTransfer{
		Sender:             creator,
		TokenId:            tokenId,
		DestinationDomain:  destinationDomain,
		Recipient:          recipient,
		Amount:             amt,
		CustomHookId:       &customHookId,
		GasLimit:           gasLimit,
		MaxFee:             maxFee,
		CustomHookMetadata: customHookMetadata,
	}
	res, err := s.warpServer().RemoteTransfer(ctx, msg)
	if err != nil {
		return "", err
	}
	return res.MessageId, nil
}
