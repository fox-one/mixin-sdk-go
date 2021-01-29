package mixin

import (
	"math/big"
	"strconv"
	"strings"

	"github.com/fox-one/msgpack"
	"github.com/shopspring/decimal"
)

const Precision = 8

type (
	Integer struct {
		i big.Int
	}
)

func init() {
	msgpack.RegisterExt(0, (*Integer)(nil))
}

func NewIntegerFromDecimal(d decimal.Decimal) (v Integer) {
	if d.Sign() <= 0 {
		panic(d)
	}
	s := d.Mul(decimal.New(1, Precision)).StringFixed(0)
	v.i.SetString(s, 10)
	return
}

func NewIntegerFromString(x string) (v Integer) {
	d, err := decimal.NewFromString(x)
	if err != nil {
		panic(err)
	}
	if d.Sign() <= 0 {
		panic(x)
	}
	s := d.Mul(decimal.New(1, Precision)).StringFixed(0)
	v.i.SetString(s, 10)
	return
}

func (x Integer) String() string {
	s := x.i.String()
	p := len(s) - Precision
	if p > 0 {
		return s[:p] + "." + s[p:]
	}
	return "0." + strings.Repeat("0", -p) + s
}

func (x Integer) MarshalMsgpack() ([]byte, error) {
	return x.i.Bytes(), nil
}

func (x *Integer) UnmarshalMsgpack(data []byte) error {
	x.i.SetBytes(data)
	return nil
}

func (x Integer) MarshalJSON() ([]byte, error) {
	s := x.String()
	return []byte(strconv.Quote(s)), nil
}

func (x *Integer) UnmarshalJSON(b []byte) error {
	unquoted, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	i := NewIntegerFromString(unquoted)
	x.i.SetBytes(i.i.Bytes())
	return nil
}
