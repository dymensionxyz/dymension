package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/gorilla/mux"
)

// Constants for health check thresholds.
const (
	maxNoHeightProgressDuration = 2 * time.Minute
	healthCheckTimeout          = 5 * time.Second // Fail fast if the node is unresponsive
)

// HealthcheckRegister registers the health check route on the provided router.
func HealthcheckRegister(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc("/healthcheck", HealthcheckRequestHandlerFn(clientCtx)).Methods(http.MethodGet)
}

// getLatestBlockTime retrieves the timestamp of the latest block committed by the node.
// It accepts a context to handle timeouts and cancellation properly, avoiding zombie routines.
func getLatestBlockTime(ctx context.Context, clientCtx client.Context) (time.Time, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get node client: %w", err)
	}

	// Fetch node status using the provided context (propagates timeout).
	status, err := node.Status(ctx)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to fetch node status: %w", err)
	}

	return status.SyncInfo.LatestBlockTime, nil
}

// HealthcheckRequestHandlerFn returns an HTTP handler for liveness/readiness checks.
func HealthcheckRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Create a context with a strict timeout.
		// If the node RPC is dead/deadlocked, we want to fail fast rather than hang the connection.
		ctx, cancel := context.WithTimeout(r.Context(), healthCheckTimeout)
		defer cancel()

		latestTime, err := getLatestBlockTime(ctx, clientCtx)
		if err != nil {
			// If we can't talk to the node, it's an internal error.
			writeJSONResponse(w, http.StatusInternalServerError, errorResponse{
				Status: "error",
				Error:  "Failed to get latest block time: " + err.Error(),
			})
			return
		}

		// Check if the node is stalling.
		// time.Since is cleaner and safer than manual subtraction with time.Now().
		elapsed := time.Since(latestTime)
		if elapsed > maxNoHeightProgressDuration {
			errMsg := fmt.Sprintf("Node is not syncing. Last block time: %s, time since: %s",
				latestTime.UTC().Format(time.RFC3339),
				elapsed.String(),
			)

			// 503 Service Unavailable is more appropriate for a sync lag than 500 Internal Error.
			writeJSONResponse(w, http.StatusServiceUnavailable, errorResponse{
				Status: "error",
				Error:  errMsg,
			})
			return
		}

		// Respond with a structured success message and 200 OK.
		writeJSONResponse(w, http.StatusOK, map[string]string{
			"status":       "ok",
			"latest_block": latestTime.UTC().Format(time.RFC3339),
		})
	}
}

// errorResponse defines a standard structure for API errors.
type errorResponse struct {
	Status string `json:"status"`          // e.g., "error" or "ok"
	Error  string `json:"error,omitempty"` // Detailed error message
}

// writeJSONResponse is a helper to marshal data to JSON and write it to the response writer.
// It replaces the heavy Cosmos Legacy Amino codec with the standard encoding/json library.
func writeJSONResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		// If we can't encode the response, log it to server stdout (best effort).
		// In a real logger setup, you would use your logger instance here.
		fmt.Printf("failed to encode json response: %v\n", err)
	}
}
