package types

// Event to fire when a DymName is set into store.
const (
	EventTypeSetDymName                  = ModuleName + "_name"
	AttributeKeyDymName                  = "name"
	AttributeKeyDymNameOwner             = "owner"
	AttributeKeyDymNameController        = "controller"
	AttributeKeyDymNameExpiryEpoch       = "expiry_epoch"
	AttributeKeyDymNameConfigCount       = "cfg_count"
	AttributeKeyDymNameHasContactDetails = "has_contact"
)

// Event to fire when refunding a deposited bidding amount of a Sell-Order.
const (
	EventTypeSoRefundBid       = ModuleName + "_bid_refund"
	AttributeKeySoRefundBidder = "bidder"
	AttributeKeySoRefundAmount = "amount"
)

// Event to fire when refunding deposited amount of a Buy-Order.
const (
	EventTypeBoRefundOffer     = ModuleName + "_bo_refund"
	AttributeKeyBoRefundBuyer  = "buyer"
	AttributeKeyBoRefundAmount = "amount"
)

// Event to fire when a SellOrder is set into store.
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

// Event to fire corresponding to the action of CRUD a SellOrder.
const (
	AttributeValueDymNameSoActionNameSet    = "set"
	AttributeValueDymNameSoActionNameDelete = "delete"
)

// Event to fire when a BuyOffer is set into store.
const (
	EventTypeBuyOffer                    = ModuleName + "_bo"
	AttributeKeyBoActionName             = "action"
	AttributeKeyBoId                     = "id"
	AttributeKeyBoName                   = "name"
	AttributeKeyBoOfferPrice             = "offer_price"
	AttributeKeyBoCounterpartyOfferPrice = "counterparty_offer_price"
)

// Event to fire corresponding to the action of CRUD a BuyOffer.
const (
	AttributeValueBoActionNameSet    = "set"
	AttributeValueBoActionNameDelete = "delete"
)

const (
	EventTypeSell         = ModuleName + "_sell"
	AttributeKeySellType  = "type"
	AttributeKeySellName  = "name"
	AttributeKeySellPrice = "price"
	AttributeKeySellTo    = "buyer"
)

const (
	AttributeValueSellTypeName  = "name"
	AttributeValueSellTypeAlias = "alias"
)
