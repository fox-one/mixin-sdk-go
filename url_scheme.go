package mixin

import (
	"encoding/base64"
	"net/url"
	"path"
)

const (
	Scheme = "mixin"
)

type SendSchemeCategory = string

const (
	SendSchemeCategoryText    SendSchemeCategory = "text"
	SendSchemeCategoryImage   SendSchemeCategory = "image"
	SendSchemeCategoryContact SendSchemeCategory = "contact"
	SendSchemeCategoryAppCard SendSchemeCategory = "app_card"
	SendSchemeCategoryLive    SendSchemeCategory = "live"
	SendSchemeCategoryPost    SendSchemeCategory = "post"
)

var URL = urlScheme{
	host: "mixin.one",
}

type urlScheme struct {
	host string
}

func (urlScheme) Users(userID string) string {
	u := url.URL{
		Scheme: Scheme,
		Host:   "users",
		Path:   userID,
	}

	return u.String()
}

func (urlScheme) Transfer(userID string) string {
	u := url.URL{
		Scheme: Scheme,
		Host:   "transfer",
		Path:   userID,
	}

	return u.String()
}

func (urlScheme) Pay(input *TransferInput) string {
	q := url.Values{}
	q.Set("asset", input.AssetID)
	q.Set("trace", input.TraceID)
	q.Set("amount", input.Amount.String())
	q.Set("recipient", input.OpponentID)
	q.Set("memo", input.Memo)

	u := url.URL{
		Scheme:   Scheme,
		Host:     "pay",
		RawQuery: q.Encode(),
	}

	return u.String()
}

func (s urlScheme) SafePay(input *TransferInput) string {
	q := url.Values{}

	if input.AssetID != "" {
		q.Set("asset", input.AssetID)
	}

	if input.TraceID != "" {
		q.Set("trace", input.TraceID)
	}

	if input.Amount.IsPositive() {
		q.Set("amount", input.Amount.String())
	}

	if input.Memo != "" {
		q.Set("memo", input.Memo)
	}

	u := url.URL{
		Scheme:   "https",
		Host:     s.host,
		Path:     "/pay",
		RawQuery: q.Encode(),
	}

	if addr, err := NewMixAddress(input.OpponentMultisig.Receivers, input.OpponentMultisig.Threshold); err == nil {
		u.Path = path.Join(u.Path, addr.String())
	} else {
		u.Path = path.Join(u.Path, input.OpponentID)
	}

	return u.String()
}

func (urlScheme) Codes(code string) string {
	u := url.URL{
		Scheme: Scheme,
		Host:   "codes",
		Path:   code,
	}

	return u.String()
}

func (urlScheme) Snapshots(snapshotID, traceID string) string {
	u := url.URL{
		Scheme: Scheme,
		Host:   "snapshots",
	}

	if snapshotID != "" {
		u.Path = snapshotID
	}

	if traceID != "" {
		query := url.Values{}
		query.Set("trace", traceID)
		u.RawQuery = query.Encode()
	}

	return u.String()
}

// Conversations scheme of a conversation
//
//	userID optional, for user conversation only, if there's not conversation with the user, messenger will create the conversation first
//
// https://developers.mixin.one/docs/schema#open-an-conversation
func (urlScheme) Conversations(conversationID, userID string) string {
	u := url.URL{
		Scheme: Scheme,
		Host:   "conversations",
	}

	if conversationID != "" {
		u.Path = conversationID
	}

	if userID != "" {
		query := url.Values{}
		query.Set("user", userID)
		u.RawQuery = query.Encode()
	}

	return u.String()
}

// Apps scheme of an app
//
//	appID required, userID of an app
//	action optional, action about this scheme, default is "open"
//	params optional, parameters of any name or type can be passed when opening the bot homepage to facilitate the development of features like invitation codes, visitor tracking, etc
//
// https://developers.mixin.one/docs/schema#popups-bot-profile
func (urlScheme) Apps(appID, action string, params map[string]string) string {
	u := url.URL{
		Scheme: Scheme,
		Host:   "apps",
	}

	if appID != "" {
		u.Path = appID
	}

	query := url.Values{}
	if action != "" {
		query.Set("action", action)
	} else {
		query.Set("action", "open")
	}
	for k, v := range params {
		query.Set(k, v)
	}
	u.RawQuery = query.Encode()

	return u.String()
}

// Send scheme of a share
//
//	category required, category of shared content
//	data required, shared content
//	conversationID optional, If you specify conversation and it is the conversation of the user's current session, the confirmation box shown above will appear, the message will be sent after the user clicks the confirmation; if the conversation is not specified or is not the conversation of the current session, an interface where the user chooses which session to share with will show up.
//
// https://developers.mixin.one/docs/schema#sharing
func (urlScheme) Send(category SendSchemeCategory, data []byte, conversationID string) string {
	u := url.URL{
		Scheme: Scheme,
		Host:   "send",
	}
	query := url.Values{}
	query.Set("category", category)
	if len(data) > 0 {
		query.Set("data", url.QueryEscape(base64.StdEncoding.EncodeToString(data)))
	}
	if conversationID != "" {
		query.Set("conversation", conversationID)
	}
	u.RawQuery = query.Encode()

	return u.String()
}
