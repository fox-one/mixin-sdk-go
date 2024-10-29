package mixin

import (
	"context"
	"time"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
)

type (
	SafeCollection struct {
		AssetKey       string        `json:"asset_key,omitempty"`
		CollectionHash mixinnet.Hash `json:"collection_hash,omitempty"`
		Description    string        `json:"description,omitempty"`
		IconURL        string        `json:"icon_url,omitempty"`
		KernelAssetID  mixinnet.Hash `json:"kernel_asset_id,omitempty"`
		MinimumPrice   string        `json:"minimum_price,omitempty"`
		Name           string        `json:"name,omitempty"`
		Supply         string        `json:"supply,omitempty"`
		Symbol         string        `json:"symbol,omitempty"`
		Type           string        `json:"type,omitempty"`
		Unit           string        `json:"unit,omitempty"`
		CreatedAt      time.Time     `json:"created_at,omitempty"`
		UpdatedAt      time.Time     `json:"updated_at,omitempty"`
	}

	SafeCollectible struct {
		CollectionHash  mixinnet.Hash `json:"collection_hash,omitempty"`
		ContentType     string        `json:"content_type,omitempty"`
		ContentURL      string        `json:"content_url,omitempty"`
		InscriptionHash mixinnet.Hash `json:"inscription_hash,omitempty"`
		OccupiedBy      string        `json:"occupied_by,omitempty"`
		Owner           string        `json:"owner,omitempty"`
		Recipient       string        `json:"recipient,omitempty"`
		Sequence        int64         `json:"sequence,omitempty"`
		Type            string        `json:"type,omitempty"`
		CreatedAt       time.Time     `json:"created_at,omitempty"`
		UpdatedAt       time.Time     `json:"updated_at,omitempty"`
	}
)

func ReadSafeCollection(ctx context.Context, collectionHash string) (*SafeCollection, error) {
	resp, err := Request(ctx).Get("/safe/inscriptions/collections/" + collectionHash)
	if err != nil {
		return nil, err
	}

	var collection SafeCollection
	if err := UnmarshalResponse(resp, &collection); err != nil {
		return nil, err
	}

	return &collection, nil
}

func ReadSafeCollectible(ctx context.Context, inscriptionHash string) (*SafeCollectible, error) {
	resp, err := Request(ctx).Get("/safe/inscriptions/items/" + inscriptionHash)
	if err != nil {
		return nil, err
	}

	var collectible SafeCollectible
	if err := UnmarshalResponse(resp, &collectible); err != nil {
		return nil, err
	}

	return &collectible, nil
}
