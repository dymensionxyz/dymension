package uhyp

import (
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	corekeeper "github.com/bcp-innovations/hyperlane-cosmos/x/core/keeper"
	types "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
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
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) coreServer() types.MsgServer {
	return corekeeper.NewMsgServerImpl(s.coreK)
}

func (s *Server) CreateDefaultMailbox(ctx sdk.Context, creator string) (hyperutil.HexAddress, error) {
	igp := CreateIGP()
	noopHook := CreateNoopHook()
	ism := CreateNoopIsm()
	localDomain := uint32(0)

	mailboxId, err := s.CreateMailbox(ctx, creator, localDomain, ism, igp, noopHook)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}

	return mailboxId, nil

}

func (s *Server) CreateMailbox(ctx sdk.Context, creator string, localDomain uint32, ismId, igpId, noopHookId hyperutil.HexAddress) (hyperutil.HexAddress, error) {

	msg := &types.MsgCreateMailbox{
		Owner:        creator,
		LocalDomain:  localDomain,
		DefaultIsm:   ismId,
		DefaultHook:  &igpId,
		RequiredHook: &igpId,
	}
	res, err := s.coreServer().CreateMailbox(ctx, msg)
	if err != nil {
		return err
	}
	var response types.MsgCreateMailboxResponse
	err = proto.Unmarshal(res.MsgResponses[0].Value, &response)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}
	mailboxId, err := hyperutil.DecodeHexAddress(response.Id)
	if err != nil {
		return hyperutil.HexAddress{}, err
	}

	return mailboxId, nil

}

func (s *Server) CreateIGP() {
}

func (s *Server) CreateNoopHook() {
}

func (s *Server) CreateNoopIsm() {
}

func (s *Server) CreateMultisigIsm() {
}
