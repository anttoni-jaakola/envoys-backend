package address

import (
	"strings"
	"testing"
)

func TestAddress_Base58(t *testing.T) {
	tests := []struct {
		name string
		a    Address
		want string
	}{
		{
			name: t.Name(),
			a:    New("0x418840e6c55b9ada326d211d818c34a994aeced808"),
			want: "TNPeeaaFB7K9cmo4uQpcU32zGK8G1NYqeL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Base58(); got != tt.want {
				t.Errorf("Base58() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAddress_Hex(t *testing.T) {
	type args struct {
		param []bool
	}
	tests := []struct {
		name string
		a    Address
		args args
		want string
	}{
		{
			name: t.Name(),
			a:    New("TNPeeaaFB7K9cmo4uQpcU32zGK8G1NYqeL"),
			want: "0x418840e6c55b9ada326d211d818c34a994aeced808",
			args: args{
				param: []bool{
					false,
				},
			},
		},
		{
			name: t.Name(),
			a:    New("TNPeeaaFB7K9cmo4uQpcU32zGK8G1NYqeL"),
			want: "0x418840e6c55b9ada326d211d818c34a994aeced808",
			args: args{
				param: []bool{
					true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Hex(tt.args.param...); got != tt.want && got != strings.TrimPrefix(tt.want, "0x") {
				t.Errorf("Hex() = %v, want %v", got, strings.ToLower(tt.want))
			}
		})
	}
}
