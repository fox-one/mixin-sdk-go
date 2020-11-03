package mixin

import (
	"bytes"
	"context"
	"errors"
)

type (
	MixinnetAddress struct {
		PrivateSpendKey Key `json:"private_spend_key"`
		PrivateViewKey  Key `json:"private_view_key"`
		PublicSpendKey  Key `json:"public_spend_key"`
		PublicViewKey   Key `json:"public_view_key"`
	}
)

func NewPublicAddress(publicSpend Key) MixinnetAddress {
	var a = MixinnetAddress{
		PublicSpendKey: publicSpend,
	}
	a.PrivateViewKey = publicSpend.DeterministicHashDerive()
	a.PublicViewKey = a.PrivateViewKey.Public()

	return a
}

// 检查 transaction 是否是由该主网地址签发。满足以下所有条件则返回  true:
//	1. 该 transaction 只有一个 input ，且该 input 的类型为普通转账类型, 即 Hash 不为空
//	2. 该 input 对应的 utxo 只有一个 keys， 即 不是多签地址 转出
//	3. 该 input 的 mask & keys 可以使用该地址的 private view 和 public spend 碰撞通过
func VerifyTransaction(ctx context.Context, addr MixinnetAddress, txHash Hash) (bool, error) {
	if !addr.PrivateViewKey.HasValue() || !addr.PublicSpendKey.HasValue() {
		return false, errors.New("invalid address: must contains both private view key and public spend key")
	}

	tx, err := GetTransaction(ctx, txHash)
	if err != nil || tx == nil {
		return false, err
	}

	if len(tx.Inputs) > 1 || tx.Inputs[0].Index < 0 || tx.Inputs[0].Hash == nil || !tx.Inputs[0].Hash.HasValue() {
		return false, nil
	}

	preTx, err := GetTransaction(ctx, *tx.Inputs[0].Hash)
	if err != nil {
		return false, err
	}

	if tx.Inputs[0].Index >= len(preTx.Outputs) {
		return false, err
	}

	output := preTx.Outputs[tx.Inputs[0].Index]
	if len(output.Keys) == 1 {
		k := ViewGhostOutputKey(&output.Keys[0], &addr.PrivateViewKey, &output.Mask, tx.Inputs[0].Index)
		if bytes.Compare(k[:], addr.PublicSpendKey[:]) == 0 {
			return true, nil
		}
	}

	return false, nil
}
