package simulation

const (
	// liquidity module simulation operation weights for messages
	DefaultWeightMsgCreatePool          int = 5
	DefaultWeightMsgDepositWithinBatch  int = 10
	DefaultWeightMsgWithdrawWithinBatch int = 10
	DefaultWeightMsgSwapWithinBatch     int = 85
)
