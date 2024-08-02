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

// Event to fire when refunding a bid of a Sell-Order.
const (
	EventTypeDymNameRefundBid       = ModuleName + "_bid_refund"
	AttributeKeyDymNameRefundBidder = "bidder"
	AttributeKeyDymNameRefundAmount = "amount"
)

// Event to fire when refunding a bid of an Offer-To-Buy.
const (
	EventTypeOtbRefundOffer     = ModuleName + "_offer_refund"
	AttributeKeyOtbRefundBuyer  = "buyer"
	AttributeKeyOtbRefundAmount = "amount"
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

// Event to fire when an OfferToBuy is set into store.
const (
	EventTypeOfferToBuy                   = ModuleName + "_otb"
	AttributeKeyOtbActionName             = "action"
	AttributeKeyOtbId                     = "id"
	AttributeKeyOtbName                   = "name"
	AttributeKeyOtbOfferPrice             = "offer_price"
	AttributeKeyOtbCounterpartyOfferPrice = "counterparty_offer_price"
)

// Event to fire corresponding to the action of CRUD a OfferToBuy.
const (
	AttributeValueOtbActionNameSet    = "set"
	AttributeValueOtbActionNameDelete = "delete"
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
