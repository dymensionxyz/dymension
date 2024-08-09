package cli

const (
	// maxDymBuyValueInteractingCLI is a hard limit for the number of Dym values can be used for Bidding/Offer... in a single CLI command.
	// It was designed to reduce the risk of input wrong number.
	// If working with higher DYM amount, please go to dApp.
	maxDymBuyValueInteractingCLI = 10_000

	// maxDymSellValueInteractingCLI is a hard limit for the number of Dym values can be used for Selling in a single CLI command.
	// It was designed to reduce the mistake that user input an unexpected high number.
	maxDymSellValueInteractingCLI = 1_000_000

	// adymToDymMultiplier is used in CLI to convert `adym` to `DYM`.
	adymToDymMultiplier = 1e18
)
