package mixin

import (
	"encoding/base64"

	"github.com/fox-one/mixin-sdk-go/nft"
)

func BuildMintCollectibleMemo(collectionID string, token int64, content []byte) string {
	b := nft.BuildMintNFO(collectionID, token, NewHash(content))
	return base64.RawURLEncoding.EncodeToString(b)
}

func GenerateCollectibleTokenID(collectionID string, token int64) string {
	b := nft.BuildTokenID(collectionID, token)
	return UUIDFromBytes(b)
}

func NewMintCollectibleTransferInput(traceID, collectionID string, token int64, content []byte) TransferInput {
	input := TransferInput{
		AssetID: nft.MintAssetId,
		Amount:  nft.MintMinimumCost,
		TraceID: traceID,
		Memo:    BuildMintCollectibleMemo(collectionID, token, content),
	}

	input.OpponentMultisig.Receivers = nft.GroupMembers
	input.OpponentMultisig.Threshold = nft.GroupThreshold
	return input
}
