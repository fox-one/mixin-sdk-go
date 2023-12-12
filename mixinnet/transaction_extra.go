package mixinnet

import (
	"encoding/base64"
	"encoding/hex"
	"strconv"
)

type (
	TransactionExtra []byte
)

// Transaction Extra

func (e TransactionExtra) String() string {
	return base64.StdEncoding.EncodeToString(e[:])
}

func (e TransactionExtra) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(e.String())), nil
}

func (e *TransactionExtra) UnmarshalJSON(b []byte) error {
	unquoted, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	data, err := hex.DecodeString(unquoted)
	if err != nil {
		if data, err = base64.StdEncoding.DecodeString(unquoted); err != nil {
			return err
		}
	}

	*e = data
	return nil
}
