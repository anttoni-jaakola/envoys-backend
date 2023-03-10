package help

import "testing"

func TestNewCode(t *testing.T) {
	type args struct {
		length  int
		numbers bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: t.Name(),
			args: args{
				length:  6,
				numbers: false,
			},
		},
		{
			name: t.Name(),
			args: args{
				length:  6,
				numbers: true,
			},
		},
		{
			name: t.Name(),
			args: args{
				length:  12,
				numbers: false,
			},
		},
		{
			name: t.Name(),
			args: args{
				length:  12,
				numbers: true,
			},
		},
		{
			name: t.Name(),
			args: args{
				length:  50,
				numbers: false,
			},
		},
		{
			name: t.Name(),
			args: args{
				length:  50,
				numbers: true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := NewCode(tt.args.length, tt.args.numbers)
			if len(code) != tt.args.length {
				t.Errorf("NewCode(%d, %t) returned a string of length %d, expected %d", tt.args.length, tt.args.numbers, len(code), tt.args.length)
			}
			if tt.args.numbers {
				for _, c := range code {
					if c < '0' || c > 'z' || (c > '9' && c < 'A') || (c > 'Z' && c < 'a') {
						t.Errorf("NewCode(%d, %t) returned a string containing non-alphanumeric characters: %s", tt.args.length, tt.args.numbers, code)
						break
					}
				}
			}
			t.Logf("Generate new code: %v", code)
		})
	}
}
