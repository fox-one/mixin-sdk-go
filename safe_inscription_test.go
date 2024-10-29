package mixin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadSafeCollection(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)

	testCollectionHash := "b3979998b8b5e705d553288bffd96d4e1cc719f3ae0b01ecac8539e1df81c16f"

	collection, err := ReadSafeCollection(ctx, testCollectionHash)
	require.NoError(err, "ReadSafeCollection")
	require.NotNil(collection)
	t.Log(collection)
}
func TestReadSafeCollectible(t *testing.T) {
	ctx := context.Background()
	require := require.New(t)

	testInscriptionHash := "94d20f04829dcfb2c6d3cdb7ba94b3f6b402eb0537d6aa48f76e14d21e84c784"

	inscription, err := ReadSafeCollectible(ctx, testInscriptionHash)

	require.NoError(err, "ReadSafeInscription")
	require.NotNil(inscription)
	t.Log(inscription)
}
