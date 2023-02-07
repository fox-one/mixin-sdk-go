package mixin

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/shopspring/decimal"
)

type Address struct {
	AddressID   string          `json:"address_id,omitempty"`
	AssetID     string          `json:"asset_id"`
	Label       string          `json:"label,omitempty"`
	Destination string          `json:"destination,omitempty"`
	Tag         string          `json:"tag,omitempty"`
	Fee         decimal.Decimal `json:"fee,omitempty"`
	Dust        decimal.Decimal `json:"dust,omitempty"`
}

type CreateAddressInput struct {
	AssetID     string `json:"asset_id"`
	Destination string `json:"destination,omitempty"`
	Tag         string `json:"tag,omitempty"`
	Label       string `json:"label,omitempty"`
}

func (c *Client) CreateAddress(ctx context.Context, input CreateAddressInput, pin string) (*Address, error) {
	var body interface{}
	if len(pin) == 6 {
		body = struct {
			CreateAddressInput
			Pin string `json:"pin,omitempty"`
		}{
			CreateAddressInput: input,
			Pin:                c.EncryptPin(pin),
		}
	} else {
		key, err := KeyFromString(pin)
		if err != nil {
			return nil, err
		}
		hash := sha256.New()
		hash.Write([]byte(fmt.Sprintf("%s%s%s%s%s",
			TIPAddressAdd,
			input.AssetID, input.Destination,
			input.Tag,
			input.Label)))
		tipBody := hash.Sum(nil)
		pin = key.Sign(tipBody).String()

		body = struct {
			CreateAddressInput
			Pin string `json:"pin_base64,omitempty"`
		}{
			CreateAddressInput: input,
			Pin:                c.EncryptPin(pin),
		}
	}

	var address Address
	if err := c.Post(ctx, "/addresses", body, &address); err != nil {
		return nil, err
	}

	return &address, nil
}

func (c *Client) ReadAddress(ctx context.Context, addressID string) (*Address, error) {
	uri := fmt.Sprintf("/addresses/%s", addressID)

	var address Address
	if err := c.Get(ctx, uri, nil, &address); err != nil {
		return nil, err
	}

	return &address, nil
}

func ReadAddress(ctx context.Context, accessToken, addressID string) (*Address, error) {
	return NewFromAccessToken(accessToken).ReadAddress(ctx, addressID)
}

func (c *Client) ReadAddresses(ctx context.Context, assetID string) ([]*Address, error) {
	uri := fmt.Sprintf("/assets/%s/addresses", assetID)

	var addresses []*Address
	if err := c.Get(ctx, uri, nil, &addresses); err != nil {
		return nil, err
	}

	return addresses, nil
}

func ReadAddresses(ctx context.Context, accessToken, assetID string) ([]*Address, error) {
	return NewFromAccessToken(accessToken).ReadAddresses(ctx, assetID)
}

func (c *Client) DeleteAddress(ctx context.Context, addressID, pin string) error {
	body := map[string]interface{}{}
	if len(pin) == 6 {
		body["pin"] = c.EncryptPin(pin)
	} else {
		key, err := KeyFromString(pin)
		if err != nil {
			return err
		}
		tipBody := []byte(fmt.Sprintf("%s%s", TIPAddressRemove, addressID))
		hash := sha256.New()
		hash.Write(tipBody)
		pin = key.Sign(hash.Sum(nil)).String()
		body["pin_base64"] = c.EncryptPin(pin)
	}

	uri := fmt.Sprintf("/addresses/%s/delete", addressID)
	return c.Post(ctx, uri, body, nil)
}
