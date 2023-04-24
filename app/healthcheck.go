package app

import (
	"context"
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/types/rest"
	"github.com/gorilla/mux"
)

var (
	LatestBlockHeight int64
	lastResult        bool
)

const progressCheckPeriod = 1 * time.Minute

func HealthcheckRegister(clientCtx client.Context, r *mux.Router) {
	go RunHealthCheck(clientCtx)
	r.HandleFunc("/healthcheck", HealthcheckRequestHandlerFn(clientCtx)).Methods("GET")
}

func getLatestHeight(clientCtx client.Context) (int64, error) {
	node, err := clientCtx.GetNode()
	if err != nil {
		return 0, err
	}

	status, err := node.Status(context.Background())
	if err != nil {
		return 0, err
	}

	return status.SyncInfo.LatestBlockHeight, nil
}

func RunHealthCheck(clientCtx client.Context) {
	LatestBlockHeight, _ := getLatestHeight(clientCtx)
	for {
		select {
		case <-clientCtx.Client.Quit():
			return
		case <-time.After(progressCheckPeriod):
			height, err := getLatestHeight(clientCtx)
			if err != nil {
				lastResult = false
				continue
			}

			if height <= LatestBlockHeight {
				lastResult = false
				continue
			}

			lastResult = true
			LatestBlockHeight = height
		}
	}
}

func HealthcheckRequestHandlerFn(clientCtx client.Context) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !lastResult {
			rest.WriteErrorResponse(w, http.StatusInternalServerError, "Node is not syncing")
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
