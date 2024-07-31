package types

const (
	EventTypeSetDymName                  = ModuleName + "_name"
	AttributeKeyDymName                  = "name"
	AttributeKeyDymNameOwner             = "owner"
	AttributeKeyDymNameController        = "controller"
	AttributeKeyDymNameExpiryEpoch       = "expiry_epoch"
	AttributeKeyDymNameConfigCount       = "cfg_count"
	AttributeKeyDymNameHasContactDetails = "has_contact"
)

const (
	EventTypeDymNameRefundBid       = ModuleName + "_bid_refund"
	AttributeKeyDymNameRefundBidder = "bidder"
	AttributeKeyDymNameRefundAmount = "amount"
)

const (
	EventTypeOtbRefundOffer     = ModuleName + "_offer_refund"
	AttributeKeyOtbRefundBuyer  = "buyer"
	AttributeKeyOtbRefundAmount = "amount"
)

const (
	EventTypeSellOrder            = ModuleName + "_so"
	AttributeKeySoActionName      = "action"
	AttributeKeySoName            = "name"
	AttributeKeySoExpiryEpoch     = "expiry_epoch"
	AttributeKeySoMinPrice        = "min_price"
	AttributeKeySoSellPrice       = "sell_price"
	AttributeKeySoHighestBidder   = "highest_bidder"
	AttributeKeySoHighestBidPrice = "highest_bid_price"
)

const (
	AttributeKeyDymNameSoActionNameSet    = "set"
	AttributeKeyDymNameSoActionNameDelete = "delete"
)

const (
	EventTypeOfferToBuy                   = ModuleName + "_otb"
	AttributeKeyOtbActionName             = "action"
	AttributeKeyOtbId                     = "id"
	AttributeKeyOtbName                   = "name"
	AttributeKeyOtbOfferPrice             = "offer_price"
	AttributeKeyOtbCounterpartyOfferPrice = "counterparty_offer_price"
)

const (
	AttributeKeyOtbActionNameSet    = "set"
	AttributeKeyOtbActionNameDelete = "delete"
)
