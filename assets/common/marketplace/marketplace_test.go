package marketplace

import (
	"net/http"
	"testing"
)

func TestMarketplace_Unit(t *testing.T) {
	type fields struct {
		client http.Client
		count  int
		scale  []float64
	}
	type args struct {
		base  string
		quote string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
	}{
		{
			name: t.Name(),
			args: args{
				base:  "btc",
				quote: "usdt",
			},
			want: 0,
		},
		{
			name: t.Name(),
			args: args{
				base:  "eth",
				quote: "usdt",
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Marketplace{
				client: tt.fields.client,
				count:  tt.fields.count,
				scale:  tt.fields.scale,
			}
			if got := p.Unit(tt.args.base, tt.args.quote); got > tt.want {
				t.Logf("Unit() response = %v, want > %v", got, tt.want)
			} else {
				t.Logf("Unit() zero = %v, want == %v", got, tt.want)
			}
		})
	}
}
