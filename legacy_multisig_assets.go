package mixin

import "context"

func ReadMultisigAssets(ctx context.Context) ([]*Asset, error) {
	resp, err := Request(ctx).Get("/network/assets/multisig")
	if err != nil {
		return nil, err
	}

	var assets []*Asset
	if err := UnmarshalResponse(resp, &assets); err != nil {
		return nil, err
	}

	return assets, nil
}
