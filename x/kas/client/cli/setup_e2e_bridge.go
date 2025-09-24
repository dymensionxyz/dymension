package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"cosmossdk.io/math"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	bridgingfeetypes "github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"

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

const (
	flagGasDenom            = "gas-denom"
	flagRemoteRouterAddress = "remote-router-address"
	flagRemoteRouterGas     = "remote-router-gas"
	flagValidators          = "validators"
	flagThreshold           = "threshold"
	flagHubDomain           = "hub-domain"
	flagCounterpartyDomain  = "counterparty-domain"
)

/*
Create ISM (MsgCreateMessageIdMultisigIsmRaw)

	Why first? The Interchain Security Module (ISM) acts as the security layer for the mailbox. The Mailbox must be linked to a default ISM at the moment of its creation. Therefore, the ISM has to exist before you can create the mailbox.

Create Mailbox (MsgCreateMailbox)

	Dependency: This step requires the ID of the ISM created in Step 1 to populate its DefaultIsm field.
	Why second? It's the central component. Both the Merkle Hook and the Warp Token will need to be associated with this specific Mailbox, so it must be created before them.

Create Merkle Tree Hook (MsgCreateMerkleTreeHook)

	Dependency: This hook needs to be linked to a specific MailboxId to track its dispatched messages. It cannot be created without a valid Mailbox ID from Step 2.

Create IGP (MsgCreateIgp)

	Dependency: The Interchain Gas Paymaster is mostly independent. It doesn't rely on the Mailbox or ISM for its creation. It can be created at any point before it's needed to pay for gas. Placing it here is perfectly fine and doesn't violate any dependencies.

Create Synthetic Token (MsgCreateSyntheticToken)

	Dependency: Similar to the Merkle Hook, the Warp Token must be associated with an OriginMailbox. This requires the Mailbox ID from Step 2.

Enroll Remote Router (MsgEnrollRemoteRouter)

	Dependency: This is the final link in the chain. Enrolling a router connects a specific token on your chain to its counterpart on a remote chain. This requires the TokenId of the synthetic token you created in Step 5.

	./hypd tx hyperlane mailbox set [mailbox-id] --default-hook [igp-hook-id] --required-hook [merkle-tree-hook-id] $HYPD_FLAGS
*/
func CmdSetupBridge() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup-bridge ",
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

			validatorsStr, _ := cmd.Flags().GetString(flagValidators)
			threshold, _ := cmd.Flags().GetUint32(flagThreshold)
			gasDenom, _ := cmd.Flags().GetString(flagGasDenom)
			remoteRouterAddr, _ := cmd.Flags().GetString(flagRemoteRouterAddress)
			remoteRouterGas, _ := cmd.Flags().GetUint64(flagRemoteRouterGas)
			hubDomain, _ := cmd.Flags().GetUint32(flagHubDomain)
			counterpartyDomain, _ := cmd.Flags().GetUint32(flagCounterpartyDomain)

			fmt.Println("validatorsStr", validatorsStr)
			fmt.Println("threshold", threshold)
			fmt.Println("gasDenom", gasDenom)
			fmt.Println("remoteRouterAddr", remoteRouterAddr)
			fmt.Println("remoteRouterGas", remoteRouterGas)
			fmt.Println("hubDomain", hubDomain)
			fmt.Println("counterpartyDomain", counterpartyDomain)

			// these are 20 byte long ethereum style addresses
			validators := strings.Split(validatorsStr, ",")
			if len(validators) == 0 || validators[0] == "" {
				return fmt.Errorf("validators flag cannot be empty")
			}

			ismId, err := createIsm(sCtx, validators, threshold)
			if err != nil {
				return fmt.Errorf("create ism: %w", err)
			}

			mailboxId, err := createMailbox(sCtx, ismId, hubDomain)
			if err != nil {
				return fmt.Errorf("create mailbox: %w", err)
			}

			tokenId, err := createSyntheticToken(sCtx, mailboxId)
			if err != nil {
				return fmt.Errorf("create synthetic token: %w", err)
			}

			merkleHookId, err := createMerkleHook(sCtx, mailboxId)
			if err != nil {
				return fmt.Errorf("create merkle hook: %w", err)
			}

			feeHookId, err := createBridgingFeeHook(sCtx, []bridgingfeetypes.HLAssetFee{{
				TokenId:     tokenId,
				InboundFee:  math.LegacyZeroDec(),
				OutboundFee: math.LegacyNewDecWithPrec(1, 2), // 0.01 == 1%,
			}})
			if err != nil {
				return fmt.Errorf("create bridging fee hook: %w", err)
			}

			aggretaionHookId, err := createAggregationHook(sCtx, []util.HexAddress{feeHookId, merkleHookId})
			if err != nil {
				return fmt.Errorf("create aggregation hook: %w", err)
			}

			if err := createIgp(sCtx, gasDenom); err != nil { // TODO: use it
				return fmt.Errorf("create igp: %w", err)
			}

			noopHookId, err := createNoopHook(sCtx)
			if err != nil {
				return fmt.Errorf("create noop hook: %w", err)
			}

			if err := setMailbox(sCtx, mailboxId, noopHookId, aggretaionHookId); err != nil {
				return fmt.Errorf("set mailbox: %w", err)
			}

			if err := enrollRemoteRouter(sCtx, tokenId, counterpartyDomain, remoteRouterAddr, remoteRouterGas); err != nil {
				return fmt.Errorf("enroll remote router: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().String(flagValidators, "", "Comma-separated list of validator hex addresses (20 bytes ethereum style)")
	cmd.Flags().Uint32(flagThreshold, 1, "Multisig threshold for the ISM")
	cmd.Flags().String(flagGasDenom, "stake", "The denomination to be used for interchain gas payments")
	cmd.Flags().String(flagRemoteRouterAddress, "", "The hex address of the remote router contract on the counterparty chain")
	cmd.Flags().Uint64(flagRemoteRouterGas, 200000, "The gas limit to use for transfers to the remote router")
	cmd.Flags().Uint32(flagHubDomain, 0, "The Hyperlane domain ID for the hub")
	cmd.Flags().Uint32(flagCounterpartyDomain, 0, "The Hyperlane domain ID for the counterparty chain")

	_ = cmd.MarkFlagRequired(flagValidators)
	_ = cmd.MarkFlagRequired(flagThreshold)
	_ = cmd.MarkFlagRequired(flagRemoteRouterAddress)
	_ = cmd.MarkFlagRequired(flagHubDomain)
	_ = cmd.MarkFlagRequired(flagCounterpartyDomain)

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

	if len(respAfter.Isms) == 0 {
		return util.HexAddress{}, fmt.Errorf("no isms found")
	}

	anyIsm := respAfter.Isms[0]
	var ism ismtypes.HyperlaneInterchainSecurityModule
	if err := ctx.registry.UnpackAny(anyIsm, &ism); err != nil {
		return util.HexAddress{}, fmt.Errorf("unpack ism any: %w", err)
	}
	id, _ := ism.GetId()
	return id, nil
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

func createMerkleHook(ctx setupCtx, mailboxId util.HexAddress) (util.HexAddress, error) {
	queryClient := pdtypes.NewQueryClient(ctx.clientCtx)

	msg := &pdtypes.MsgCreateMerkleTreeHook{
		Owner:     ctx.clientCtx.GetFromAddress().String(),
		MailboxId: mailboxId,
	}

	if err := broadcastAndWait(ctx, msg); err != nil {
		return util.HexAddress{}, fmt.Errorf("broadcast create hook tx: %w", err)
	}

	hooksAfter, err := queryClient.MerkleTreeHooks(context.Background(), &pdtypes.QueryMerkleTreeHooksRequest{Pagination: &query.PageRequest{Limit: 1000}})
	if err != nil {
		return util.HexAddress{}, fmt.Errorf("query hooks after creation: %w", err)
	}

	if len(hooksAfter.MerkleTreeHooks) == 0 {
		return util.HexAddress{}, fmt.Errorf("no merkle tree hooks found")
	}

	hook := hooksAfter.MerkleTreeHooks[0]
	hookId, err := util.DecodeHexAddress(hook.Id)
	if err != nil {
		return util.HexAddress{}, fmt.Errorf("decode created hook ID '%s': %w", hook.Id, err)
	}
	return hookId, nil
}

func createBridgingFeeHook(ctx setupCtx, fees []bridgingfeetypes.HLAssetFee) (util.HexAddress, error) {
	queryClient := bridgingfeetypes.NewQueryClient(ctx.clientCtx)

	msg := &bridgingfeetypes.MsgCreateBridgingFeeHook{
		Owner: ctx.clientCtx.GetFromAddress().String(),
		Fees:  fees,
	}

	if err := broadcastAndWait(ctx, msg); err != nil {
		return util.HexAddress{}, fmt.Errorf("broadcast create fee hook tx: %w", err)
	}

	hooksAfter, err := queryClient.FeeHooks(context.Background(), &bridgingfeetypes.QueryFeeHooksRequest{Pagination: &query.PageRequest{Limit: 1000}})
	if err != nil {
		return util.HexAddress{}, fmt.Errorf("query fee hooks after creation: %w", err)
	}

	if len(hooksAfter.FeeHooks) == 0 {
		return util.HexAddress{}, fmt.Errorf("no briding fee hooks found")
	}

	return hooksAfter.FeeHooks[0].Id, nil
}

func createAggregationHook(ctx setupCtx, hookIds []util.HexAddress) (util.HexAddress, error) {
	queryClient := bridgingfeetypes.NewQueryClient(ctx.clientCtx)

	msg := &bridgingfeetypes.MsgCreateAggregationHook{
		Owner:   ctx.clientCtx.GetFromAddress().String(),
		HookIds: hookIds,
	}

	if err := broadcastAndWait(ctx, msg); err != nil {
		return util.HexAddress{}, fmt.Errorf("broadcast create aggregation hook tx: %w", err)
	}

	hooksAfter, err := queryClient.AggregationHooks(context.Background(), &bridgingfeetypes.QueryAggregationHooksRequest{Pagination: &query.PageRequest{Limit: 1000}})
	if err != nil {
		return util.HexAddress{}, fmt.Errorf("query aggregation hooks after creation: %w", err)
	}

	if len(hooksAfter.AggregationHooks) == 0 {
		return util.HexAddress{}, fmt.Errorf("no aggregation hooks found")
	}

	return hooksAfter.AggregationHooks[0].Id, nil
}

func createNoopHook(ctx setupCtx) (util.HexAddress, error) {
	queryClient := pdtypes.NewQueryClient(ctx.clientCtx)

	msg := &pdtypes.MsgCreateNoopHook{
		Owner: ctx.clientCtx.GetFromAddress().String(),
	}

	if err := broadcastAndWait(ctx, msg); err != nil {
		return util.HexAddress{}, fmt.Errorf("broadcast create noop hook tx: %w", err)
	}

	hooksAfter, err := queryClient.NoopHooks(context.Background(), &pdtypes.QueryNoopHooksRequest{Pagination: &query.PageRequest{Limit: 1000}})
	if err != nil {
		return util.HexAddress{}, fmt.Errorf("query noop hooks after creation: %w", err)
	}

	if len(hooksAfter.NoopHooks) == 0 {
		return util.HexAddress{}, fmt.Errorf("no noop hooks found")
	}

	return hooksAfter.NoopHooks[0].Id, nil
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

func setMailbox(ctx setupCtx, mailboxId util.HexAddress, defaultHookId util.HexAddress, requiredHookId util.HexAddress) error {
	msg := &coretypes.MsgSetMailbox{
		Owner:        ctx.clientCtx.GetFromAddress().String(),
		MailboxId:    mailboxId,
		DefaultHook:  &defaultHookId,
		RequiredHook: &requiredHookId,
	}

	if err := broadcastAndWait(ctx, msg); err != nil {
		return fmt.Errorf("broadcast set mailbox tx: %w", err)
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

	if len(tokensAfter.Tokens) == 0 {
		return util.HexAddress{}, fmt.Errorf("no tokens found")
	}

	token := tokensAfter.Tokens[0]
	tokenId, err := util.DecodeHexAddress(token.Id)
	if err != nil {
		return util.HexAddress{}, fmt.Errorf("decode created Token ID '%s': %w", token.Id, err)
	}
	return tokenId, nil
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
