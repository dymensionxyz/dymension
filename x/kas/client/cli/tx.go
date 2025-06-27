package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	"github.com/dymensionxyz/dymension/v3/x/kas/types"
)

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
	fmt.Println("1. Creating Message ID Multisig ISM (Raw)...")
	queryClient := ismtypes.NewQueryClient(ctx.clientCtx)

	respBefore, err := queryClient.Isms(context.Background(), &ismtypes.QueryIsmsRequest{Pagination: &query.PageRequest{Limit: 1000}})
	if err != nil {
		return util.HexAddress{}, fmt.Errorf("failed to query isms before creation: %w", err)
	}

	existingIds := make(map[string]struct{})
	for _, anyIsm := range respBefore.Isms {
		var ism ismtypes.HyperlaneInterchainSecurityModule
		if err := ctx.registry.UnpackAny(anyIsm, &ism); err != nil {
			return util.HexAddress{}, fmt.Errorf("failed to unpack ism any: %w", err)
		}
		id, _ := ism.GetId()
		existingIds[id.String()] = struct{}{}
	}

	msg := &ismtypes.MsgCreateMessageIdMultisigIsmRaw{
		Creator:    ctx.clientCtx.GetFromAddress().String(),
		Validators: validators,
		Threshold:  threshold,
	}

	if err := broadcastAndWait(ctx, msg); err != nil {
		return util.HexAddress{}, fmt.Errorf("failed to broadcast create ism tx: %w", err)
	}

	respAfter, err := queryClient.Isms(context.Background(), &ismtypes.QueryIsmsRequest{Pagination: &query.PageRequest{Limit: 1000}})
	if err != nil {
		return util.HexAddress{}, fmt.Errorf("failed to query isms after creation: %w", err)
	}

	for _, anyIsm := range respAfter.Isms {
		var ism ismtypes.HyperlaneInterchainSecurityModule
		if err := ctx.registry.UnpackAny(anyIsm, &ism); err != nil {
			return util.HexAddress{}, fmt.Errorf("failed to unpack ism any: %w", err)
		}
		id, _ := ism.GetId()
		if _, found := existingIds[id.String()]; !found {
			fmt.Printf("ISM created successfully. ID: %s\n", id.String())
			return id, nil
		}
	}

	return util.HexAddress{}, fmt.Errorf("could not find newly created ISM")
}

func createMailbox(ctx setupCtx, ismId util.HexAddress, hubDomain uint32) (util.HexAddress, error) {
	fmt.Println("2. Creating Mailbox...")
	queryClient := coretypes.NewQueryClient(ctx.clientCtx)

	respBefore, err := queryClient.Mailboxes(context.Background(), &coretypes.QueryMailboxesRequest{Pagination: &query.PageRequest{Limit: 1000}})
	if err != nil {
		return util.HexAddress{}, fmt.Errorf("failed to query mailboxes before creation: %w", err)
	}

	existingIds := make(map[string]struct{})
	for _, mbox := range respBefore.Mailboxes {
		existingIds[mbox.Id.String()] = struct{}{}
	}

	msg := &coretypes.MsgCreateMailbox{
		Owner:       ctx.clientCtx.GetFromAddress().String(),
		DefaultIsm:  ismId,
		LocalDomain: hubDomain,
	}

	if err := broadcastAndWait(ctx, msg); err != nil {
		return util.HexAddress{}, fmt.Errorf("failed to broadcast create mailbox tx: %w", err)
	}

	respAfter, err := queryClient.Mailboxes(context.Background(), &coretypes.QueryMailboxesRequest{Pagination: &query.PageRequest{Limit: 1000}})
	if err != nil {
		return util.HexAddress{}, fmt.Errorf("failed to query mailboxes after creation: %w", err)
	}

	for _, mbox := range respAfter.Mailboxes {
		if _, found := existingIds[mbox.Id.String()]; !found {
			fmt.Printf("Mailbox created successfully. ID: %s\n", mbox.Id.String())
			return mbox.Id, nil
		}
	}

	return util.HexAddress{}, fmt.Errorf("could not find newly created mailbox")
}

func createMerkleHook(ctx setupCtx, mailboxId util.HexAddress) (string, error) {
	fmt.Println("3. Creating Merkle Tree Hook...")
	queryClient := pdtypes.NewQueryClient(ctx.clientCtx)

	hooksBefore, err := queryClient.MerkleTreeHooks(context.Background(), &pdtypes.QueryMerkleTreeHooksRequest{Pagination: &query.PageRequest{Limit: 1000}})
	if err != nil {
		return "", fmt.Errorf("failed to query hooks before creation: %w", err)
	}

	existingHookIds := make(map[string]struct{})
	for _, hook := range hooksBefore.MerkleTreeHooks {
		existingHookIds[hook.Id] = struct{}{}
	}

	msg := &pdtypes.MsgCreateMerkleTreeHook{
		Owner:     ctx.clientCtx.GetFromAddress().String(),
		MailboxId: mailboxId,
	}

	if err := broadcastAndWait(ctx, msg); err != nil {
		return "", fmt.Errorf("failed to broadcast create hook tx: %w", err)
	}

	hooksAfter, err := queryClient.MerkleTreeHooks(context.Background(), &pdtypes.QueryMerkleTreeHooksRequest{Pagination: &query.PageRequest{Limit: 1000}})
	if err != nil {
		return "", fmt.Errorf("failed to query hooks after creation: %w", err)
	}

	for _, hook := range hooksAfter.MerkleTreeHooks {
		if _, found := existingHookIds[hook.Id]; !found {
			fmt.Printf("Merkle Tree Hook created successfully. ID: %s\n", hook.Id)
			return hook.Id, nil
		}
	}

	return "", fmt.Errorf("could not find newly created merkle hook")
}

func createIgp(ctx setupCtx, gasDenom string) error {
	fmt.Println("4. Creating Interchain Gas Paymaster (IGP)...")
	msg := &pdtypes.MsgCreateIgp{
		Owner: ctx.clientCtx.GetFromAddress().String(),
		Denom: gasDenom,
	}

	if err := broadcastAndWait(ctx, msg); err != nil {
		return fmt.Errorf("failed to broadcast create igp tx: %w", err)
	}

	fmt.Println("IGP created successfully.")
	return nil
}

func createSyntheticToken(ctx setupCtx, mailboxId util.HexAddress) (util.HexAddress, error) {
	fmt.Println("5. Creating Synthetic Token...")
	queryClient := warptypes.NewQueryClient(ctx.clientCtx)

	tokensBefore, err := queryClient.Tokens(context.Background(), &warptypes.QueryTokensRequest{Pagination: &query.PageRequest{Limit: 1000}})
	if err != nil {
		return util.HexAddress{}, fmt.Errorf("failed to query tokens before creation: %w", err)
	}

	existingTokenIds := make(map[string]struct{})
	for _, token := range tokensBefore.Tokens {
		existingTokenIds[token.Id] = struct{}{}
	}

	msg := &warptypes.MsgCreateSyntheticToken{
		Owner:         ctx.clientCtx.GetFromAddress().String(),
		OriginMailbox: mailboxId,
	}

	if err := broadcastAndWait(ctx, msg); err != nil {
		return util.HexAddress{}, fmt.Errorf("failed to broadcast create token tx: %w", err)
	}

	tokensAfter, err := queryClient.Tokens(context.Background(), &warptypes.QueryTokensRequest{Pagination: &query.PageRequest{Limit: 1000}})
	if err != nil {
		return util.HexAddress{}, fmt.Errorf("failed to query tokens after creation: %w", err)
	}

	for _, token := range tokensAfter.Tokens {
		if _, found := existingTokenIds[token.Id]; !found {
			tokenId, err := util.DecodeHexAddress(token.Id)
			if err != nil {
				return util.HexAddress{}, fmt.Errorf("failed to decode created Token ID '%s': %w", token.Id, err)
			}
			fmt.Printf("Synthetic Token created successfully. ID: %s\n", tokenId.String())
			return tokenId, nil
		}
	}

	return util.HexAddress{}, fmt.Errorf("could not find newly created token")
}

func enrollRemoteRouter(ctx setupCtx, tokenId util.HexAddress, domain uint32, addr string, gas uint64) error {
	fmt.Println("6. Enrolling Remote Router...")

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
			Gas:              sdk.NewIntFromUint64(gas),
		},
	}

	if err := broadcastAndWait(ctx, msg); err != nil {
		return fmt.Errorf("failed to enroll remote router: %w", err)
	}

	fmt.Println("Remote Router enrolled successfully.")
	return nil
}

func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdSetupBridge())
	return cmd
}

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

			fmt.Println("Hyperlane bridge setup complete!")
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
