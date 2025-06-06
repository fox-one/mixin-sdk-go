package mixin

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/btcsuite/btcutil/base58"
	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
	"github.com/gofrs/uuid/v5"
)

const (
	MixAddressPrefix  = "MIX"
	MixAddressVersion = byte(2)
)

type (
	MixAddress struct {
		Version     byte
		Threshold   byte
		uuidMembers []uuid.UUID
		xinMembers  []*mixinnet.Address
	}
)

func RequireNewMixAddress(members []string, threshold byte) *MixAddress {
	addr, err := NewMixAddress(members, threshold)
	if err != nil {
		panic(err)
	}
	return addr
}

func RequireNewMainnetMixAddress(members []string, threshold byte) *MixAddress {
	addr, err := NewMainnetMixAddress(members, threshold)
	if err != nil {
		panic(err)
	}
	return addr
}

func NewMixAddress(members []string, threshold byte) (*MixAddress, error) {
	if len(members) > 255 {
		return nil, fmt.Errorf("invlaid member count: %d", len(members))
	}
	if int(threshold) == 0 || int(threshold) > len(members) {
		return nil, fmt.Errorf("invlaid members threshold: %d", threshold)
	}
	ma := &MixAddress{
		Version:   MixAddressVersion,
		Threshold: threshold,
	}
	for _, s := range members {
		u, err := uuid.FromString(s)
		if err != nil {
			return nil, fmt.Errorf("invalid member (%v): %v", s, err)
		}
		ma.uuidMembers = append(ma.uuidMembers, u)
	}
	return ma, nil
}

func NewMainnetMixAddress(members []string, threshold byte) (*MixAddress, error) {
	if len(members) > 255 {
		return nil, fmt.Errorf("invlaid member count: %d", len(members))
	}
	if int(threshold) == 0 || int(threshold) > len(members) {
		return nil, fmt.Errorf("invlaid members threshold: %d", threshold)
	}
	ma := &MixAddress{
		Version:   MixAddressVersion,
		Threshold: threshold,
	}
	for _, s := range members {
		a, err := mixinnet.AddressFromString(s)
		if err != nil {
			return nil, fmt.Errorf("invalid member (%v): %v", s, err)
		}
		ma.xinMembers = append(ma.xinMembers, &a)
	}
	return ma, nil
}

func (ma *MixAddress) Members() []string {
	var members []string
	if len(ma.uuidMembers) > 0 {
		for _, u := range ma.uuidMembers {
			members = append(members, u.String())
		}
	} else {
		for _, a := range ma.xinMembers {
			members = append(members, a.String())
		}
	}
	return members
}

func (ma *MixAddress) String() string {
	payload := []byte{ma.Version, ma.Threshold}
	if l := len(ma.uuidMembers); l > 0 {
		if l > 255 {
			panic(l)
		}
		payload = append(payload, byte(l))
		for _, u := range ma.uuidMembers {
			payload = append(payload, u.Bytes()...)
		}
	} else {
		l := len(ma.xinMembers)
		if l > 255 {
			panic(l)
		}
		payload = append(payload, byte(l))
		for _, a := range ma.xinMembers {
			payload = append(payload, a.PublicSpendKey[:]...)
			payload = append(payload, a.PublicViewKey[:]...)
		}
	}

	data := append([]byte(MixAddressPrefix), payload...)
	checksum := mixinnet.NewHash(data)
	payload = append(payload, checksum[:4]...)
	return MixAddressPrefix + base58.Encode(payload)
}

func MixAddressFromString(s string) (*MixAddress, error) {
	var ma MixAddress
	if !strings.HasPrefix(s, MixAddressPrefix) {
		return nil, fmt.Errorf("invalid address prefix %s", s)
	}
	data := base58.Decode(s[len(MixAddressPrefix):])
	if len(data) < 3+16+4 {
		return nil, fmt.Errorf("invalid address length %d", len(data))
	}
	payload := data[:len(data)-4]
	checksum := mixinnet.NewHash(append([]byte(MixAddressPrefix), payload...))
	if !bytes.Equal(checksum[:4], data[len(data)-4:]) {
		return nil, fmt.Errorf("invalid address checksum %x", checksum[:4])
	}

	total := payload[2]
	ma.Version = payload[0]
	ma.Threshold = payload[1]
	if ma.Version != MixAddressVersion {
		return nil, fmt.Errorf("invalid address version %d", ma.Version)
	}
	if ma.Threshold == 0 || ma.Threshold > total || total > 64 {
		return nil, fmt.Errorf("invalid address threshold %d/%d", ma.Threshold, total)
	}

	mp := payload[3:]
	if len(mp) == 16*int(total) {
		for i := 0; i < int(total); i++ {
			id, err := uuid.FromBytes(mp[i*16 : i*16+16])
			if err != nil {
				return nil, fmt.Errorf("invalid uuid member %s", s)
			}
			ma.uuidMembers = append(ma.uuidMembers, id)
		}
	} else if len(mp) == 64*int(total) {
		for i := 0; i < int(total); i++ {
			var a mixinnet.Address
			copy(a.PublicSpendKey[:], mp[i*64:i*64+32])
			copy(a.PublicViewKey[:], mp[i*64+32:i*64+64])
			ma.xinMembers = append(ma.xinMembers, &a)
		}
	} else {
		return nil, fmt.Errorf("invalid address members list %s", s)
	}

	return &ma, nil
}
