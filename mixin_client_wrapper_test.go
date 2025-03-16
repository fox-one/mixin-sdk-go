package mixin

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/shopspring/decimal"
)

func Test_buildTransferMany(t *testing.T) {

	memberAmountsArgs1 := []MemberAmount{}
	for i := 0; i < 566; i++ {
		memberAmountsArgs1 = append(memberAmountsArgs1, MemberAmount{
			Member: []string{fmt.Sprintf("member_id_%d", i)},
			Amount: decimal.NewFromInt(int64(i)),
		})
	}

	memberAmountsResult1 := make([][]MemberAmount, 3)
	for i := 0; i < MAX_UTXO_NUM; i++ {
		memberAmountsResult1[0] = append(memberAmountsResult1[0], MemberAmount{
			Member: []string{fmt.Sprintf("member_id_%d", i)},
			Amount: decimal.NewFromInt(int64(i)),
		})
	}

	for i := MAX_UTXO_NUM; i < MAX_UTXO_NUM*2; i++ {
		memberAmountsResult1[1] = append(memberAmountsResult1[1], MemberAmount{
			Member: []string{fmt.Sprintf("member_id_%d", i)},
			Amount: decimal.NewFromInt(int64(i)),
		})
	}
	for i := MAX_UTXO_NUM * 2; i < 566; i++ {
		memberAmountsResult1[2] = append(memberAmountsResult1[2], MemberAmount{
			Member: []string{fmt.Sprintf("member_id_%d", i)},
			Amount: decimal.NewFromInt(int64(i)),
		})
	}

	type args struct {
		memberAmounts []MemberAmount
	}
	tests := []struct {
		name    string
		args    args
		want    [][]MemberAmount
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				memberAmounts: memberAmountsArgs1,
			},
			want: memberAmountsResult1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTransferMany(tt.args.memberAmounts)
			t.Log("got:", got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildTransferMany() = %v, want %v", got, tt.want)
			}
		})
	}
}
