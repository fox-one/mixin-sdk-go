package mixin

import (
	"filippo.io/edwards25519"
	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
)

func SafeSignTransaction(tx *mixinnet.Transaction, spendKey mixinnet.Key, views []mixinnet.Key, k uint16) error {
	y, err := spendKey.ToScalar()
	if err != nil {
		return err
	}

	txHash, err := tx.TransactionHash()
	if err != nil {
		return err
	}

	tx.Signatures = make([]map[uint16]*mixinnet.Signature, len(tx.Inputs))

	for idx, view := range views {
		x, err := view.ToScalar()
		if err != nil {
			panic(err)
		}
		t := edwards25519.NewScalar().Add(x, y)
		var key mixinnet.Key
		copy(key[:], t.Bytes())
		sig := key.SignHash(txHash)
		tx.Signatures[idx] = map[uint16]*mixinnet.Signature{
			k: &sig,
		}
	}

	return nil
}
