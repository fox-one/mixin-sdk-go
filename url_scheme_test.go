package mixin

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestUrlScheme_Transfer(t *testing.T) {
	userID := newUUID()
	url := URL.Transfer(userID)
	assert.Equal(t, "mixin://transfer/"+userID, url)
}

func TestUrlScheme_SafePay(t *testing.T) {
	input := &TransferInput{
		AssetID:    newUUID(),
		OpponentID: newUUID(),
		Amount:     decimal.NewFromInt(100),
		TraceID:    newUUID(),
		Memo:       "test",
	}

	url := URL.SafePay(input)
	t.Log(url)
}

func TestUrlScheme_Apps(t *testing.T) {
	t.Run("default action", func(t *testing.T) {
		appID := newUUID()
		action := ""
		url := URL.Apps(appID, action, nil)
		assert.Equal(t, "mixin://apps/"+appID+"?action=open", url)
	})
	t.Run("specify action", func(t *testing.T) {
		appID := newUUID()
		action := "close"
		url := URL.Apps(appID, action, nil)
		assert.Equal(t, "mixin://apps/"+appID+"?action="+action, url)
	})
	t.Run("specify params", func(t *testing.T) {
		appID := newUUID()
		action := ""
		params := map[string]string{"k1": "v1", "k2": "v2"}
		url := URL.Apps(appID, action, params)
		assert.Contains(t, url, "mixin://apps/"+appID+"?action=open")
		assert.Contains(t, url, "k1=v1")
		assert.Contains(t, url, "k2=v2")
	})
}
