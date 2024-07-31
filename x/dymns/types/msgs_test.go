package types

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/stretchr/testify/require"
)

func TestMsgs_Signers(t *testing.T) {
	t.Run("get signers", func(t *testing.T) {
		//goland:noinspection GoDeprecation,SpellCheckingInspection
		msgs := []sdk.Msg{
			&MsgRegisterName{
				Owner: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			},
			&MsgTransferOwnership{
				Owner: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			},
			&MsgSetController{
				Owner: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			},
			&MsgUpdateResolveAddress{
				Controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			},
			&MsgUpdateDetails{
				Controller: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			},
			&MsgPutAdsSellName{
				Owner: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			},
			&MsgCancelAdsSellName{
				Owner: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			},
			&MsgPurchaseName{
				Buyer: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			},
			&MsgOfferBuyName{
				Buyer: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			},
			&MsgCancelOfferBuyName{
				Buyer: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			},
			&MsgAcceptOfferBuyName{
				Owner: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			},
		}

		for _, msg := range msgs {
			require.Len(t, msg.GetSigners(), 1)
		}
	})

	t.Run("bad signers should panic", func(t *testing.T) {
		msgs := []sdk.Msg{
			&MsgRegisterName{},
			&MsgTransferOwnership{},
			&MsgSetController{},
			&MsgUpdateResolveAddress{},
			&MsgUpdateDetails{},
			&MsgPutAdsSellName{},
			&MsgCancelAdsSellName{},
			&MsgPurchaseName{},
			&MsgOfferBuyName{},
			&MsgCancelOfferBuyName{},
			&MsgAcceptOfferBuyName{},
		}

		for _, msg := range msgs {
			require.Panics(t, func() {
				_ = msg.GetSigners()
			})
		}
	})
}

func TestMsgs_ImplementLegacyMsg(t *testing.T) {
	//goland:noinspection GoDeprecation
	msgs := []legacytx.LegacyMsg{
		&MsgRegisterName{},
		&MsgTransferOwnership{},
		&MsgSetController{},
		&MsgUpdateResolveAddress{},
		&MsgUpdateDetails{},
		&MsgPutAdsSellName{},
		&MsgCancelAdsSellName{},
		&MsgPurchaseName{},
		&MsgOfferBuyName{},
		&MsgCancelOfferBuyName{},
		&MsgAcceptOfferBuyName{},
	}

	for _, msg := range msgs {
		require.Equal(t, RouterKey, msg.Route())
		require.NotEmpty(t, msg.Type())
		require.NotEmpty(t, msg.GetSignBytes())
	}
}

func TestMsgs_Type(t *testing.T) {
	require.Equal(t, "register_name", (&MsgRegisterName{}).Type())
	require.Equal(t, "transfer_ownership", (&MsgTransferOwnership{}).Type())
	require.Equal(t, "set_controller", (&MsgSetController{}).Type())
	require.Equal(t, "update_resolve_address", (&MsgUpdateResolveAddress{}).Type())
	require.Equal(t, "update_details", (&MsgUpdateDetails{}).Type())
	require.Equal(t, "put_ads_sell_name", (&MsgPutAdsSellName{}).Type())
	require.Equal(t, "cancel_ads_sell_name", (&MsgCancelAdsSellName{}).Type())
	require.Equal(t, "purchase_name", (&MsgPurchaseName{}).Type())
	require.Equal(t, "offer_buy_name", (&MsgOfferBuyName{}).Type())
	require.Equal(t, "cancel_offer_buy_name", (&MsgCancelOfferBuyName{}).Type())
	require.Equal(t, "accept_offer_buy_name", (&MsgAcceptOfferBuyName{}).Type())
}
