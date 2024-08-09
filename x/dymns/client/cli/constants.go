package cli

const (
	// MaxDymBuyValueInteractingCLI is a hard limit for the number of Dym values can be used for Bidding/Offer... in a single CLI command.
	// It was designed to reduce the risk of input wrong number.
	// If working with higher DYM amount, please go to dApp.
	MaxDymBuyValueInteractingCLI = 10_000

	// MaxDymSellValueInteractingCLI is a hard limit for the number of Dym values can be used for Selling in a single CLI command.
	// It was designed to reduce the mistake that user input an unexpected high number.
	MaxDymSellValueInteractingCLI = 1_000_000
)
