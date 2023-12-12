package mixinnet

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewPublicMixinnetAddress(t *testing.T) {
	r := rand.Reader

	{
		a := GenerateAddress(r, true)
		b := GenerateAddress(r, true)
		if a.PrivateViewKey.String() == b.PrivateViewKey.String() {
			t.Errorf("same PrivateViewKey generated %v, %v", a.PrivateViewKey, b.PrivateViewKey)
		}
	}

	{
		pubSpend, err := KeyFromString("d03ac2718891838840c55f681b6b049af5b9efbf0d7d2a06d6741bbc17f68262")
		require.Nil(t, err)

		addr := AddressFromPublicSpend(pubSpend)
		require.NotNil(t, addr)
		require.Equal(t, "XINUF4GcHPYHUimJxvzoamCJjek6aGBqJAAwZVPGbbpSMGrvjngaZuYjnTyf4or9M7j71z4QzZS44FqvdEiT1fYZrmvSJyfj", addr.String())
	}
}
