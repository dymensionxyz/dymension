package agent

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAutoCLIValidationQueries(t *testing.T) {
	options := (AppModule{}).AutoCLIOptions()
	methods := make(map[string]bool)
	for _, command := range options.Query.RpcCommandOptions {
		methods[command.RpcMethod] = true
	}
	for _, method := range []string{"ValidationRequest", "ValidationResponses", "ValidationRequestsByAgent"} {
		require.True(t, methods[method], method)
	}
}
