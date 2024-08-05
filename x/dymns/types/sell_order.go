package types

import (
	"fmt"
	"sort"

	errorsmod "cosmossdk.io/errors"

	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
)

// GetIdentity returns the unique identity of the SO
func (m *SellOrder) GetIdentity() string {
	return fmt.Sprintf("%s|%d|%d", m.GoodsId, m.Type, m.ExpireAt)
}

// HasSetSellPrice returns true if the sell price is set
func (m *SellOrder) HasSetSellPrice() bool {
	return m.SellPrice != nil && !m.SellPrice.Amount.IsNil() && !m.SellPrice.IsZero()
}

// HasExpiredAtCtx returns true if the SO has expired at given context
func (m *SellOrder) HasExpiredAtCtx(ctx sdk.Context) bool {
	return m.HasExpired(ctx.BlockTime().Unix())
}

// HasExpired returns true if the SO has expired at given epoch
func (m *SellOrder) HasExpired(nowEpoch int64) bool {
	return m.ExpireAt < nowEpoch
}

// HasFinishedAtCtx returns true if the SO has expired or completed at given context
func (m *SellOrder) HasFinishedAtCtx(ctx sdk.Context) bool {
	return m.HasFinished(ctx.BlockTime().Unix())
}

// HasFinished returns true if the SO has expired or completed at given epoch
func (m *SellOrder) HasFinished(nowEpoch int64) bool {
	if m.HasExpired(nowEpoch) {
		return true
	}

	if !m.HasSetSellPrice() {
		// when no sell price is set, must wait until completed auction
		return false
	}

	// complete condition: bid >= sell price

	if m.HighestBid == nil {
		return false
	}

	return m.HighestBid.Price.IsGTE(*m.SellPrice)
}

// Validate performs basic validation for the SellOrder.
func (m *SellOrder) Validate() error {
	if m == nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "SO is nil")
	}

	if m.Type == MarketOrderType_MOT_DYM_NAME {
		if m.GoodsId == "" {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "Dym-Name of SO is empty")
		}

		if !dymnsutils.IsValidDymName(m.GoodsId) {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "Dym-Name of SO is not a valid dym name")
		}
	} else if m.Type == MarketOrderType_MOT_ALIAS {
		if m.GoodsId == "" {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "alias of SO is empty")
		}

		if !dymnsutils.IsValidAlias(m.GoodsId) {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "alias of SO is not a valid alias")
		}
	} else {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "invalid SO type: %s", m.Type)
	}

	if m.ExpireAt == 0 {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "SO expiry is empty")
	}

	if m.MinPrice.Amount.IsNil() || m.MinPrice.IsZero() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "SO min price is zero")
	} else if m.MinPrice.IsNegative() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "SO min price is negative")
	} else if err := m.MinPrice.Validate(); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "SO min price is invalid: %v", err)
	}

	if m.HasSetSellPrice() {
		if m.SellPrice.IsNegative() {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "SO sell price is negative")
		} else if err := m.SellPrice.Validate(); err != nil {
			return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "SO sell price is invalid: %v", err)
		}

		if m.SellPrice.Denom != m.MinPrice.Denom {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "SO sell price denom is different from min price denom")
		}

		if m.SellPrice.IsLT(m.MinPrice) {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "SO sell price is less than min price")
		}
	}

	if m.HighestBid == nil {
		// valid, means no bid yet
	} else if err := m.HighestBid.Validate(); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "SO highest bid is invalid: %v", err)
	} else if m.HighestBid.Price.IsLT(m.MinPrice) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "SO highest bid price is less than min price")
	} else if m.HasSetSellPrice() && m.SellPrice.IsLT(m.HighestBid.Price) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "SO sell price is less than highest bid price")
	}

	return nil
}

// Validate performs basic validation for the SellOrderBid.
func (m *SellOrderBid) Validate() error {
	if m == nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "SO bid is nil")
	}

	if m.Bidder == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "SO bidder is empty")
	}

	if !dymnsutils.IsValidBech32AccountAddress(m.Bidder, true) {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "SO bidder is not a valid bech32 account address")
	}

	if m.Price.Amount.IsNil() || m.Price.IsZero() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "SO bid price is zero")
	} else if m.Price.IsNegative() {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "SO bid price is negative")
	} else if err := m.Price.Validate(); err != nil {
		return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "SO bid price is invalid: %v", err)
	}

	return nil
}

// Validate performs basic validation for the HistoricalSellOrders.
func (m *HistoricalSellOrders) Validate() error {
	if m == nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "historical SOs is nil")
	}

	if len(m.SellOrders) > 0 {
		goodsId := m.SellOrders[0].GoodsId
		orderType := m.SellOrders[0].Type
		uniqueIdentity := make(map[string]bool)
		// Describe usage of Go Map: only used for validation
		for _, so := range m.SellOrders {
			if err := so.Validate(); err != nil {
				return err
			}

			if so.GoodsId != goodsId {
				return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "historical SOs have different goods ID: %s != %s", goodsId, so.GoodsId)
			}
			if so.Type != orderType {
				return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "historical SOs have different order type: %s != %s", orderType, so.Type)
			}

			if _, duplicated := uniqueIdentity[so.GetIdentity()]; duplicated {
				return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "historical SO is not unique: %s", so.GetIdentity())
			}
			uniqueIdentity[so.GetIdentity()] = true
		}
	}

	return nil
}

// GetSdkEvent returns the sdk event contains information of Sell-Order record.
// Fired when Sell-Order record is set into store.
func (m SellOrder) GetSdkEvent(actionName string) sdk.Event {
	var sellPrice sdk.Coin
	if m.HasSetSellPrice() {
		sellPrice = *m.SellPrice
	} else {
		sellPrice = sdk.NewCoin(m.MinPrice.Denom, sdk.ZeroInt())
	}

	var attrHighestBidder, attrHighestBidPrice sdk.Attribute
	if m.HighestBid != nil {
		attrHighestBidder = sdk.NewAttribute(AttributeKeySoHighestBidder, m.HighestBid.Bidder)
		attrHighestBidPrice = sdk.NewAttribute(AttributeKeySoHighestBidPrice, m.HighestBid.Price.String())
	} else {
		attrHighestBidder = sdk.NewAttribute(AttributeKeySoHighestBidder, "")
		attrHighestBidPrice = sdk.NewAttribute(AttributeKeySoHighestBidPrice, sdk.NewCoin(m.MinPrice.Denom, sdk.ZeroInt()).String())
	}

	return sdk.NewEvent(
		EventTypeSellOrder,
		sdk.NewAttribute(AttributeKeySoGoodsId, m.GoodsId),
		sdk.NewAttribute(AttributeKeySoType, m.Type.String()),
		sdk.NewAttribute(AttributeKeySoExpiryEpoch, fmt.Sprintf("%d", m.ExpireAt)),
		sdk.NewAttribute(AttributeKeySoMinPrice, m.MinPrice.String()),
		sdk.NewAttribute(AttributeKeySoSellPrice, sellPrice.String()),
		attrHighestBidder,
		attrHighestBidPrice,
		sdk.NewAttribute(AttributeKeySoActionName, actionName),
	)
}

func (m ActiveSellOrdersExpiration) Validate() error {
	if len(m.Records) > 0 {
		uniqueName := make(map[string]bool)
		// Describe usage of Go Map: only used for validation
		allNames := make([]string, len(m.Records))

		for i, record := range m.Records {
			if record.ExpireAt < 1 {
				return errorsmod.Wrap(gerrc.ErrInvalidArgument, "active SO expiry is empty")
			}

			if _, duplicated := uniqueName[record.GoodsId]; duplicated {
				return errorsmod.Wrapf(gerrc.ErrInvalidArgument, "active SO is not unique: %s", record.GoodsId)
			}

			uniqueName[record.GoodsId] = true
			allNames[i] = record.GoodsId
		}

		if !sort.StringsAreSorted(allNames) {
			return errorsmod.Wrap(gerrc.ErrInvalidArgument, "active SO names are not sorted")
		}
	}

	return nil
}

func (m *ActiveSellOrdersExpiration) Sort() {
	if len(m.Records) < 2 {
		return
	}

	// TODO DymNS: add sort by other fields if any added

	sort.Slice(m.Records, func(i, j int) bool {
		return m.Records[i].GoodsId < m.Records[j].GoodsId
	})
}

func (m *ActiveSellOrdersExpiration) Add(goodsId string, expiry int64) {
	newRecord := ActiveSellOrdersExpirationRecord{GoodsId: goodsId, ExpireAt: expiry}

	if len(m.Records) < 1 {
		m.Records = []ActiveSellOrdersExpirationRecord{newRecord}
		return
	}

	foundAtIdx := -1

	for i, record := range m.Records {
		if record.GoodsId == goodsId {
			foundAtIdx = i
			break
		}
	}

	if foundAtIdx > -1 {
		m.Records[foundAtIdx].ExpireAt = expiry
	} else {
		m.Records = append(m.Records, newRecord)
	}

	m.Sort()
}

func (m *ActiveSellOrdersExpiration) Remove(goodsId string) {
	if len(m.Records) < 1 {
		return
	}

	modified := make([]ActiveSellOrdersExpirationRecord, 0, len(m.Records))
	for _, record := range m.Records {
		if record.GoodsId != goodsId {
			modified = append(modified, record)
		}
	}
	m.Records = modified
}
