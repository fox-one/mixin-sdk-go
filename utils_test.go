package mixin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenUuidFromStrings(t *testing.T) {
	uuid1 := "46a2645c-3264-4569-b462-823b5cb968e7"
	uuid2 := "bde0e466-d8d8-4bea-8bab-89a85203158c"

	ID1 := GenUuidFromStrings(uuid1, uuid2)
	ID2 := GenUuidFromStrings(uuid1, uuid2)

	if ID1 != ID2 {
		t.Errorf("expected ID1 and ID2 to be the same, got %s and %s", ID1, ID2)
	}

	ID4 := GenUuidFromStrings()
	ID5 := GenUuidFromStrings()
	assert.Equal(t, ID4, ID5)
}
