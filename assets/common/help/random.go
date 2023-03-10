package help

import (
	"math/rand"
	"time"
)

// The purpose of the above code is to create a string containing all the symbols used in English writing. This can be
// used to validate user input or generate random passwords.
const allSymbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()_+"

// NewCode - This function is used to generate a random string of a specified length with either all symbols or numbers. The
// rand.Seed() function is used to ensure that the string is truly random by using the current Unix time as a seed. The
// generateRandomString() function is then used to generate a random string of the specified length using the symbols
// provided. If the length of the generated string does not match the specified length, the function will continue to
// generate strings until the correct length is reached.
func NewCode(length int, numbers bool) string {

	// This statement is used to generate a random number using a seed value. The seed value used is the current time in
	// nanoseconds. This ensures that a new random number is generated each time the program is run.
	rand.Seed(time.Now().UnixNano())

	// This code is setting the variable symbols to the value of the allSymbols variable if the variable numbers is true. If
	// the variable numbers is false, symbols will be set to the string "1234567890".
	symbols := allSymbols
	if numbers {
		symbols = "1234567890"
	}

	// This code is used to generate a random string of a specified length using a set of symbols. The variable "result" is
	// initialized with an empty string, and a loop is used to continually generate a random string until it reaches the
	// desired length. The "generateRandomString()" function is used to generate a random string of a specified length using the given symbols.
	var result string
	for len(result) != length {
		result = generateRandomString(length, symbols)
	}

	return result
}

// generateRandomString - This function generates a random string of a given length with characters from the given string of symbols. This could
// be used to generate a random password, an activation code, or any other random string.
func generateRandomString(length int, symbols string) string {

	// The purpose of the code is to create a new empty slice of bytes with the length specified by the variable "length".
	// The result of the operation is stored in the variable "result".
	result := make([]byte, length)

	// The purpose of this code is to assign a random symbol from the "symbols" array to each element in the "result" array.
	// The "for i := range result" loop iterates over the elements in the "result" array, and the "result[i] =
	// symbols[rand.Intn(len(symbols))]" line assigns a random symbol from the "symbols" array to the current element in the
	// "result" array.
	for i := range result {
		result[i] = symbols[rand.Intn(len(symbols))]
	}

	return string(result)
}
