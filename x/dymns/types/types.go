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
	AttributeKeySoAssetId         = "asset_id"
	AttributeKeySoAssetType       = "asset_type"
	AttributeKeySoExpiryEpoch     = "expiry_epoch"
	AttributeKeySoMinPrice        = "min_price"
	AttributeKeySoSellPrice       = "sell_price"
	AttributeKeySoHighestBidder   = "highest_bidder"
	AttributeKeySoHighestBidPrice = "highest_bid_price"
)

// Event to fire corresponding to the action of CRUD a SellOrder.
const (
	AttributeValueSoActionNameSet    = "set"
	AttributeValueSoActionNameDelete = "delete"
)

// Event to fire when a BuyOrder is set into store.
const (
	EventTypeBuyOrder                    = ModuleName + "_bo"
	AttributeKeyBoActionName             = "action"
	AttributeKeyBoId                     = "id"
	AttributeKeyBoAssetId                = "asset_id"
	AttributeKeyBoAssetType              = "asset_type"
	AttributeKeyBoBuyer                  = "buyer"
	AttributeKeyBoOfferPrice             = "offer_price"
	AttributeKeyBoCounterpartyOfferPrice = "counterparty_offer_price"
)

// Event to fire corresponding to the action of CRUD a BuyOrder.
const (
	AttributeValueBoActionNameSet    = "set"
	AttributeValueBoActionNameDelete = "delete"
)

const (
	EventTypeSell = ModuleName + "_sell"
	// TODO DymNS: fires this for alias in hook
	AttributeKeySellAssetType = "asset_type"
	AttributeKeySellName      = "name"
	AttributeKeySellPrice     = "price"
	AttributeKeySellTo        = "buyer"
)

const (
	AttributeValueSellTypeName = "name"

	AttributeValueSellTypeAlias = "alias"
)
