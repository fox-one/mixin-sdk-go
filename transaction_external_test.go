package mixin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadExternalTransactions(t *testing.T) {
	ctx := context.Background()
	transactions, err := ReadExternalTransactions(ctx, "", "", "")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	for _, transaction := range transactions {
		assert.NotEmpty(t, transaction.TransactionID)
		assert.True(t, transaction.Amount.IsPositive())
	}
}
