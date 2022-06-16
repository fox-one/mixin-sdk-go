package mixin

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/binary"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
)

func newUUID() string {
	return uuid.Must(uuid.NewV4()).String()
}

func UniqueConversationID(userID, recipientID string) string {
	minID, maxID := userID, recipientID
	if strings.Compare(userID, recipientID) > 0 {
		maxID, minID = userID, recipientID
	}

	return uuidHash([]byte(minID + maxID))
}

func uuidHash(b []byte) string {
	h := md5.New()
	h.Write(b)
	sum := h.Sum(nil)
	sum[6] = (sum[6] & 0x0f) | 0x30
	sum[8] = (sum[8] & 0x3f) | 0x80
	return uuid.FromBytesOrNil(sum).String()
}

func RandomPin() string {
	var b [8]byte
	_, err := rand.Read(b[:])
	if err != nil {
		panic(err)
	}
	c := binary.LittleEndian.Uint64(b[:]) % 1000000
	if c < 100000 {
		c = 100000 + c
	}

	return strconv.FormatUint(c, 10)
}

func RandomTraceID() string {
	return newUUID()
}
