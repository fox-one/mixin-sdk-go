package mixinnet

const (
	ExtraSizeGeneralLimit    = 256
	ExtraSizeStorageStep     = 1024
	ExtraSizeStorageCapacity = 1024 * 1024 * 4
	ExtraStoragePriceStep    = "0.001"
	SliceCountLimit          = 256
	ReferencesCountLimit     = 2
)

var (
	XINAssetId Hash
)

func init() {
	XINAssetId = NewHash([]byte("c94ac88f-4671-3976-b60a-09064f1811e8"))
}

func (tx *Transaction) ExtraLimit() int {
	if tx.Version < TxVersionReferences {
		return ExtraSizeGeneralLimit
	}
	if tx.Asset != XINAssetId {
		return ExtraSizeGeneralLimit
	}
	if len(tx.Outputs) < 1 {
		return ExtraSizeGeneralLimit
	}
	out := tx.Outputs[0]
	if len(out.Keys) != 1 {
		return ExtraSizeGeneralLimit
	}
	if out.Type != OutputTypeScript {
		return ExtraSizeGeneralLimit
	}
	if out.Script.String() != "fffe40" {
		return ExtraSizeGeneralLimit
	}
	step := IntegerFromString(ExtraStoragePriceStep)
	if out.Amount.Cmp(step) < 0 {
		return ExtraSizeGeneralLimit
	}
	cells := out.Amount.Count(step)
	limit := cells * ExtraSizeStorageStep
	if limit > ExtraSizeStorageCapacity {
		return ExtraSizeStorageCapacity
	}
	return int(limit)
}
