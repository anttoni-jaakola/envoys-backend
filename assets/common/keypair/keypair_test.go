package keypair

import (
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"testing"
)

func TestCrossChain_New(t *testing.T) {
	type fields struct {
		extended *hdkeychain.ExtendedKey
	}
	type args struct {
		secret   string
		bytea    []byte
		platform pbspot.Platform
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantA   string
		wantP   string
		wantErr bool
	}{
		{
			name: t.Name(),
			args: args{
				secret:   "test",
				platform: pbspot.Platform_ETHEREUM,
			},
		},
		{
			name: t.Name(),
			args: args{
				secret:   "test",
				platform: pbspot.Platform_TRON,
			},
		},
		{
			name: t.Name(),
			args: args{
				secret:   "test",
				platform: pbspot.Platform_BITCOIN,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &CrossChain{
				extended: tt.fields.extended,
			}
			gotA, gotP, err := s.New(tt.args.secret, tt.args.bytea, tt.args.platform)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotA != tt.wantA {
				t.Logf("New() gotA = %v, want %v", gotA, tt.wantA)
			}
			if gotP != tt.wantP {
				t.Logf("New() gotP = %v, want %v", gotP, tt.wantP)
			}
		})
	}
}
