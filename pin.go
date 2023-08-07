package mixin

import (
	"context"
	"crypto/ed25519"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"
)

func (c *Client) VerifyPin(ctx context.Context, pin string) error {
	body := map[string]interface{}{}
	if len(pin) > 6 {
		timestamp := uint64(time.Now().UnixNano())
		tipBody := []byte(fmt.Sprintf("%s%032d", TIPVerify, timestamp))
		body["timestamp"] = timestamp

		if pinBts, err := hex.DecodeString(pin); err == nil {
			switch len(pinBts) {
			case ed25519.PrivateKeySize:
				body["pin_base64"] = c.EncryptPin(hex.EncodeToString(ed25519.Sign(pinBts, tipBody)))

			case 32:
				var key Key
				copy(key[:], pinBts)
				body["pin_base64"] = c.EncryptPin(key.Sign(tipBody).String())
			}
		}

	}

	if _, ok := body["pin_base64"]; !ok {
		body["pin"] = c.EncryptPin(pin)
	}

	return c.Post(ctx, "/pin/verify", body, nil)
}

func (c *Client) ModifyPin(ctx context.Context, pin, newPin string) error {
	body := map[string]interface{}{}

	if pin != "" {
		body["old_pin"] = c.EncryptPin(pin)
	}

	if len(newPin) > 6 {
		counter := make([]byte, 8)
		binary.BigEndian.PutUint64(counter, 1)
		newPin = newPin + hex.EncodeToString(counter)
	}

	body["pin"] = c.EncryptPin(newPin)

	return c.Post(ctx, "/pin/update", body, nil)
}

var (
	pinRegex = regexp.MustCompile(`^\d{6}$`)
)

// ValidatePinPattern validate the pin with pinRegex
func ValidatePinPattern(pin string) error {
	if len(pin) > 6 {
		if pinBts, err := hex.DecodeString(pin); err == nil && len(pinBts) == 32 {
			return nil
		}
	}
	if !pinRegex.MatchString(pin) {
		return fmt.Errorf("pin must match regex pattern %q", pinRegex.String())
	}

	return nil
}
