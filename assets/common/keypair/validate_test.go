package keypair

import (
	"github.com/cryptogateway/backend-envoys/server/proto/pbspot"
	"testing"
)

func TestValidateCryptoAddress(t *testing.T) {
	type args struct {
		address  string
		platform pbspot.Platform
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: t.Name(),
			args: args{
				address:  "0x0820fd02c0e00db67e3201ef1177cd96742c112d",
				platform: pbspot.Platform_ETHEREUM,
			},
		},
		{
			name: t.Name(),
			args: args{
				address:  "TPaBBdn4GFKq9M3NVTwobxdDNYAyPCoNfQ",
				platform: pbspot.Platform_TRON,
			},
		},
		{
			name: t.Name(),
			args: args{
				address:  "1H5hgupW1Zu81oDotVbwXChoDYNBTmoupU",
				platform: pbspot.Platform_BITCOIN,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateCryptoAddress(tt.args.address, tt.args.platform); (err != nil) != tt.wantErr {
				t.Errorf("ValidateCryptoAddress() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				t.Logf("ValidateCryptoAddress() access = %v", tt.args.address)
			}
		})
	}
}
