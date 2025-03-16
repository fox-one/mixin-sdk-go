package mixin

import (
	"context"
	"errors"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
	"github.com/shopspring/decimal"
)

var (
	ErrConfigNil             = errors.New("config is nil")
	ErrNotEnoughUtxos        = errors.New("not enough utxos")
	ErrInscriptionNotFound   = errors.New("inscription not found")
	ErrMultInscriptionsFound = errors.New("multiple inscriptions found")
)

var (
	defaultMaxRetry = 3

	AGGREGRATE_UTXO_MEMO = "arrgegate utxos"
)

const (
	MAX_UTXO_NUM = 255
)

type ClientWrapper struct {
	*Client
	user     *User
	SpendKey mixinnet.Key

	transferMutex sync.Mutex
}

func NewMixinClientWrapper(keystore *Keystore, spendKeyStr string) (*ClientWrapper, error) {
	client, err := NewFromKeystore(keystore)
	if err != nil {
		return nil, err
	}
	user, err := client.UserMe(context.Background())
	if err != nil {
		return nil, err
	}

	spendKey, err := mixinnet.ParseKeyWithPub(spendKeyStr, user.SpendPublicKey)
	if err != nil {
		return nil, err
	}
	clientWrapper := &ClientWrapper{
		Client:        client,
		SpendKey:      spendKey,
		user:          user,
		transferMutex: sync.Mutex{},
	}

	return clientWrapper, nil
}

type TransferOneRequest struct {
	RequestId string
	AssetId   string

	Member string
	Amount decimal.Decimal
	Memo   string
}

type MemberAmount struct {
	Member []string
	Amount decimal.Decimal
}

type TransferManyRequest struct {
	RequestId string
	AssetId   string

	MemberAmount []MemberAmount
	Memo         string
}

type InscriptionTransferRequest struct {
	RequestId   string
	AssetId     string
	Inscription string

	Memo   string
	Member string
}

// 主动聚合utxos 至 utxo 数量不超过 255 个
func (c *ClientWrapper) SyncArrgegateUtxos(ctx context.Context, assetId string) (utxos []*SafeUtxo, err error) {
	c.transferMutex.Lock()
	defer c.transferMutex.Unlock()

	utxos = make([]*SafeUtxo, 0)

	for {
		requestId := RandomTraceID()
		utxos, err = c.SafeListUtxos(ctx, SafeListUtxoOption{
			Asset:     assetId,
			State:     SafeUtxoStateUnspent,
			Threshold: 1,
		})

		if err != nil {
			return
		}

		if len(utxos) <= MAX_UTXO_NUM {
			// 主动聚合完成
			break
		}

		// 将utxos分割成 255 个大小的数组
		utxoSlice := make([]*SafeUtxo, 0, MAX_UTXO_NUM)
		utxoSliceAmount := decimal.Zero
		for i := 0; i < len(utxos); i++ {
			if utxos[i].InscriptionHash.HasValue() { // skip inscription
				continue
			}
			utxoSlice = append(utxoSlice, utxos[i])
			utxoSliceAmount = utxoSliceAmount.Add(utxos[i].Amount)
		}

		// 2: build transaction
		b := NewSafeTransactionBuilder(utxoSlice)
		b.Memo = AGGREGRATE_UTXO_MEMO
		var tx *mixinnet.Transaction
		tx, err = c.MakeTransaction(ctx, b, []*TransactionOutput{
			{
				Address: RequireNewMixAddress([]string{c.ClientID}, 1),
				Amount:  utxoSliceAmount,
			},
		})
		if err != nil {
			return nil, err
		}
		var raw string
		raw, err = tx.Dump()
		if err != nil {
			return
		}

		// 3. create transaction
		var request *SafeTransactionRequest
		request, err = c.SafeCreateTransactionRequest(ctx, &SafeTransactionRequestInput{
			RequestID:      requestId,
			RawTransaction: raw,
		})
		if err != nil {
			return
		}

		// 4. sign transaction
		err = SafeSignTransaction(
			tx,
			c.SpendKey,
			request.Views,
			0,
		)
		if err != nil {
			return
		}

		var signedRaw string
		signedRaw, err = tx.Dump()
		if err != nil {
			return
		}

		// 5. submit transaction
		_, err = c.SafeSubmitTransactionRequest(ctx, &SafeTransactionRequestInput{
			RequestID:      requestId,
			RawTransaction: signedRaw,
		})
		if err != nil {
			return
		}

		// 重试读取交易状态
		const defaultMaxRetryTimes = 3
		retryTimes := 0
		for {
			if retryTimes >= defaultMaxRetryTimes {
				break
			}

			retryTimes++
			time.Sleep(time.Second * time.Duration(retryTimes))
			_, err = c.SafeReadTransactionRequest(ctx, requestId)
			if err != nil {
				return
			} else {
				break
			}
		}

		// wait 250ms ...
		time.Sleep(time.Second >> 2)
	}
	return
}

func (m *ClientWrapper) TransferOneWithRetry(ctx context.Context, req *TransferOneRequest) error {
	var err error
	for i := 0; i < defaultMaxRetry; i++ {
		if _, err = m.transferOne(ctx, req); err != nil {
			time.Sleep(time.Second << i)
			continue
		} else {
			return nil
		}
	}
	return err
}

func (c *ClientWrapper) transferOne(ctx context.Context, req *TransferOneRequest) (*SafeTransactionRequest, error) {
	var err error
	var utxos []*SafeUtxo

	utxos, err = c.SyncArrgegateUtxos(ctx, req.AssetId)
	if err != nil {
		return nil, err
	}

	c.transferMutex.Lock()
	defer c.transferMutex.Unlock()

	for i := 0; i < 3 && len(utxos) == 0; i++ {
		utxos, _ = c.SafeListUtxos(ctx, SafeListUtxoOption{
			Asset:     req.AssetId,
			State:     SafeUtxoStateUnspent,
			Threshold: 1,
			Limit:     500,
		})
		if len(utxos) > 0 {
			break
		}
		time.Sleep(time.Second << 1)
	}

	if len(utxos) == 0 {
		return nil, ErrNotEnoughUtxos
	}

	for i := 0; i < len(utxos); i++ {
		if utxos[i].InscriptionHash.HasValue() {
			utxos = append(utxos[:i], utxos[i+1:]...)
			i--
		}
	}

	// 1: select utxos
	sort.Slice(utxos, func(i, j int) bool {
		return utxos[i].Amount.LessThanOrEqual(utxos[j].Amount)
	})
	var useAmount decimal.Decimal
	var useUtxos []*SafeUtxo

	for _, utxo := range utxos {
		useAmount = useAmount.Add(utxo.Amount)
		useUtxos = append(useUtxos, utxo)

		if len(useUtxos) > MAX_UTXO_NUM {
			useUtxos = useUtxos[1:]
			useAmount = useAmount.Sub(utxos[0].Amount)
		}

		if useAmount.GreaterThanOrEqual(req.Amount) {
			break
		}
	}

	if useAmount.LessThan(req.Amount) {
		return nil, ErrNotEnoughUtxos
	}

	// 2: build transaction
	b := NewSafeTransactionBuilder(useUtxos)
	b.Memo = req.Memo

	txOutout := &TransactionOutput{
		Address: RequireNewMixAddress([]string{req.Member}, 1),
		Amount:  req.Amount,
	}

	tx, err := c.MakeTransaction(ctx, b, []*TransactionOutput{txOutout})
	if err != nil {
		return nil, err
	}

	raw, err := tx.Dump()
	if err != nil {
		return nil, err
	}

	// 3. create transaction
	request, err := c.SafeCreateTransactionRequest(ctx, &SafeTransactionRequestInput{
		RequestID:      req.RequestId,
		RawTransaction: raw,
	})
	if err != nil {
		return nil, err
	}
	// 4. sign transaction
	err = SafeSignTransaction(
		tx,
		c.SpendKey,
		request.Views,
		0,
	)
	if err != nil {
		return nil, err
	}
	signedRaw, err := tx.Dump()
	if err != nil {
		return nil, err
	}

	// 5. submit transaction
	_, err = c.SafeSubmitTransactionRequest(ctx, &SafeTransactionRequestInput{
		RequestID:      req.RequestId,
		RawTransaction: signedRaw,
	})
	if err != nil {
		return nil, err
	}

	// 6. read transaction
	req1, err := c.SafeReadTransactionRequest(ctx, req.RequestId)
	if err != nil {
		return nil, err
	}
	return req1, nil
}

func (m *ClientWrapper) TransferManyWithRetry(ctx context.Context, req *TransferManyRequest) error {
	if len(req.MemberAmount) < MAX_UTXO_NUM {
		return m.transferManyWithRetry(ctx, req)
	} else {
		memberAmountArray := buildTransferMany(req.MemberAmount)
		for i, memberAmount := range memberAmountArray {
			req := &TransferManyRequest{
				RequestId:    GenUuidFromStrings(req.RequestId, strconv.Itoa(i)),
				AssetId:      req.AssetId,
				MemberAmount: memberAmount,
				Memo:         req.Memo,
			}

			err := m.transferManyWithRetry(ctx, req)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *ClientWrapper) transferManyWithRetry(ctx context.Context, req *TransferManyRequest) error {
	var err error
	for i := 0; i < defaultMaxRetry; i++ {
		if _, err = m.transferMany(ctx, req); err != nil {
			time.Sleep(time.Second << i)
			continue
		} else {
			return nil
		}
	}
	return err
}

func (m *ClientWrapper) transferMany(ctx context.Context, req *TransferManyRequest) (*SafeTransactionRequest, error) {
	var utxos []*SafeUtxo
	var err error
	utxos, err = m.SyncArrgegateUtxos(ctx, req.AssetId)
	if err != nil {
		return nil, err
	}

	m.transferMutex.Lock()
	defer m.transferMutex.Unlock()

	totalAmount := decimal.Zero
	for _, item := range req.MemberAmount {
		totalAmount = totalAmount.Add(item.Amount)
	}

	retryCount := 0
	for len(utxos) == 0 && retryCount < 3 {
		// 1. 将utxos聚合
		utxos, _ = m.SafeListUtxos(ctx, SafeListUtxoOption{
			Asset:     req.AssetId,
			State:     SafeUtxoStateUnspent,
			Threshold: 1,
			Limit:     500,
		})
		time.Sleep(time.Second << retryCount)
		retryCount++
	}
	if len(utxos) == 0 {
		return nil, ErrNotEnoughUtxos
	}

	// 1: select utxos
	sort.Slice(utxos, func(i, j int) bool {
		return utxos[i].Amount.LessThanOrEqual(utxos[j].Amount)
	})

	var useAmount decimal.Decimal
	var useUtxos []*SafeUtxo
	for _, utxo := range utxos {
		useAmount = useAmount.Add(utxo.Amount)
		useUtxos = append(useUtxos, utxo)

		if len(useUtxos) > MAX_UTXO_NUM {
			useUtxos = useUtxos[1:]
			useAmount = useAmount.Sub(utxos[0].Amount)
		}

		if useAmount.GreaterThanOrEqual(totalAmount) {
			break
		}
	}

	if useAmount.LessThan(totalAmount) {
		return nil, ErrNotEnoughUtxos
	}

	// 2: build transaction
	b := NewSafeTransactionBuilder(useUtxos)
	b.Memo = req.Memo

	txOutout := make([]*TransactionOutput, len(req.MemberAmount))
	for i := 0; i < len(req.MemberAmount); i++ {
		txOutout[i] = &TransactionOutput{
			Address: RequireNewMixAddress(req.MemberAmount[i].Member, byte(len(req.MemberAmount[i].Member))),
			Amount:  req.MemberAmount[i].Amount,
		}
	}

	tx, err := m.MakeTransaction(ctx, b, txOutout)
	if err != nil {
		return nil, err
	}

	raw, err := tx.Dump()
	if err != nil {
		return nil, err
	}

	// 3. create transaction
	request, err := m.SafeCreateTransactionRequest(ctx, &SafeTransactionRequestInput{
		RequestID:      req.RequestId,
		RawTransaction: raw,
	})
	if err != nil {
		return nil, err
	}
	// 4. sign transaction
	err = SafeSignTransaction(
		tx,
		m.SpendKey,
		request.Views,
		0,
	)
	if err != nil {
		return nil, err
	}
	signedRaw, err := tx.Dump()
	if err != nil {
		return nil, err
	}

	// 5. submit transaction
	_, err = m.SafeSubmitTransactionRequest(ctx, &SafeTransactionRequestInput{
		RequestID:      req.RequestId,
		RawTransaction: signedRaw,
	})
	if err != nil {
		return nil, err
	}

	// 6. read transaction
	req1, err := m.SafeReadTransactionRequest(ctx, req.RequestId)
	if err != nil {
		return nil, err
	}
	return req1, nil
}

// 一个功能函数，将一个数组中的多个元素切分成 n个数组，每个数组长度最多不超过255个
func buildTransferMany(memberAmounts []MemberAmount) [][]MemberAmount {
	result := make([][]MemberAmount, (len(memberAmounts)+MAX_UTXO_NUM-1)/MAX_UTXO_NUM)
	for i := 0; i < len(result); i++ {
		start := i * MAX_UTXO_NUM
		end := (i + 1) * MAX_UTXO_NUM
		if end > len(memberAmounts) {
			end = len(memberAmounts)
		}
		result[i] = memberAmounts[start:end]
	}
	return result
}

func (m *ClientWrapper) InscriptionTransfer(ctx context.Context, req *InscriptionTransferRequest) error {
	var utxos []*SafeUtxo
	utxos, _ = m.Client.SafeListUtxos(ctx, SafeListUtxoOption{
		Asset:     req.AssetId,
		State:     SafeUtxoStateUnspent,
		Threshold: 1,
		Limit:     500,
	})

	for i := range utxos {
		if utxos[i].InscriptionHash.String() == req.Inscription {
			utxos = []*SafeUtxo{utxos[i]}
		}
	}

	if len(utxos) == 0 {
		return ErrInscriptionNotFound
	}

	if len(utxos) > 1 {
		return ErrMultInscriptionsFound
	}

	b := NewSafeTransactionBuilder(utxos)
	b.Memo = req.Memo

	txOutout := &TransactionOutput{
		Address: RequireNewMixAddress([]string{req.Member}, 1),
		Amount:  utxos[0].Amount,
	}

	tx, err := m.Client.MakeTransaction(ctx, b, []*TransactionOutput{txOutout})
	if err != nil {
		return err
	}

	raw, err := tx.Dump()
	if err != nil {
		return err
	}

	// 3. create transaction
	request, err := m.Client.SafeCreateTransactionRequest(ctx, &SafeTransactionRequestInput{
		RequestID:      req.RequestId,
		RawTransaction: raw,
	})
	if err != nil {
		return err
	}
	// 4. sign transaction
	err = SafeSignTransaction(
		tx,
		m.SpendKey,
		request.Views,
		0,
	)
	if err != nil {
		return err
	}
	signedRaw, err := tx.Dump()
	if err != nil {
		return err
	}

	// 5. submit transaction
	_, err = m.Client.SafeSubmitTransactionRequest(ctx, &SafeTransactionRequestInput{
		RequestID:      req.RequestId,
		RawTransaction: signedRaw,
	})
	if err != nil {
		return err
	}

	// 6. read transaction
	_, err = m.Client.SafeReadTransactionRequest(ctx, req.RequestId)
	if err != nil {
		return err
	}
	return nil
}
