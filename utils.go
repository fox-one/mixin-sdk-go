package mixin

import (
	"crypto/md5"
	"io"
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

	h := md5.New()
	_, _ = io.WriteString(h, minID+maxID)
	sum := h.Sum(nil)
	sum[6] = (sum[6] & 0x0f) | 0x30
	sum[8] = (sum[8] & 0x3f) | 0x80
	return uuid.FromBytesOrNil(sum).String()
}
