package apptesting

import (
	"encoding/json"
	"testing"
	"time"

	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	app "github.com/dymensionxyz/dymension/v3/app"

	"github.com/stretchr/testify/require"
)

func TestMessageAuthzSerialization(t *testing.T, msg sdk.Msg) {
	someDate := time.Date(1, 1, 1, 1, 1, 1, 1, time.UTC)
	const (
		mockGranter string = "cosmos1abc"
		mockGrantee string = "cosmos1xyz"
	)

	var (
		mockMsgGrant  authz.MsgGrant
		mockMsgRevoke authz.MsgRevoke
		mockMsgExec   authz.MsgExec
	)

	encCdc := app.MakeEncodingConfig()
	amino := encCdc.Amino

	// Authz: Grant Msg
	typeURL := sdk.MsgTypeURL(msg)
	expDate := someDate.Add(time.Hour)
	grant, err := authz.NewGrant(someDate, authz.NewGenericAuthorization(typeURL), &expDate)
	require.NoError(t, err)

	msgGrant := authz.MsgGrant{Granter: mockGranter, Grantee: mockGrantee, Grant: grant}
	msgGrantBytes := json.RawMessage(sdk.MustSortJSON(amino.MustMarshalJSON(&msgGrant)))
	err = amino.UnmarshalJSON(msgGrantBytes, &mockMsgGrant)
	require.NoError(t, err)

	// Authz: Revoke Msg
	msgRevoke := authz.MsgRevoke{Granter: mockGranter, Grantee: mockGrantee, MsgTypeUrl: typeURL}
	msgRevokeByte := json.RawMessage(sdk.MustSortJSON(amino.MustMarshalJSON(&msgRevoke)))
	err = amino.UnmarshalJSON(msgRevokeByte, &mockMsgRevoke)
	require.NoError(t, err)

	// Authz: Exec Msg
	msgAny, err := cdctypes.NewAnyWithValue(msg)
	require.NoError(t, err)
	msgExec := authz.MsgExec{Grantee: mockGrantee, Msgs: []*cdctypes.Any{msgAny}}
	execMsgByte := json.RawMessage(sdk.MustSortJSON(amino.MustMarshalJSON(&msgExec)))
	err = amino.UnmarshalJSON(execMsgByte, &mockMsgExec)
	require.NoError(t, err)
	require.Equal(t, msgExec.Msgs[0].Value, mockMsgExec.Msgs[0].Value)
}
