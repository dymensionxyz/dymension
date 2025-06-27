package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bcp-innovations/hyperlane-cosmos/util"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/spf13/cobra"

	ismtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"
	pdtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/types"
	coretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/types"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	"github.com/dymensionxyz/dymension/v3/x/kas/types"
)

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

			// Step 1: Create MessageIdMultisigISMRaw ISM
			fmt.Println("1. Creating Message ID Multisig ISM (Raw)...")

			ismQueryClient := ismtypes.NewQueryClient(clientCtx)
			ismsBefore, err := ismQueryClient.Isms(context.Background(), &ismtypes.QueryIsmsRequest{Pagination: &query.PageRequest{Limit: 1000}})
			if err != nil {
				return fmt.Errorf("failed to query isms before creation: %w", err)
			}
			existingIsmIds := make(map[string]struct{})
			registry := codectypes.NewInterfaceRegistry()
			ismtypes.RegisterInterfaces(registry)
			for _, anyIsm := range ismsBefore.Isms {
				var ism ismtypes.HyperlaneInterchainSecurityModule
				if err := registry.UnpackAny(anyIsm, &ism); err != nil {
					return fmt.Errorf("failed to unpack ism any: %w", err)
				}
				id, _ := ism.GetId()
				existingIsmIds[id.String()] = struct{}{}
			}

			msgCreateIsm := &ismtypes.MsgCreateMessageIdMultisigIsmRaw{
				Creator:    clientCtx.GetFromAddress().String(),
				Validators: validators,
				Threshold:  threshold,
			}
			if err := tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgCreateIsm); err != nil {
				return fmt.Errorf("failed to broadcast create ism tx: %w", err)
			}
			time.Sleep(6 * time.Second)

			ismsAfter, err := ismQueryClient.Isms(context.Background(), &ismtypes.QueryIsmsRequest{Pagination: &query.PageRequest{Limit: 1000}})
			if err != nil {
				return fmt.Errorf("failed to query isms after creation: %w", err)
			}
			var newIsmIdStr string
			for _, anyIsm := range ismsAfter.Isms {
				var ism ismtypes.HyperlaneInterchainSecurityModule
				if err := registry.UnpackAny(anyIsm, &ism); err != nil {
					return fmt.Errorf("failed to unpack ism any: %w", err)
				}
				id, _ := ism.GetId()
				if _, found := existingIsmIds[id.String()]; !found {
					newIsmIdStr = id.String()
					break
				}
			}
			if newIsmIdStr == "" {
				return fmt.Errorf("could not find newly created ISM")
			}
			ismId, err := util.DecodeHexAddress(newIsmIdStr)
			if err != nil {
				return fmt.Errorf("failed to decode created ISM ID '%s': %w", newIsmIdStr, err)
			}
			fmt.Printf("ISM created successfully. ID: %s\n", ismId.String())

			// Step 2: Create Mailbox
			fmt.Println("2. Creating Mailbox...")
			mailboxQueryClient := coretypes.NewQueryClient(clientCtx)
			mailboxesBefore, err := mailboxQueryClient.Mailboxes(context.Background(), &coretypes.QueryMailboxesRequest{Pagination: &query.PageRequest{Limit: 1000}})
			if err != nil {
				return fmt.Errorf("failed to query mailboxes before creation: %w", err)
			}
			existingMailboxIds := make(map[string]struct{})
			for _, mbox := range mailboxesBefore.Mailboxes {
				existingMailboxIds[mbox.Id.String()] = struct{}{}
			}

			msgCreateMailbox := &coretypes.MsgCreateMailbox{
				Owner:       clientCtx.GetFromAddress().String(),
				DefaultIsm:  ismId,
				LocalDomain: hubDomain,
			}
			if err := tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgCreateMailbox); err != nil {
				return fmt.Errorf("failed to broadcast create mailbox tx: %w", err)
			}
			time.Sleep(6 * time.Second)

			mailboxesAfter, err := mailboxQueryClient.Mailboxes(context.Background(), &coretypes.QueryMailboxesRequest{Pagination: &query.PageRequest{Limit: 1000}})
			if err != nil {
				return fmt.Errorf("failed to query mailboxes after creation: %w", err)
			}
			var newMailboxIdStr string
			for _, mbox := range mailboxesAfter.Mailboxes {
				if _, found := existingMailboxIds[mbox.Id.String()]; !found {
					newMailboxIdStr = mbox.Id.String()
					break
				}
			}
			if newMailboxIdStr == "" {
				return fmt.Errorf("could not find newly created mailbox")
			}
			mailboxId, err := util.DecodeHexAddress(newMailboxIdStr)
			if err != nil {
				return fmt.Errorf("failed to decode created Mailbox ID '%s': %w", newMailboxIdStr, err)
			}
			fmt.Printf("Mailbox created successfully. ID: %s\n", mailboxId.String())

			// Step 3: Create Merkle Tree Hook
			fmt.Println("3. Creating Merkle Tree Hook...")
			hookQueryClient := pdtypes.NewQueryClient(clientCtx)
			hooksBefore, err := hookQueryClient.MerkleTreeHooks(context.Background(), &pdtypes.QueryMerkleTreeHooksRequest{Pagination: &query.PageRequest{Limit: 1000}})
			if err != nil {
				return fmt.Errorf("failed to query hooks before creation: %w", err)
			}
			existingHookIds := make(map[string]struct{})
			for _, hook := range hooksBefore.MerkleTreeHooks {
				existingHookIds[hook.Id] = struct{}{}
			}

			msgCreateHook := &pdtypes.MsgCreateMerkleTreeHook{
				Owner:     clientCtx.GetFromAddress().String(),
				MailboxId: mailboxId,
			}
			if err := tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgCreateHook); err != nil {
				return fmt.Errorf("failed to broadcast create hook tx: %w", err)
			}
			time.Sleep(6 * time.Second)

			hooksAfter, err := hookQueryClient.MerkleTreeHooks(context.Background(), &pdtypes.QueryMerkleTreeHooksRequest{Pagination: &query.PageRequest{Limit: 1000}})
			if err != nil {
				return fmt.Errorf("failed to query hooks after creation: %w", err)
			}
			var newMerkleHookIdStr string
			for _, hook := range hooksAfter.MerkleTreeHooks {
				if _, found := existingHookIds[hook.Id]; !found {
					newMerkleHookIdStr = hook.Id
					break
				}
			}
			if newMerkleHookIdStr == "" {
				return fmt.Errorf("could not find newly created merkle hook")
			}
			fmt.Printf("Merkle Tree Hook created successfully. ID: %s\n", newMerkleHookIdStr)

			// Step 4: Create Interchain Gas Paymaster (IGP)
			fmt.Println("4. Creating Interchain Gas Paymaster (IGP)...")
			msgCreateIgp := &pdtypes.MsgCreateIgp{
				Owner: clientCtx.GetFromAddress().String(),
				Denom: gasDenom,
			}
			if err := tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgCreateIgp); err != nil {
				return fmt.Errorf("failed to broadcast create igp tx: %w", err)
			}
			time.Sleep(6 * time.Second)
			fmt.Println("IGP created successfully.")

			// Step 5: Create Synthetic Token
			fmt.Println("5. Creating Synthetic Token...")
			tokenQueryClient := warptypes.NewQueryClient(clientCtx)
			tokensBefore, err := tokenQueryClient.Tokens(context.Background(), &warptypes.QueryTokensRequest{Pagination: &query.PageRequest{Limit: 1000}})
			if err != nil {
				return fmt.Errorf("failed to query tokens before creation: %w", err)
			}
			existingTokenIds := make(map[string]struct{})
			for _, token := range tokensBefore.Tokens {
				existingTokenIds[token.Id] = struct{}{}
			}

			msgCreateToken := &warptypes.MsgCreateSyntheticToken{
				Owner:         clientCtx.GetFromAddress().String(),
				OriginMailbox: mailboxId,
			}
			if err := tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgCreateToken); err != nil {
				return fmt.Errorf("failed to broadcast create token tx: %w", err)
			}
			time.Sleep(6 * time.Second)

			tokensAfter, err := tokenQueryClient.Tokens(context.Background(), &warptypes.QueryTokensRequest{Pagination: &query.PageRequest{Limit: 1000}})
			if err != nil {
				return fmt.Errorf("failed to query tokens after creation: %w", err)
			}
			var newTokenIdStr string
			for _, token := range tokensAfter.Tokens {
				if _, found := existingTokenIds[token.Id]; !found {
					newTokenIdStr = token.Id
					break
				}
			}
			if newTokenIdStr == "" {
				return fmt.Errorf("could not find newly created token")
			}
			tokenId, err := util.DecodeHexAddress(newTokenIdStr)
			if err != nil {
				return fmt.Errorf("failed to decode created Token ID '%s': %w", newTokenIdStr, err)
			}
			fmt.Printf("Synthetic Token created successfully. ID: %s\n", tokenId.String())

			// Step 6: Enroll Remote Router
			fmt.Println("6. Enrolling Remote Router...")
			remoteRouterContract, err := util.DecodeHexAddress(remoteRouterAddr)
			if err != nil {
				return fmt.Errorf("invalid remote router address '%s': %w", remoteRouterAddr, err)
			}
			msgEnrollRouter := &warptypes.MsgEnrollRemoteRouter{
				Owner:   clientCtx.GetFromAddress().String(),
				TokenId: tokenId,
				RemoteRouter: &warptypes.RemoteRouter{
					ReceiverDomain:   counterpartyDomain,
					ReceiverContract: remoteRouterContract.String(),
					Gas:              sdk.NewIntFromUint64(remoteRouterGas),
				},
			}

			if err := tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msgEnrollRouter); err != nil {
				return fmt.Errorf("failed to enroll remote router: %w", err)
			}

			fmt.Println("Remote Router enrolled successfully.")
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
