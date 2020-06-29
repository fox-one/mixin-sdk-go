package mixin

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func newKeystoreFromTestData(t *testing.T) *Keystore {
	path := "./testdata/keystore.json"
	f, err := os.Open(path)
	if err != nil {
		t.Errorf("os.Open(%v): %v", path, err)
		t.FailNow()
	}

	defer f.Close()

	var store Keystore
	if err := json.NewDecoder(f).Decode(&store); err != nil {
		t.Errorf("json decode failed: %v", err)
		t.FailNow()
	}

	return &store
}

func TestKeystoreAuth(t *testing.T) {
	s := newKeystoreFromTestData(t)

	auth, err := AuthFromKeystore(s)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	sig := SignRaw("GET", "/me", nil)
	token := auth.SignToken(sig, newRequestID(), time.Minute)

	me, err := UserMe(context.TODO(), token)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	assert.Equal(t, s.ClientID, me.UserID, "client id should be same")
}
