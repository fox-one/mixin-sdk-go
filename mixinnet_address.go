package mixin

import (
	"bytes"
	"context"
	"crypto/rand"
	"errors"
	"io"
	"strconv"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/shopspring/decimal"
)

const MainNetworkID = "XIN"

type (
	MixinnetAddress struct {
		PrivateSpendKey Key `json:"private_spend_key"`
		PrivateViewKey  Key `json:"private_view_key"`
		PublicSpendKey  Key `json:"public_spend_key"`
		PublicViewKey   Key `json:"public_view_key"`
	}
)

func NewMixinnetAddress(rand io.Reader, public ...bool) *MixinnetAddress {
	var a = MixinnetAddress{
		PrivateSpendKey: NewKey(rand),
		PrivateViewKey:  NewKey(rand),
	}

	if len(public) > 0 && public[0] {
		a.PrivateViewKey = a.PublicSpendKey.DeterministicHashDerive()
	}
	a.PublicSpendKey = a.PrivateSpendKey.Public()
	a.PublicViewKey = a.PrivateViewKey.Public()

	return &a
}

func NewMixinnetAddressFromString(s string) (MixinnetAddress, error) {
	var a MixinnetAddress
	if !strings.HasPrefix(s, MainNetworkID) {
		return a, errors.New("invalid address network")
	}
	data := base58.Decode(s[len(MainNetworkID):])
	if len(data) != 68 {
		return a, errors.New("invalid address format")
	}
	checksum := NewHash(append([]byte(MainNetworkID), data[:64]...))
	if !bytes.Equal(checksum[:4], data[64:]) {
		return a, errors.New("invalid address checksum")
	}
	copy(a.PublicSpendKey[:], data[:32])
	copy(a.PublicViewKey[:], data[32:])
	return a, nil
}

func NewMixinnetAddressFromPublicSpend(publicSpend Key) *MixinnetAddress {
	var a = MixinnetAddress{
		PublicSpendKey: publicSpend,
	}
	a.PrivateViewKey = publicSpend.DeterministicHashDerive()
	a.PublicViewKey = a.PrivateViewKey.Public()

	return &a
}

func (a MixinnetAddress) String() string {
	data := append([]byte(MainNetworkID), a.PublicSpendKey[:]...)
	data = append(data, a.PublicViewKey[:]...)
	checksum := NewHash(data)
	data = append(a.PublicSpendKey[:], a.PublicViewKey[:]...)
	data = append(data, checksum[:4]...)
	return MainNetworkID + base58.Encode(data)
}

func (a MixinnetAddress) Hash() Hash {
	return NewHash(append(a.PublicSpendKey[:], a.PublicViewKey[:]...))
}

func (a MixinnetAddress) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(a.String())), nil
}

func (a *MixinnetAddress) UnmarshalJSON(b []byte) error {
	unquoted, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	m, err := NewMixinnetAddressFromString(unquoted)
	if err != nil {
		return err
	}
	a.PrivateSpendKey = m.PrivateSpendKey
	a.PrivateViewKey = m.PrivateViewKey
	a.PublicSpendKey = m.PublicSpendKey
	a.PublicViewKey = m.PublicViewKey
	return nil
}

func (a MixinnetAddress) CreateUTXO(outputIndex int, amount decimal.Decimal) *Output {
	r := NewKey(rand.Reader)
	pubGhost := DeriveGhostPublicKey(&r, &a.PublicViewKey, &a.PublicSpendKey, outputIndex)
	return &Output{
		Type:   0,
		Script: NewThresholdScript(1),
		Amount: NewIntegerFromDecimal(amount),
		Mask:   r.Public(),
		Keys:   []Key{*pubGhost},
	}
}

// 检查 transaction 是否是由该主网地址签发。满足以下所有条件则返回  true:
//	1. 所有 input 对应的 utxo 只有一个 keys， 即 不是多签地址 转出
//	2. 该 input 的 mask & keys 可以使用该地址的 private view 和 public spend 碰撞通过
func VerifyTransaction(ctx context.Context, addr *MixinnetAddress, txHash Hash) (bool, error) {
	if !addr.PrivateViewKey.HasValue() || !addr.PublicSpendKey.HasValue() {
		return false, errors.New("invalid address: must contains both private view key and public spend key")
	}

	tx, err := GetTransaction(ctx, txHash)
	if err != nil || !tx.Asset.HasValue() {
		return false, err
	}

	for _, input := range tx.Inputs {
		preTx, err := GetTransaction(ctx, *input.Hash)
		if err != nil || !preTx.Asset.HasValue() {
			return false, err
		}

		if input.Index >= len(preTx.Outputs) {
			return false, err
		}

		output := preTx.Outputs[input.Index]
		if len(output.Keys) != 1 {
			return false, nil
		}
		k := ViewGhostOutputKey(&output.Keys[0], &addr.PrivateViewKey, &output.Mask, input.Index)
		if !bytes.Equal(k[:], addr.PublicSpendKey[:]) {
			return false, nil
		}
	}

	return true, nil
}
