// copy from github.com/MixinNetwork/nfo/nft

package nft

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

const (
	Prefix  = "NFO"
	Version = 0x00
)

var (
	MintAssetId        = "c94ac88f-4671-3976-b60a-09064f1811e8"
	MintMinimumCost, _ = decimal.NewFromString("0.001")

	GroupMembers = []string{
		"4b188942-9fb0-4b99-b4be-e741a06d1ebf",
		"dd655520-c919-4349-822f-af92fabdbdf4",
		"047061e6-496d-4c35-b06b-b0424a8a400d",
		"acf65344-c778-41ee-bacb-eb546bacfb9f",
		"a51006d0-146b-4b32-a2ce-7defbf0d7735",
		"cf4abd9c-2cfa-4b5a-b1bd-e2b61a83fabd",
		"50115496-7247-4e2c-857b-ec8680756bee",
	}
	GroupThreshold uint8 = 5

	DefaultCollectionID = uuid.Nil.String()
	DefaultChain, _     = uuid.FromString("43d61dcd-e413-450d-80b8-101d5e903357")
	DefaultClass, _     = hex.DecodeString("3c8c161a18ae2c8b14fda1216fff7da88c419b5d")
)

type NFOMemo struct {
	Prefix  string
	Version byte

	Mask       uint64
	Chain      uuid.UUID // 16 bytes
	Class      []byte    // 64 bytes contract address
	Collection uuid.UUID // 16 bytes
	Token      []byte    // 64 bytes hash of content

	Extra []byte
}

func tokenBytes(token int64) []byte {
	return big.NewInt(token).Bytes()
}

func BuildMintNFO(collection string, token int64, metaHash []byte) []byte {
	gid := uuid.FromStringOrNil(collection)
	nfo := NFOMemo{
		Prefix:     Prefix,
		Version:    Version,
		Chain:      DefaultChain,
		Class:      DefaultClass,
		Collection: gid,
		Token:      tokenBytes(token),
		Extra:      metaHash,
	}
	nfo.Mark([]int{0})
	return nfo.Encode()
}

func BuildTokenID(collection string, token int64) []byte {
	b := bytes.NewBuffer(make([]byte, 0, 16+20+16+8))
	b.Write(DefaultChain.Bytes())
	b.Write(DefaultClass)
	b.Write(uuid.FromStringOrNil(collection).Bytes())
	b.Write(tokenBytes(token))

	return b.Bytes()
}

func (nm *NFOMemo) Mark(indexes []int) {
	for _, i := range indexes {
		if i >= 64 || i < 0 {
			panic(fmt.Errorf("invalid NFO memo index %d", i))
		}
		nm.Mask ^= 1 << uint64(i)
	}
}

func (nm *NFOMemo) Indexes() []int {
	keys := make([]int, 0)
	for i := uint64(0); i < 64; i++ {
		mask := uint64(1) << i
		if nm.Mask&mask == mask {
			keys = append(keys, int(i))
		}
	}
	return keys
}

func (nm *NFOMemo) WillMint() bool {
	return nm.Mask != 0
}

func (nm *NFOMemo) Encode() []byte {
	nw := new(nfoWriter)
	nw.write([]byte(nm.Prefix))
	nw.writeByte(nm.Version)
	if nm.Mask != 0 {
		nw.writeByte(1)
		nw.writeUint64(nm.Mask)
		nw.writeUUID(nm.Chain)
		nw.writeSlice(nm.Class)
		nw.writeSlice(nm.Collection.Bytes())
		nw.writeSlice(nm.Token)
		st := tokenBytesStrip(nm.Token)
		if bytes.Compare(nm.Token, st) != 0 {
			panic(hex.EncodeToString(nm.Token))
		}
	} else {
		nw.writeByte(0)
	}
	nw.writeSlice(nm.Extra)
	return nw.Bytes()
}

func DecodeNFOMemo(b []byte) (*NFOMemo, error) {
	if len(b) < 4 {
		return nil, fmt.Errorf("NFO length %d", len(b))
	}
	if string(b[:3]) != Prefix {
		return nil, fmt.Errorf("NFO prefix %v", b[:3])
	}
	if b[3] != Version {
		return nil, fmt.Errorf("NFO version %v", b[3])
	}
	nr := &nfoReader{*bytes.NewReader(b[4:])}
	nm := &NFOMemo{
		Prefix:  Prefix,
		Version: Version,
	}

	hint, err := nr.ReadByte()
	if err != nil {
		return nil, err
	}

	if hint == 1 {
		nm.Mask, err = nr.readUint64()
		if err != nil {
			return nil, err
		}
		if nm.Mask != 1 {
			return nil, fmt.Errorf("invalid mask %v", nm.Indexes())
		}
		nm.Chain, err = nr.readUUID()
		if err != nil {
			return nil, err
		}
		if nm.Chain != DefaultChain {
			return nil, fmt.Errorf("invalid chain %s", nm.Chain.String())
		}
		nm.Class, err = nr.readBytes()
		if err != nil {
			return nil, err
		}
		if bytes.Compare(nm.Class, DefaultClass) != 0 {
			return nil, fmt.Errorf("invalid class %s", hex.EncodeToString(nm.Class))
		}
		collection, err := nr.readBytes()
		if err != nil {
			return nil, err
		}
		nm.Collection, err = uuid.FromBytes(collection)
		if err != nil {
			return nil, err
		}
		nm.Token, err = nr.readBytes()
		if err != nil {
			return nil, err
		}
		st := tokenBytesStrip(nm.Token)
		if bytes.Compare(nm.Token, st) != 0 {
			return nil, fmt.Errorf("invalid token format %s", hex.EncodeToString(nm.Token))
		}
	}
	nm.Extra, err = nr.readBytes()
	if err != nil {
		return nil, err
	}

	return nm, nil
}

type nfoReader struct{ bytes.Reader }

func (nr *nfoReader) readUint64() (uint64, error) {
	var b [8]byte
	err := nr.read(b[:])
	if err != nil {
		return 0, err
	}
	d := binary.BigEndian.Uint64(b[:])
	return d, nil
}

func (nr *nfoReader) readUint32() (uint32, error) {
	var b [4]byte
	err := nr.read(b[:])
	if err != nil {
		return 0, err
	}
	d := binary.BigEndian.Uint32(b[:])
	return d, nil
}

func (nr *nfoReader) readUUID() (uuid.UUID, error) {
	var b [16]byte
	err := nr.read(b[:])
	if err != nil {
		return uuid.Nil, err
	}
	return uuid.FromBytes(b[:])
}

func (nr *nfoReader) readBytes() ([]byte, error) {
	l, err := nr.ReadByte()
	if err != nil {
		return nil, err
	}
	if l == 0 {
		return nil, nil
	}
	b := make([]byte, l)
	err = nr.read(b)
	return b, err
}

func (nr *nfoReader) read(b []byte) error {
	l, err := nr.Read(b)
	if err != nil {
		return err
	}
	if l != len(b) {
		return fmt.Errorf("data short %d %d", l, len(b))
	}
	return nil
}

type nfoWriter struct{ bytes.Buffer }

func (nw *nfoWriter) writeUUID(u uuid.UUID) {
	nw.write(u.Bytes())
}

func (nw *nfoWriter) writeSlice(b []byte) {
	l := len(b)
	if l >= 128 {
		panic(l)
	}
	nw.writeByte(byte(l))
	nw.write(b)
}

func (nw *nfoWriter) writeByte(b byte) {
	err := nw.WriteByte(b)
	if err != nil {
		panic(err)
	}
}

func (nw *nfoWriter) write(b []byte) {
	l, err := nw.Write(b)
	if err != nil {
		panic(err)
	}
	if l != len(b) {
		panic(b)
	}
}

func (nw *nfoWriter) writeUint64(d uint64) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, d)
	nw.write(b)
}

func (nw *nfoWriter) writeUint32(d uint32) {
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, d)
	nw.write(b)
}

func tokenBytesStrip(b []byte) []byte {
	b = new(big.Int).SetBytes(b).Bytes()
	if len(b) == 0 {
		return []byte{0}
	}
	return b
}
