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

func TestAutoCLIInvoiceSettlementQuery(t *testing.T) {
	for _, command := range (AppModule{}).AutoCLIOptions().Query.RpcCommandOptions {
		if command.RpcMethod == "InvoiceSettlement" {
			require.Len(t, command.PositionalArgs, 3)
			require.Equal(t, "agent_id", command.PositionalArgs[0].ProtoField)
			require.Equal(t, "recipient", command.PositionalArgs[1].ProtoField)
			require.Equal(t, "invoice_ref", command.PositionalArgs[2].ProtoField)
			return
		}
	}
	t.Fatal("InvoiceSettlement query not registered")
}
