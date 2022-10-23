package address

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/shengdoushi/base58"
	"strings"
)

type Address []byte

// New - string/byte address to byte.
func New(src interface{}, input ...bool) Address {

	switch src.(type) {
	case []byte:
		return src.([]byte)
	case string:

		param := src.(string)

		switch len(param) {
		case 34:

			decode, err := base58.Decode(param, base58.BitcoinAlphabet)
			if err != nil {
				return nil
			}

			return decode[:21]
		case 42, 44:

			if input != nil && input[0] == true {
				if strings.Contains(param, "0x41") {
					param = strings.TrimPrefix(param, "0x41")
				}
			} else {
				if strings.Contains(param, "0x") {
					param = strings.TrimPrefix(param, "0x")
				}
			}

			decode, err := hex.DecodeString(param)
			if err != nil {
				return nil
			}

			return decode
		case 66, 64:

			if strings.Contains(param, "0x") {
				param = strings.TrimPrefix(param, "0x")
			}

			decode, err := hex.DecodeString(param)
			if err != nil {
				return nil
			}

			return decode
		}

	}

	return nil
}

// Hex - byte encode to string.
func (a Address) Hex(param ...bool) string {
	compose := hex.EncodeToString(a)
	if len(compose) == 0 {
		compose = "0"
	}

	if len(compose) == 64 {
		compose = compose[24:]
	}

	if param != nil && param[0] {

		if !strings.HasPrefix(compose, "41") {
			compose = fmt.Sprintf("41%v", compose)
		}

		return compose
	}

	return "0x" + compose
}

// Base58 - byte to string.
func (a Address) Base58() string {

	b := a

	compose := b.Hex(true)
	if strings.HasPrefix(compose, "41") {

		serialize, err := hex.DecodeString(compose)
		if err != nil {
			return ""
		}

		a = serialize
	}

	h256h0 := sha256.New()
	h256h0.Write(a)
	h0 := h256h0.Sum(nil)

	h256h1 := sha256.New()
	h256h1.Write(h0)
	h1 := h256h1.Sum(nil)

	input := a
	input = append(input, h1[:4]...)

	return base58.Encode(input, base58.BitcoinAlphabet)
}
