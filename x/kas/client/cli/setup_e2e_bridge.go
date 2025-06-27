package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bcp-innovations/hyperlane-cosmos/util"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	ismtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"
	pdtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/types"
	coretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
)

func CmdSetupBridge() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup-bridge",
		Short: "Sets up the Hyperlane core and warp modules for a bridge test.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			registry := codectypes.NewInterfaceRegistry()
			ismtypes.RegisterInterfaces(registry)

			sCtx := setupCtx{
				clientCtx: clientCtx,
				cmd:       cmd,
				registry:  registry,
			}

			validatorsStr, _ := cmd.Flags().GetString("validators")
			threshold, _ := cmd.Flags().GetUint32("threshold")
			gasDenom, _ := cmd.Flags().GetString("gas-denom")
			remoteRouterAddr, _ := cmd.Flags().GetString("remote-router-address")
			remoteRouterGas, _ := cmd.Flags().GetUint64("remote-router-gas")

			validators := strings.Split(validatorsStr, ",")
			if len(validators) == 0 || validators[0] == "" {
				return fmt.Errorf("validators flag cannot be empty")
			}

			hubDomain := uint32(1260813472)
			counterpartyDomain := uint32(80808082)

			ismId, err := createIsm(sCtx, validators, threshold)
			if err != nil {
				return err
			}

			mailboxId, err := createMailbox(sCtx, ismId, hubDomain)
			if err != nil {
				return err
			}

			_, err = createMerkleHook(sCtx, mailboxId)
			if err != nil {
				return err
			}

			if err := createIgp(sCtx, gasDenom); err != nil {
				return err
			}

			tokenId, err := createSyntheticToken(sCtx, mailboxId)
			if err != nil {
				return err
			}

			if err := enrollRemoteRouter(sCtx, tokenId, counterpartyDomain, remoteRouterAddr, remoteRouterGas); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().String("validators", "", "Comma-separated list of validator hex addresses")
	cmd.Flags().Uint32("threshold", 0, "Multisig threshold for the ISM")
	cmd.Flags().String("gas-denom", "stake", "The denomination to be used for interchain gas payments")
	cmd.Flags().String("remote-router-address", "", "The hex address of the remote router contract on the counterparty chain")
	cmd.Flags().Uint64("remote-router-gas", 200000, "The gas limit to use for transfers to the remote router")

	_ = cmd.MarkFlagRequired("validators")
	_ = cmd.MarkFlagRequired("threshold")
	_ = cmd.MarkFlagRequired("remote-router-address")

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

type setupCtx struct {
	clientCtx client.Context
	cmd       *cobra.Command
	registry  codectypes.InterfaceRegistry
}

func broadcastAndWait(ctx setupCtx, msg sdk.Msg) error {
	if err := tx.GenerateOrBroadcastTxCLI(ctx.clientCtx, ctx.cmd.Flags(), msg); err != nil {
		return err
	}
	time.Sleep(6 * time.Second)
	return nil
}

func createIsm(ctx setupCtx, validators []string, threshold uint32) (util.HexAddress, error) {
	queryClient := ismtypes.NewQueryClient(ctx.clientCtx)

	msg := &ismtypes.MsgCreateMessageIdMultisigIsmRaw{
		Creator:    ctx.clientCtx.GetFromAddress().String(),
		Validators: validators,
		Threshold:  threshold,
	}

	if err := broadcastAndWait(ctx, msg); err != nil {
		return util.HexAddress{}, fmt.Errorf("broadcast create ism tx: %w", err)
	}

	respAfter, err := queryClient.Isms(context.Background(), &ismtypes.QueryIsmsRequest{Pagination: &query.PageRequest{Limit: 1000}})
	if err != nil {
		return util.HexAddress{}, fmt.Errorf("query isms after creation: %w", err)
	}

	for _, anyIsm := range respAfter.Isms {
		var ism ismtypes.HyperlaneInterchainSecurityModule
		if err := ctx.registry.UnpackAny(anyIsm, &ism); err != nil {
			return util.HexAddress{}, fmt.Errorf("unpack ism any: %w", err)
		}
		id, _ := ism.GetId()
		return id, nil
	}

	return util.HexAddress{}, fmt.Errorf("find newly created ISM")
}

func createMailbox(ctx setupCtx, ismId util.HexAddress, hubDomain uint32) (util.HexAddress, error) {
	queryClient := coretypes.NewQueryClient(ctx.clientCtx)

	msg := &coretypes.MsgCreateMailbox{
		Owner:       ctx.clientCtx.GetFromAddress().String(),
		DefaultIsm:  ismId,
		LocalDomain: hubDomain,
	}

	if err := broadcastAndWait(ctx, msg); err != nil {
		return util.HexAddress{}, fmt.Errorf("broadcast create mailbox tx: %w", err)
	}

	respAfter, err := queryClient.Mailboxes(context.Background(), &coretypes.QueryMailboxesRequest{Pagination: &query.PageRequest{Limit: 1000}})
	if err != nil {
		return util.HexAddress{}, fmt.Errorf("query mailboxes after creation: %w", err)
	}

	for _, mbox := range respAfter.Mailboxes {
		return mbox.Id, nil
	}

	return util.HexAddress{}, fmt.Errorf("find newly created mailbox")
}

func createMerkleHook(ctx setupCtx, mailboxId util.HexAddress) (string, error) {
	queryClient := pdtypes.NewQueryClient(ctx.clientCtx)

	msg := &pdtypes.MsgCreateMerkleTreeHook{
		Owner:     ctx.clientCtx.GetFromAddress().String(),
		MailboxId: mailboxId,
	}

	if err := broadcastAndWait(ctx, msg); err != nil {
		return "", fmt.Errorf("broadcast create hook tx: %w", err)
	}

	hooksAfter, err := queryClient.MerkleTreeHooks(context.Background(), &pdtypes.QueryMerkleTreeHooksRequest{Pagination: &query.PageRequest{Limit: 1000}})
	if err != nil {
		return "", fmt.Errorf("query hooks after creation: %w", err)
	}

	for _, hook := range hooksAfter.MerkleTreeHooks {
		return hook.Id, nil
	}

	return "", fmt.Errorf("find newly created merkle hook")
}

// TODO: needed? need to send from relayer?
func createIgp(ctx setupCtx, gasDenom string) error {
	msg := &pdtypes.MsgCreateIgp{
		Owner: ctx.clientCtx.GetFromAddress().String(),
		Denom: gasDenom,
	}

	if err := broadcastAndWait(ctx, msg); err != nil {
		return fmt.Errorf("broadcast create igp tx: %w", err)
	}

	return nil
}

func createSyntheticToken(ctx setupCtx, mailboxId util.HexAddress) (util.HexAddress, error) {
	queryClient := warptypes.NewQueryClient(ctx.clientCtx)

	msg := &warptypes.MsgCreateSyntheticToken{
		Owner:         ctx.clientCtx.GetFromAddress().String(),
		OriginMailbox: mailboxId,
	}

	if err := broadcastAndWait(ctx, msg); err != nil {
		return util.HexAddress{}, fmt.Errorf("broadcast create token tx: %w", err)
	}

	tokensAfter, err := queryClient.Tokens(context.Background(), &warptypes.QueryTokensRequest{Pagination: &query.PageRequest{Limit: 1000}})
	if err != nil {
		return util.HexAddress{}, fmt.Errorf("query tokens after creation: %w", err)
	}

	for _, token := range tokensAfter.Tokens {
		tokenId, err := util.DecodeHexAddress(token.Id)
		if err != nil {
			return util.HexAddress{}, fmt.Errorf("decode created Token ID '%s': %w", token.Id, err)
		}
		return tokenId, nil
	}

	return util.HexAddress{}, fmt.Errorf("find newly created token")
}

func enrollRemoteRouter(ctx setupCtx, tokenId util.HexAddress, domain uint32, addr string, gas uint64) error {
	remoteRouterContract, err := util.DecodeHexAddress(addr)
	if err != nil {
		return fmt.Errorf("invalid remote router address '%s': %w", addr, err)
	}

	msg := &warptypes.MsgEnrollRemoteRouter{
		Owner:   ctx.clientCtx.GetFromAddress().String(),
		TokenId: tokenId,
		RemoteRouter: &warptypes.RemoteRouter{
			ReceiverDomain:   domain,
			ReceiverContract: remoteRouterContract.String(),
			Gas:              math.NewIntFromUint64(gas),
		},
	}

	if err := broadcastAndWait(ctx, msg); err != nil {
		return fmt.Errorf("enroll remote router: %w", err)
	}

	return nil
}
