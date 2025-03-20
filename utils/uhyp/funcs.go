package uhyp

import (
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	ismkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/keeper"
	ismtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"
	pdkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/keeper"
	pdtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/types"
	corekeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/keeper"
	types "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
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
}

func NewServer() *Server {
	return &Server{}
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
