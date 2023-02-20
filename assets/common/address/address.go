package address

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/shengdoushi/base58"
	"strings"
)

type Address []byte

// New - The purpose of this code is to decode a parameter that can be of type []byte, string, or hexadecimal string and
// convert it to a byte array. Depending on the type, different decoding techniques are used, such as Base58 and hex. The
// code also checks for errors during the decoding process and returns nil if an error is found.
func New(src interface{}, input ...bool) Address {

	// The switch statement in the given code is a type switch. It is used to check the type of the variable src and then
	// execute a certain block of code depending on the type of the variable. This is useful when working with types that
	// have different behaviors and methods.
	switch src.(type) {
	case []byte:

		// This is a type assertion used to convert a value of interface type to a value of type []byte. It is used to convert
		// a value of interface type (the "src" variable in this case) to a value of type []byte. It is used to retrieve the
		// underlying []byte value of an interface and manipulate it.
		return src.([]byte)
	case string:

		param := src.(string)

		// The switch statement is used to execute one set of code among many possible sets of code depending on the value of
		// the parameter passed to it. In this case, the parameter is a variable called 'len(param)' which is presumably the
		// length of a given parameter. The switch statement will execute a different set of code depending on the value of the parameter.
		switch len(param) {
		case 34:

			// This code is attempting to decode a parameter using the Base58 algorithm and the BitcoinAlphabet. If the decoding
			// fails, it will return nil.
			decode, err := base58.Decode(param, base58.BitcoinAlphabet)
			if err != nil {
				return nil
			}

			return decode[:21]
		case 42, 44:

			// This checks if the input is not nil and if the first element of the input is truthy (not false or nil). It is used
			// to check if the input has a valid value before it is used.
			if input != nil && input[0] {

				// This code is checking for the presence of a hexadecimal value (0x41) at the beginning of the string 'param'. If it
				// is present, it is trimmed from the string. This code is likely used to ensure that the string does not contain any
				// extraneous characters that might interfere with its use.
				if strings.Contains(param, "0x41") {
					param = strings.TrimPrefix(param, "0x41")
				}

			} else {

				// This code is checking to see if the string parameter (param) contains the characters "0x". If it does, it removes
				// the "0x" prefix from the string.
				if strings.Contains(param, "0x") {
					param = strings.TrimPrefix(param, "0x")
				}
			}

			// This code is used to decode a parameter from a hexadecimal string to a byte array. It checks for any errors that
			// may occur during the decoding process and if an error is found, it returns nil.
			decode, err := hex.DecodeString(param)
			if err != nil {
				return nil
			}

			return decode
		case 66, 64:

			// This code is checking to see if the string parameter (param) contains the characters "0x". If it does, it removes
			// the "0x" prefix from the string.
			if strings.Contains(param, "0x") {
				param = strings.TrimPrefix(param, "0x")
			}

			// This code is used to decode a parameter from a hexadecimal string to a byte array. It checks for any errors that
			// may occur during the decoding process and if an error is found, it returns nil.
			decode, err := hex.DecodeString(param)
			if err != nil {
				return nil
			}

			return decode
		}

	}

	return nil
}

// Hex - This function is used to generate a hexadecimal string from an Address type. The hexadecimal string is returned with
// or without the "0x" prefix depending on the parameter passed in. If the parameter is true, the string will be returned
// with a "41" prefix, if it is false, the string will be returned with a "0x" prefix.
func (a Address) Hex(param ...bool) string {

	// This code is used to convert a byte array (a) to its hexadecimal representation. If the byte array is empty, it
	// assigns a "0" to the variable "compose" instead.
	compose := hex.EncodeToString(a)
	if len(compose) == 0 {
		compose = "0"
	}

	// The purpose of this code is to check if the length of the "compose" variable is equal to 64. If it is, then the code
	// will take the value of the "compose" variable and slice it from the 24th character onwards.
	if len(compose) == 64 {
		compose = compose[24:]
	}

	// This code checks to see if the variable 'param' is not nil and if the first element of the variable 'param' is true.
	// This is used to make sure that the variable 'param' exists and is not empty, and that the first element in the
	// variable is true.
	if param != nil && param[0] {

		// This code is checking if the string "compose" begins with the characters "41". If it does not, it is adding the
		// characters "41" to the beginning of the string. This is likely used to ensure that the string always begins with
		// "41" for some purpose.
		if !strings.HasPrefix(compose, "41") {
			compose = fmt.Sprintf("41%v", compose)
		}

		return compose
	}

	return "0x" + compose
}

// Base58 - This function is part of an Address struct and is used to generate a Base58 encoded string from the struct data. The
// generated string is created by appending a 4 byte hash of the original data to the original data and then encoding the
// result using Base58 encoding. This provides a way of securely representing the address data as a string.
func (a Address) Base58() string {

	b := a

	// This code assigns the result of the Hex() method of the b object to the compose variable. The Hex() method takes a
	// boolean parameter which determines if the output should be in upper or lower case.
	compose := b.Hex(true)

	// This code checks to see if a string (compose) begins with the value "41". If it does, the code within the "if" block
	// will be executed. This could be used to check if a string contains a specific set of characters for some purpose.
	if strings.HasPrefix(compose, "41") {

		// This code is used to decode a parameter from a hexadecimal string to a byte array. It checks for any errors that
		// may occur during the decoding process and if an error is found, it returns nil.
		serialize, err := hex.DecodeString(compose)
		if err != nil {
			return ""
		}

		a = serialize
	}

	// h256h0 is a variable that holds a reference to a new instance of the SHA256 cryptographic hash algorithm. This
	// instance is used to generate a hash of a given data set. The generated hash can then be used to verify the integrity
	// of the data set.
	h256h0 := sha256.New()

	// The purpose of h256h0.Write(a) is to write the byte slice a to the underlying array. This is used to write to a
	// specific memory address, allowing for the manipulation and storage of data.
	h256h0.Write(a)

	// h0 := h256h0.Sum(nil) is a hashing function that calculates the hash of the data that has been passed to it. The
	// purpose of the function is to create a unique identifier for the data that is passed in, allowing it to be stored,
	// identified and retrieved more easily.
	h0 := h256h0.Sum(nil)

	// h256h1 is used to create an instance of the SHA-256 hash algorithm. This is used to generate a
	// cryptographic hash value of data, which can be used in a variety of cryptographic operations such as digital
	// signatures.
	h256h1 := sha256.New()

	// The purpose of this code is to write the content of h0 into the h256h1 object. h256h1 is likely a type of object that
	// can contain data, such as a buffer, and h0 is likely an array of bytes. This code is used to copy the content of h0
	// into h256h1.
	h256h1.Write(h0)

	// h1 := h256h1.Sum(nil) is used to compute a 256-bit hash of the data in h256h1 and store it in the variable h1. This
	// is done by using the Sum() method of the h256h1 object which takes in a slice of bytes as an argument. The argument
	// is set to nil in this case which indicates that no additional data should be added to the hash computation.
	h1 := h256h1.Sum(nil)

	// The purpose of this code is to append the first four elements of the slice h1 to the slice a. The three dots (...)
	// after h1[:4] is necessary to unpack the slice h1[:4] into individual elements.
	input := a
	input = append(input, h1[:4]...)

	return base58.Encode(input, base58.BitcoinAlphabet)
}
