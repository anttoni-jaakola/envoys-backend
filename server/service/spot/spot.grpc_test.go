package spot

import (
	"context"
	"github.com/cryptogateway/backend-envoys/assets"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"reflect"
	"testing"
)

func TestService_GetPair(t *testing.T) {

	conn := &assets.Context{
		Development:     true,
		PostgresConnect: "postgres://envoys:envoys@localhost/envoys?sslmode=disable",
	}
	conn.Write()

	type fields struct {
		Context *assets.Context
	}

	type args struct {
		in0 context.Context
		req *pbspot.GetRequestPair
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pbspot.ResponsePair
		wantErr bool
	}{
		{
			name: t.Name(),
			args: args{
				in0: context.Background(),
				req: &pbspot.GetRequestPair{
					BaseUnit:  "eth",
					QuoteUnit: "usd",
				},
			},
			fields: fields{
				Context: conn,
			},
		},
		{
			name: t.Name(),
			args: args{
				in0: context.Background(),
				req: &pbspot.GetRequestPair{
					BaseUnit:  "trx",
					QuoteUnit: "usdt",
				},
			},
			fields: fields{
				Context: conn,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Service{
				Context: tt.fields.Context,
			}
			got, err := e.GetPair(tt.args.in0, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPair() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("GetPair() response = %v", got)
		})
	}
}

func TestService_GetCandles(t *testing.T) {

	conn := &assets.Context{
		Development:     true,
		PostgresConnect: "postgres://envoys:envoys@localhost/envoys?sslmode=disable",
	}
	conn.Write()

	type fields struct {
		Context *assets.Context
	}
	type args struct {
		in0 context.Context
		req *pbspot.GetRequestCandles
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pbspot.ResponseCandles
		wantErr bool
	}{
		{
			name: t.Name(),
			args: args{
				in0: context.Background(),
				req: &pbspot.GetRequestCandles{
					BaseUnit:  "eth",
					QuoteUnit: "usd",
				},
			},
			fields: fields{
				Context: conn,
			},
		},
		{
			name: t.Name(),
			args: args{
				in0: context.Background(),
				req: &pbspot.GetRequestCandles{
					BaseUnit:  "trx",
					QuoteUnit: "usdt",
				},
			},
			fields: fields{
				Context: conn,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Service{
				Context: tt.fields.Context,
			}
			got, err := e.GetCandles(tt.args.in0, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCandles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("GetCandles() response = %v", got)
		})
	}
}

func TestService_GetMarkers(t *testing.T) {

	conn := &assets.Context{
		Development:     true,
		PostgresConnect: "postgres://envoys:envoys@localhost/envoys?sslmode=disable",
	}
	conn.Write()

	type fields struct {
		Context *assets.Context
	}
	type args struct {
		in0 context.Context
		req *pbspot.GetRequestMarkers
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pbspot.ResponseMarker
		wantErr bool
	}{
		{
			name: t.Name(),
			args: args{
				in0: context.Background(),
				req: &pbspot.GetRequestMarkers{},
			},
			fields: fields{
				Context: conn,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Service{
				Context: tt.fields.Context,
			}
			got, err := e.GetMarkers(tt.args.in0, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMarkers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("GetMarkers() response = %v", got)
		})
	}
}

func TestService_GetPairs(t *testing.T) {

	conn := &assets.Context{
		Development:     true,
		PostgresConnect: "postgres://envoys:envoys@localhost/envoys?sslmode=disable",
	}
	conn.Write()

	type fields struct {
		Context *assets.Context
	}
	type args struct {
		in0 context.Context
		req *pbspot.GetRequestPairs
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pbspot.ResponsePair
		wantErr bool
	}{
		{
			name: t.Name(),
			args: args{
				in0: context.Background(),
				req: &pbspot.GetRequestPairs{
					Symbol: "usdt",
				},
			},
			fields: fields{
				Context: conn,
			},
		},
		{
			name: t.Name(),
			args: args{
				in0: context.Background(),
				req: &pbspot.GetRequestPairs{
					Symbol: "eth",
				},
			},
			fields: fields{
				Context: conn,
			},
		},
		{
			name: t.Name(),
			args: args{
				in0: context.Background(),
				req: &pbspot.GetRequestPairs{
					Symbol: "trx",
				},
			},
			fields: fields{
				Context: conn,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Service{
				Context: tt.fields.Context,
			}
			got, err := e.GetPairs(tt.args.in0, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPairs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("GetPairs() response = %v", got)
		})
	}
}

func TestService_GetPrice(t *testing.T) {

	conn := &assets.Context{
		Development:     true,
		PostgresConnect: "postgres://envoys:envoys@localhost/envoys?sslmode=disable",
	}
	conn.Write()

	type fields struct {
		Context *assets.Context
	}
	type args struct {
		in0 context.Context
		req *pbspot.GetRequestPriceManual
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pbspot.ResponsePrice
		wantErr bool
	}{
		{
			name: t.Name(),
			args: args{
				in0: context.Background(),
				req: &pbspot.GetRequestPriceManual{
					BaseUnit:  "eth",
					QuoteUnit: "usdt",
				},
			},
			fields: fields{
				Context: conn,
			},
		},
		{
			name: t.Name(),
			args: args{
				in0: context.Background(),
				req: &pbspot.GetRequestPriceManual{
					BaseUnit:  "trx",
					QuoteUnit: "usdt",
				},
			},
			fields: fields{
				Context: conn,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Service{
				Context: tt.fields.Context,
			}
			got, err := e.GetPrice(tt.args.in0, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPrice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("GetPrice() response = %v", got)
		})
	}
}

func TestService_GetSymbol(t *testing.T) {

	conn := &assets.Context{
		Development:     true,
		PostgresConnect: "postgres://envoys:envoys@localhost/envoys?sslmode=disable",
	}
	conn.Write()

	type fields struct {
		Context *assets.Context
	}
	type args struct {
		in0 context.Context
		req *pbspot.GetRequestSymbol
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pbspot.ResponseSymbol
		wantErr bool
	}{
		{
			name: t.Name(),
			args: args{
				in0: context.Background(),
				req: &pbspot.GetRequestSymbol{
					BaseUnit:  "eth",
					QuoteUnit: "usdt",
				},
			},
			fields: fields{
				Context: conn,
			},
			want: &pbspot.ResponseSymbol{
				Success: true,
			},
		},
		{
			name: t.Name(),
			args: args{
				in0: context.Background(),
				req: &pbspot.GetRequestSymbol{
					BaseUnit:  "trx",
					QuoteUnit: "usdt",
				},
			},
			fields: fields{
				Context: conn,
			},
			want: &pbspot.ResponseSymbol{
				Success: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Service{
				Context: tt.fields.Context,
			}
			got, err := e.GetSymbol(tt.args.in0, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSymbol() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSymbol() got = %v, want %v", got, tt.want)
			} else {
				t.Logf("GetPrice() response = %v", got)
			}
		})
	}
}
