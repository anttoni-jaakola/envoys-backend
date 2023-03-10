package help

import (
	"net"
	"reflect"
	"testing"
	"time"
)

func TestPing(t *testing.T) {
	type args struct {
		rpc string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: t.Name(),
			args: args{
				rpc: "http://127.0.0.1:3081",
			},
		},
		{
			name: t.Name(),
			args: args{
				rpc: "http://127.0.0.1:3082",
			},
		},
		{
			name: t.Name(),
			args: args{
				rpc: "http://127.0.0.1:8080",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Ping(tt.args.rpc); got != tt.want {
				t.Logf("Ping() got = %v, want %v", got, tt.want)
			} else {
				t.Errorf("Ping() error = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_retryDial(t *testing.T) {
	timeout := time.Second * 15
	type args struct {
		addr    string
		timeout time.Duration
	}
	tests := []struct {
		name    string
		args    args
		want    net.Conn
		wantErr bool
	}{
		{
			name: t.Name(),
			args: args{
				addr:    "127.0.0.1:3081",
				timeout: timeout,
			},
		},
		{
			name: t.Name(),
			args: args{
				addr:    "127.0.0.1:3082",
				timeout: timeout,
			},
		},
		{
			name: t.Name(),
			args: args{
				addr:    "127.0.0.1:8080",
				timeout: timeout,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := retryDial(tt.args.addr, tt.args.timeout)
			if (err != nil) != tt.wantErr {
				t.Errorf("retryDial() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Logf("retryDial() got = %v, want %v", got, tt.want)
			}
		})
	}
}
