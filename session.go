package mixin

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"sort"
)

const (
	SessionPlatformIOS     = "iOS"
	SessionPlatformAndroid = "Android"
	SessionPlatformDesktop = "Desktop"
)

type Session struct {
	UserID    string `json:"user_id,omitempty"`
	SessionID string `json:"session_id,omitempty"`
	PublicKey string `json:"public_key,omitempty"`
	Platform  string `json:"platform,omitempty"`
}

func IsEncryptedMessageSupported(sessions []*Session) bool {
	for _, session := range sessions {
		if session.PublicKey == "" {
			return false
		}
	}

	return true
}

func GenerateSessionChecksum(sessions []*Session) string {
	ids := make([]string, len(sessions))
	for i, session := range sessions {
		ids[i] = session.SessionID
	}

	if len(ids) == 0 {
		return ""
	}

	sort.Strings(ids)
	h := md5.New()
	for _, id := range ids {
		_, _ = io.WriteString(h, id)
	}
	sum := h.Sum(nil)
	return hex.EncodeToString(sum[:])
}

func (c *Client) FetchSessions(ctx context.Context, ids []string) ([]*Session, error) {
	var sessions []*Session

	if err := c.Post(ctx, "/sessions/fetch", ids, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}
