package mixin

import (
	"encoding/base64"

	"github.com/fox-one/mixin-sdk-go/nft"
	"golang.org/x/crypto/sha3"
)

func MetaHash(content []byte) []byte {
	b := sha3.Sum256(content)
	return b[:]
}

func BuildMintCollectibleMemo(collectionID string, token int64, metaHash []byte) string {
	b := nft.BuildMintNFO(collectionID, token, metaHash)
	return base64.RawURLEncoding.EncodeToString(b)
}

func GenerateCollectibleTokenID(collectionID string, token int64) string {
	b := nft.BuildTokenID(collectionID, token)
	return uuidHash(b)
}

func NewMintCollectibleTransferInput(traceID, collectionID string, token int64, metaHash []byte) TransferInput {
	input := TransferInput{
		AssetID: nft.MintAssetId,
		Amount:  nft.MintMinimumCost,
		TraceID: traceID,
		Memo:    BuildMintCollectibleMemo(collectionID, token, metaHash),
	}

	input.OpponentMultisig.Receivers = nft.GroupMembers
	input.OpponentMultisig.Threshold = nft.GroupThreshold
	return input
}
