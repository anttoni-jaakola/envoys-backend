package future_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbfuture"
	"github.com/cryptogateway/backend-envoys/server/service/v2/future"
)

func TestApi_FutureSetOrder(t *testing.T) {
	type fields struct{}
	tests := []struct {
		name   string
		fields fields
		args   pbfuture.SetRequestOrder
		// args args
		want float64
		// want SetOrderResponse
		// wantErr bool
	}{
		{
			name: t.Name(),
			// fields: {},
			args: pbfuture.SetRequestOrder{
				Assigning:  "close",
				Position:   "short",
				OrderType:  "limit",
				BaseUnit:   "eth",
				QuoteUnit:  "usd",
				Price:      28000.0,
				Quantity:   0.1,
				Leverage:   10,
				TakeProfit: 0.0,
				StopLoss:   0.0,
				Mode:       "cross",
			},
			want: 0,
		},
		// {
		// 	name: t.Name(),
		// 	args: pbfuture.SetRequestOrder{
		// 		Assigning:  "open",
		// 		Position:   "long",
		// 		OrderType:  "limit",
		// 		BaseUnit:   "btc",
		// 		QuoteUnit:  "usd",
		// 		Price:      29000.0,
		// 		Quantity:   1,
		// 		Leverage:   10,
		// 		TakeProfit: 0.0,
		// 		StopLoss:   0.0,
		// 		Mode:       "cross",
		// 	},
		// 	want: 0,
		// }, {
		// 	name: t.Name(),
		// 	args: pbfuture.SetRequestOrder{
		// 		Assigning:  "close",
		// 		Position:   "long",
		// 		OrderType:  "limit",
		// 		BaseUnit:   "btc",
		// 		QuoteUnit:  "usd",
		// 		Price:      28500.0,
		// 		Quantity:   1,
		// 		Leverage:   10,
		// 		TakeProfit: 0.0,
		// 		StopLoss:   0.0,
		// 		Mode:       "cross",
		// 	},
		// 	want: 0,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fmt.Printf(tt.args.Assigning)

			dir, err := os.Getwd()

			root := strings.Join(strings.Split(dir, "/")[0:5], "/")
			fmt.Println(root)

			if err != nil {
				panic(err)
			}
			option := assets.Context{
				StoragePath: root,
			}
			option.Write()

			p := &future.Service{
				Context: &option,
			}
			ctx := context.Background()
			_, got := p.SetOrder(ctx, &tt.args)
			t.Log(got)
		})
	}
}

// func FutureTestApi_ClosePosition(t *testing.T) {

// }
