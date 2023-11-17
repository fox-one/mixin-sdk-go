package mixin

import (
	"bytes"
	"encoding/binary"
	"sort"
)

const (
	MinimumEncodingVersion = 0x1
	MaximumEncodingInt     = 0xFFFF

	AggregatedSignaturePrefix      = 0xFF01
	AggregatedSignatureSparseMask  = byte(0x01)
	AggregatedSignatureOrdinayMask = byte(0x00)
)

var (
	magic = []byte{0x77, 0x77}
	null  = []byte{0x00, 0x00}
)

type Encoder struct {
	buf *bytes.Buffer
}

func NewEncoder() *Encoder {
	return &Encoder{buf: new(bytes.Buffer)}
}

func NewMinimumEncoder() *Encoder {
	enc := NewEncoder()
	enc.Write(magic)
	enc.Write([]byte{0x00, MinimumEncodingVersion})
	return enc
}

func (enc *Encoder) EncodeTransaction(signed *Transaction) []byte {
	switch signed.Version {
	case TxVersionCommonEncoding,
		TxVersionBlake3Hash,
		TxVersionReferences:
	default:
		panic(signed)
	}

	enc.Write(magic)
	enc.Write([]byte{0x00, signed.Version})
	enc.Write(signed.Asset[:])

	il := len(signed.Inputs)
	enc.WriteInt(il)
	for _, in := range signed.Inputs {
		enc.EncodeInput(in)
	}

	ol := len(signed.Outputs)
	enc.WriteInt(ol)
	for _, out := range signed.Outputs {
		enc.EncodeOutput(out, signed.Version)
	}

	if signed.Version >= TxVersionReferences {
		rl := len(signed.References)
		enc.WriteInt(rl)
		for _, r := range signed.References {
			enc.Write(r[:])
		}

		el := len(signed.Extra)
		if el > ExtraSizeStorageCapacity {
			panic(el)
		}
		enc.WriteUint32(uint32(el))
		enc.Write(signed.Extra)
	} else {
		el := len(signed.Extra)
		enc.WriteInt(el)
		enc.Write(signed.Extra)
	}

	if signed.AggregatedSignature != nil {
		enc.EncodeAggregatedSignature(signed.AggregatedSignature)
	} else {
		sl := len(signed.Signatures)
		if sl == MaximumEncodingInt {
			panic(sl)
		}
		enc.WriteInt(sl)
		for _, sm := range signed.Signatures {
			enc.EncodeSignatures(sm)
		}
	}

	return enc.buf.Bytes()
}

func (enc *Encoder) EncodeInput(in *Input) {
	enc.Write(in.Hash[:])
	enc.WriteUint16(uint16(in.Index))

	enc.WriteInt(len(in.Genesis))
	enc.Write(in.Genesis)

	if d := in.Deposit; d == nil {
		enc.Write(null)
	} else {
		enc.Write(magic)
		enc.Write(d.Chain[:])

		enc.WriteInt(len(d.AssetKey))
		enc.Write([]byte(d.AssetKey))

		enc.WriteInt(len(d.Transaction))
		enc.Write([]byte(d.Transaction))

		enc.WriteUint64(d.Index)
		enc.WriteInteger(d.Amount)
	}

	if m := in.Mint; m == nil {
		enc.Write(null)
	} else {
		enc.Write(magic)

		enc.WriteInt(len(m.Group))
		enc.Write([]byte(m.Group))

		enc.WriteUint64(m.Batch)
		enc.WriteInteger(m.Amount)
	}
}

func (enc *Encoder) EncodeOutput(o *Output, ver uint8) {
	enc.Write([]byte{0x00, o.Type})
	enc.WriteInteger(o.Amount)
	enc.WriteInt(len(o.Keys))
	for _, k := range o.Keys {
		enc.Write(k[:])
	}

	enc.Write(o.Mask[:])
	enc.WriteInt(len(o.Script))
	enc.Write(o.Script)

	if w := o.Withdrawal; w == nil {
		enc.Write(null)
	} else {
		if ver < TxVersionHashSignature {
			enc.Write(magic)
			enc.Write(w.Chain[:])

			enc.WriteInt(len(w.AssetKey))
			enc.Write([]byte(w.AssetKey))

			enc.WriteInt(len(w.Address))
			enc.Write([]byte(w.Address))

			enc.WriteInt(len(w.Tag))
			enc.Write([]byte(w.Tag))
		} else {
			enc.Write(magic)

			enc.WriteInt(len(w.Address))
			enc.Write([]byte(w.Address))

			enc.WriteInt(len(w.Tag))
			enc.Write([]byte(w.Tag))
		}
	}
}

func (enc *Encoder) EncodeSignatures(sm map[uint16]*Signature) {
	ss, off := make([]struct {
		Index uint16
		Sig   *Signature
	}, len(sm)), 0
	for j, sig := range sm {
		ss[off].Index = j
		ss[off].Sig = sig
		off += 1
	}
	sort.Slice(ss, func(i, j int) bool { return ss[i].Index < ss[j].Index })

	enc.WriteInt(len(ss))
	for _, sp := range ss {
		enc.WriteUint16(sp.Index)
		enc.Write(sp.Sig[:])
	}
}

func (enc *Encoder) Write(b []byte) {
	l, err := enc.buf.Write(b)
	if err != nil {
		panic(err)
	}
	if l != len(b) {
		panic(b)
	}
}

func (enc *Encoder) WriteByte(b byte) error {
	err := enc.buf.WriteByte(b)
	if err != nil {
		panic(err)
	}
	return nil
}

func (enc *Encoder) WriteInt(d int) {
	if d > MaximumEncodingInt {
		panic(d)
	}
	b := uint16ToByte(uint16(d))
	enc.Write(b)
}

func (enc *Encoder) WriteUint16(d uint16) {
	if d > MaximumEncodingInt {
		panic(d)
	}
	b := uint16ToByte(d)
	enc.Write(b)
}

func (enc *Encoder) WriteUint32(d uint32) {
	b := uint32ToBytes(d)
	enc.Write(b)
}

func (enc *Encoder) WriteUint64(d uint64) {
	b := uint64ToByte(d)
	enc.Write(b)
}

func (enc *Encoder) WriteInteger(d Integer) {
	b := d.i.Bytes()
	enc.WriteInt(len(b))
	enc.Write(b)
}

func uint16ToByte(d uint16) []byte {
	b := make([]byte, 2)
	binary.BigEndian.PutUint16(b, d)
	return b
}

func uint32ToBytes(d uint32) []byte {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, d)
	return b
}

func uint64ToByte(d uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, d)
	return b
}

func (enc *Encoder) EncodeAggregatedSignature(js *AggregatedSignature) {
	enc.WriteInt(MaximumEncodingInt)
	enc.WriteInt(AggregatedSignaturePrefix)
	enc.Write(js.Signature[:])
	if len(js.Signers) == 0 {
		enc.WriteByte(AggregatedSignatureOrdinayMask)
		enc.WriteInt(0)
		return
	}
	for i, m := range js.Signers {
		if i > 0 && m <= js.Signers[i-1] {
			panic(js.Signers)
		}
		if m > MaximumEncodingInt {
			panic(js.Signers)
		}
	}

	max := js.Signers[len(js.Signers)-1]
	if max/8+1 > len(js.Signers)*2 {
		enc.WriteByte(AggregatedSignatureSparseMask)
		enc.WriteInt(len(js.Signers))
		for _, m := range js.Signers {
			enc.WriteInt(m)
		}
		return
	}

	masks := make([]byte, max/8+1)
	for _, m := range js.Signers {
		masks[m/8] = masks[m/8] ^ (1 << (m % 8))
	}
	enc.WriteByte(AggregatedSignatureOrdinayMask)
	enc.WriteInt(len(masks))
	enc.Write(masks)
}
