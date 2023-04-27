package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
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
			rest.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get latest block time")
			return
		}

		if time.Now().UTC().Sub(latestTime).Minutes() > maxNoHeightProgressDuration.Minutes() {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, fmt.Sprintf("Node is not syncing. Last block time: %s, curr time: %s", latestTime.String(), time.Now().UTC().String()))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
