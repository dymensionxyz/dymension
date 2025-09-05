package keeper

import (
	"context"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/golang-jwt/jwt/v5"
)

// FastFinalizeWithTEE handles TEE attestation-based fast finalization
func (k msgServer) FastFinalizeWithTEE(goCtx context.Context, msg *types.MsgFastFinalizeWithTEE) (*types.MsgFastFinalizeWithTEEResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	
	// Get TEE config
	params := k.GetParams(ctx)
	teeConfig := params.TeeConfig
	
	if !teeConfig.Enabled {
		return nil, gerrc.ErrFailedPrecondition.Wrap("TEE fast finalization is not enabled")
	}
	
	// Verify the creator is the active sequencer for the rollapp
	rollapp, found := k.GetRollapp(ctx, msg.RollappId)
	if !found {
		return nil, gerrc.ErrNotFound.Wrapf("rollapp: %s", msg.RollappId)
	}
	
	// Get the active sequencer
	proposer := k.SequencerK.GetProposer(ctx, msg.RollappId)
	if proposer.Sentinel() {
		return nil, gerrc.ErrNotFound.Wrap("no active sequencer for rollapp")
	}
	
	if proposer.Address != msg.Creator {
		return nil, gerrc.ErrPermissionDenied.Wrapf("only active sequencer can submit TEE attestation: expected %s, got %s", 
			proposer.Address, msg.Creator)
	}
	
	// Verify the PEM certificate SHA1 matches the configured value
	pemSHA1 := sha1.Sum(msg.PemCert)
	pemSHA1Hex := hex.EncodeToString(pemSHA1[:])
	
	if !strings.EqualFold(pemSHA1Hex, teeConfig.GcpRootCertSha1) {
		return nil, gerrc.ErrInvalidArgument.Wrapf("PEM cert SHA1 mismatch: expected %s, got %s", 
			teeConfig.GcpRootCertSha1, pemSHA1Hex)
	}
	
	// Parse and validate the JWT token
	token, err := jwt.Parse(msg.AttestationToken, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing algorithm
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		
		// Extract x5c certificate chain from headers
		x5cInterface, ok := token.Header["x5c"]
		if !ok {
			return nil, fmt.Errorf("x5c header not found in token")
		}
		
		x5c, ok := x5cInterface.([]interface{})
		if !ok || len(x5c) == 0 {
			return nil, fmt.Errorf("invalid x5c header format")
		}
		
		// Parse the certificate chain
		var certs []*x509.Certificate
		for i, certStr := range x5c {
			certDER, err := base64.StdEncoding.DecodeString(certStr.(string))
			if err != nil {
				return nil, fmt.Errorf("failed to decode certificate %d: %w", i, err)
			}
			
			cert, err := x509.ParseCertificate(certDER)
			if err != nil {
				return nil, fmt.Errorf("failed to parse certificate %d: %w", i, err)
			}
			certs = append(certs, cert)
		}
		
		// Parse the PEM root certificate
		block, _ := pem.Decode(msg.PemCert)
		if block == nil {
			return nil, fmt.Errorf("failed to parse PEM block")
		}
		
		rootCert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse root certificate: %w", err)
		}
		
		// Create root cert pool
		rootPool := x509.NewCertPool()
		rootPool.AddCert(rootCert)
		
		// Create intermediate cert pool
		intermediatePool := x509.NewCertPool()
		for i := 1; i < len(certs)-1; i++ {
			intermediatePool.AddCert(certs[i])
		}
		
		// Verify the certificate chain
		opts := x509.VerifyOptions{
			Roots:         rootPool,
			Intermediates: intermediatePool,
			CurrentTime:   ctx.BlockTime(),
		}
		
		if _, err := certs[0].Verify(opts); err != nil {
			return nil, fmt.Errorf("certificate chain verification failed: %w", err)
		}
		
		// Return the leaf certificate's public key for JWT verification
		return certs[0].PublicKey.(*rsa.PublicKey), nil
	})
	
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to parse/validate JWT token")
	}
	
	if !token.Valid {
		return nil, gerrc.ErrInvalidArgument.Wrap("invalid JWT token")
	}
	
	// Extract and validate claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, gerrc.ErrInvalidArgument.Wrap("failed to extract JWT claims")
	}
	
	// Check token expiration
	if exp, ok := claims["exp"].(float64); ok {
		expTime := time.Unix(int64(exp), 0)
		if ctx.BlockTime().After(expTime) {
			return nil, gerrc.ErrInvalidArgument.Wrap("token has expired")
		}
	}
	
	// Verify the hardware type is TDX
	if hwModel, ok := claims["hwmodel"].(string); ok {
		if hwModel != "GCP-TDX" && hwModel != "TDX" {
			return nil, gerrc.ErrInvalidArgument.Wrapf("invalid hardware model: %s, expected TDX", hwModel)
		}
	}
	
	// Verify the image digest is in the allowed list
	if imageDigest, ok := claims["submods"].(map[string]interface{}); ok {
		if container, ok := imageDigest["container"].(map[string]interface{}); ok {
			if digest, ok := container["image_digest"].(string); ok {
				found := false
				for _, allowed := range teeConfig.AllowedImageDigests {
					if digest == allowed {
						found = true
						break
					}
				}
				if !found {
					return nil, gerrc.ErrInvalidArgument.Wrapf("image digest not in allowed list: %s", digest)
				}
			}
		}
	}
	
	// Verify the nonce matches expected format
	if nonceStr, ok := claims["eat_nonce"].(string); ok {
		// The nonce should be a hash of the state_index, chain_id, and last_block_hash
		expectedNonce := k.calculateTEENonce(ctx, msg.Nonce)
		if nonceStr != expectedNonce {
			return nil, gerrc.ErrInvalidArgument.Wrap("nonce mismatch")
		}
	} else {
		return nil, gerrc.ErrInvalidArgument.Wrap("nonce not found in token")
	}
	
	// Verify the state index matches what's in the nonce
	if msg.StateIndex != msg.Nonce.StateIndex {
		return nil, gerrc.ErrInvalidArgument.Wrap("state index mismatch between message and nonce")
	}
	
	// Verify the chain ID matches the rollapp
	if msg.Nonce.ChainId != rollapp.RollappId {
		return nil, gerrc.ErrInvalidArgument.Wrapf("chain ID mismatch: expected %s, got %s", 
			rollapp.RollappId, msg.Nonce.ChainId)
	}
	
	// Fast finalize the states up to the given state index
	statesFinalized, err := k.FastFinalizeStatesWithTEE(ctx, msg.RollappId, msg.StateIndex)
	if err != nil {
		return nil, errorsmod.Wrap(err, "failed to fast finalize states")
	}
	
	// Emit event
	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeTEEFastFinalization,
			sdk.NewAttribute(types.AttributeKeyRollappId, msg.RollappId),
			sdk.NewAttribute(types.AttributeKeyStateIndex, fmt.Sprintf("%d", msg.StateIndex)),
			sdk.NewAttribute(types.AttributeKeyStatesFinalized, fmt.Sprintf("%d", statesFinalized)),
		),
	)
	
	return &types.MsgFastFinalizeWithTEEResponse{
		StatesFinalized: statesFinalized,
	}, nil
}

// calculateTEENonce calculates the expected nonce hash
func (k msgServer) calculateTEENonce(ctx sdk.Context, nonce types.TEENonce) string {
	// Create a deterministic string from the nonce data
	nonceData := fmt.Sprintf("%d:%s:%x", nonce.StateIndex, nonce.ChainId, nonce.LastBlockHash)
	
	// Calculate SHA1 hash (matching what TEE would calculate)
	hash := sha1.Sum([]byte(nonceData))
	return hex.EncodeToString(hash[:])
}

// FastFinalizeStatesWithTEE finalizes states up to the given state index
func (k msgServer) FastFinalizeStatesWithTEE(ctx sdk.Context, rollappID string, stateIndex uint64) (uint64, error) {
	// Get the finalization queue for this rollapp
	queue, err := k.GetFinalizationQueueByRollapp(ctx, rollappID)
	if err != nil {
		return 0, errorsmod.Wrap(err, "failed to get finalization queue")
	}
	
	statesFinalized := uint64(0)
	for _, item := range queue {
		// Process each state in the finalization queue for this creation height
		for _, stateInfoIdx := range item.FinalizationQueue {
			// Only finalize states up to the specified index
			if stateInfoIdx.Index > stateIndex {
				continue
			}
			
			// Check if already finalized
			stateInfo, found := k.GetStateInfo(ctx, rollappID, stateInfoIdx.Index)
			if !found {
				continue
			}
			
			if stateInfo.Status == commontypes.Status_FINALIZED {
				continue
			}
			
			// Finalize this state
			stateInfo.Status = commontypes.Status_FINALIZED
			k.SetStateInfo(ctx, stateInfo)
			
			statesFinalized++
			
			// Emit finalization event for each state
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventTypeStateFinalized,
					sdk.NewAttribute(types.AttributeKeyRollappId, rollappID),
					sdk.NewAttribute(types.AttributeKeyStateIndex, fmt.Sprintf("%d", stateInfoIdx.Index)),
					sdk.NewAttribute(types.AttributeKeyFinalizedBy, "TEE"),
				),
			)
		}
		
		// Remove entire queue entry once all its states are processed
		if len(item.FinalizationQueue) > 0 {
			err = k.RemoveFinalizationQueue(ctx, item.CreationHeight, rollappID)
			if err != nil {
				// Log but don't fail
				ctx.Logger().Error("Failed to remove from finalization queue", "error", err)
			}
		}
	}
	
	return statesFinalized, nil
}