package mixinnet

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRPC(t *testing.T) {
	ctx := context.Background()

	t.Run("legacy-network", func(t *testing.T) {
		client := NewClient(DefaultLegacyConfig)

		info, err := client.ReadConsensusInfo(ctx)
		require.Nil(t, err, "ReadConsensusInfo")
		require.Greater(t, info.Graph.Topology, uint64(0))

		hash, err := HashFromString("abadfce0377eae4e0057289ddcd067068171f034814c476a0d2c4a7807f222a7")
		require.Nil(t, err)

		tx, err := client.GetTransaction(ctx, hash)
		require.Nil(t, err)

		tx.Hash = nil
		hash1, err := tx.TransactionHash()
		require.Nil(t, err)
		require.Equal(t, hash[:], hash1[:], "hash not matched: %v != %v", hash, hash1)

		tx1, err := client.SendRawTransaction(ctx, "77770003a99c2e0e2b1da4d648755ef19bd95139acbbe6564cfb06dec7cd34931ca72cdc000100b5127275c76409d54e8e56b17308ee1e7686fbdd624ec31beb5f897adba6e80000000000000000000100a300060138e6ae9f0000000000000000000000000000000000000000000000000000000000000000000000000000000040d1da563f64423cdd7544d9dadb0ffc5af1c403c7f900c0d9c4e627c6d03ca8bfb27b4a84b6627c77a57074db780d16968e902614a99f02dbc5267ddfaaf38935ffffff017cd3a768b3c56c35ff0b1c3ba88ff389d21fc8ed491d2962288c02efa17a29b730a6f3f06cbd7b96f5f512f639f837941f4da22bedd7e3ce3933a8a520b3f40d00000101")
		require.Nil(t, err)
		hash2, err := tx1.TransactionHash()
		require.Nil(t, err)
		require.Equal(t, hash[:], hash2[:], "hash not matched: %v != %v", hash, hash2)

		utxo, err := client.GetUTXO(ctx, *tx.Inputs[0].Hash, tx.Inputs[0].Index)
		require.Nil(t, err)
		require.Equal(t, hash[:], utxo.Lock[:])
		require.Equal(t, tx.Inputs[0].Hash[:], utxo.Hash[:])
		require.Equal(t, tx.Inputs[0].Index, utxo.Index)
	})

	t.Run("safe-network", func(t *testing.T) {
		client := NewClient(DefaultSafeConfig)

		info, err := client.ReadConsensusInfo(ctx)
		require.Nil(t, err, "ReadConsensusInfo")
		require.Greater(t, info.Graph.Topology, uint64(0))

		hash, err := HashFromString("97734eaedd70ab91a23f84dbef398538d7829da51a43cfcb8df81d4f010d7688")
		require.Nil(t, err)

		tx, err := client.GetTransaction(ctx, hash)
		require.Nil(t, err)

		hash1, err := tx.TransactionHash()
		require.Nil(t, err)
		require.Equal(t, hash[:], hash1[:], "hash not matched: %v != %v", hash, hash1)

		tx1, err := client.SendRawTransaction(ctx, "77770005a99c2e0e2b1da4d648755ef19bd95139acbbe6564cfb06dec7cd34931ca72cdc00015365637a68a57c8f2ef391f760a0ac262af95c52edc305f9201fb55997aabcaa0000000000000000000100a300060138e6ae9f000000000000000000000000000000000000000000000000000000000000000000000000000000000000000040b77f9f5f35c70cdfb7800e9d3ffcc48d4b2f069806054b0f57d004633d53f3aafafd4e3b1d1c91330d22bc4a0b57594d88bab7f132ffe9417bad92c8ba3b7bcbffffff019ea442f3e17c2def22aa0c742270d07dc1f866613dfb57175263ddf975e6b33187acaa9918dbf1867c5bb3c5fda5a7c8319b0dd4dc08a46cb9161d304f689f0400000101")
		require.Nil(t, err)
		hash2, err := tx1.TransactionHash()
		require.Nil(t, err)
		require.Equal(t, hash[:], hash2[:], "hash not matched: %v != %v", hash, hash2)

		utxo, err := client.GetUTXO(ctx, *tx.Inputs[0].Hash, tx.Inputs[0].Index)
		require.Nil(t, err)
		require.Equal(t, hash[:], utxo.Lock[:])
	})
}
