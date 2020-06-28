package mixin

import (
	"net/url"
	"path"
)

const Scheme = "mixin"

var URLScheme urlScheme

type urlScheme struct{}

func (urlScheme) User(userID string) string {
	u := url.URL{
		Scheme: Scheme,
		Path:   path.Join("/users", userID),
	}

	return u.String()
}

func (urlScheme) Transfer(userID string) string {
	u := url.URL{
		Scheme: Scheme,
		Path:   path.Join("/transfer", userID),
	}

	return u.String()
}
