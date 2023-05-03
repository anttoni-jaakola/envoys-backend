package future

import (
	"context"
	"fmt"
	"testing"

	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/server/proto/v2/pbfuture"
)

func FutureTestApi_SetORder(t *testing.T) {
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
				Assigning:  "open",
				Position:   "long",
				Trading:    "limit",
				BaseUnit:   "eth",
				QuoteUnit:  "usd",
				Price:      28000.0,
				Quantity:   1000,
				Leverage:   10,
				TakeProfit: 0.0,
				StopLoss:   0.0,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			fmt.Printf(tt.args.Assigning)

			option := assets.Context{
				StoragePath: "/",
			}

			p := &Service{
				Context: &option,
			}
			ctx := context.Background()
			_, got := p.SetOrder(ctx, &tt.args)
			t.Logf(got.Error())
		})
	}
}

func FutureTestApi_ClosePosition(t *testing.T) {

}
