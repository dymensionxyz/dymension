package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	"github.com/gorilla/mux"
)

const maxNoHeightProgressDuration = 2 * time.Minute

func HealthcheckRegister(clientCtx client.Context, r *mux.Router) {
	r.HandleFunc("/healthcheck", HealthcheckRequestHandlerFn(clientCtx)).Methods("GET")
}

func getLatestBlockTime(clientCtx client.Context) (time.Time, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return time.Time{}, err
	}

	status, err := node.Status(context.Background())
	if err != nil {
		return time.Time{}, err
	}

	return status.SyncInfo.LatestBlockTime, nil
}

func HealthcheckRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		latestTime, err := getLatestBlockTime(clientCtx)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to get latest block time")
			return
		}

		if time.Now().UTC().Sub(latestTime).Minutes() > maxNoHeightProgressDuration.Minutes() {
			writeErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Node is not syncing. Last block time: %s, curr time: %s", latestTime.String(), time.Now().UTC().String()))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// errorResponse defines the attributes of a JSON error response.
type errorResponse struct {
	Code  int    `json:"code,omitempty"`
	Error string `json:"error"`
}

// newErrorResponse creates a new errorResponse instance.
func newErrorResponse(code int, err string) errorResponse {
	return errorResponse{Code: code, Error: err}
}

func writeErrorResponse(w http.ResponseWriter, status int, err string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(legacy.Cdc.MustMarshalJSON(newErrorResponse(0, err)))
}
