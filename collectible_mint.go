package mixin

import (
	"encoding/base64"

	"github.com/fox-one/mixin-sdk-go/nft"
	"github.com/gofrs/uuid"
)

func BuildMintCollectibleMemo(collectionID, tokenID string, content []byte) string {
	tokenUUID := uuid.FromStringOrNil(tokenID)
	nfo := nft.BuildMintNFO(collectionID, tokenUUID.Bytes(), NewHash(content))
	return base64.RawURLEncoding.EncodeToString(nfo)
}

func NewMintCollectibleTransferInput(traceID, collectionID, tokenID string, content []byte) TransferInput {
	input := TransferInput{
		AssetID: nft.MintAssetId,
		Amount:  nft.MintMinimumCost,
		TraceID: traceID,
		Memo:    BuildMintCollectibleMemo(collectionID, tokenID, content),
	}

	input.OpponentMultisig.Receivers = nft.GroupMembers
	input.OpponentMultisig.Threshold = nft.GroupThreshold
	return input
}
