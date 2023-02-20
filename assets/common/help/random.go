package help

import (
	"fmt"
	"math/rand"
	"regexp"
	"time"
)

const allSymbols = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890!@#$%^&*()_+"

// NewCode - This function is used to generate a random code of a specified length and character set. The length and character set
// are specified by the parameters 'length' and 'numbers'. The code is generated using a random number generator, and a
// regex pattern is used to ensure that the code matches the specified character set.
func NewCode(length int, numbers bool) string {

	// This function is used to set the initial seed value for the math/rand package. It uses the current time to initialize
	// the random number generator, ensuring that each time the program is run it will produce a different sequence of
	// random numbers.
	rand.Seed(time.Now().UnixNano())

	// This code is used to generate a regular expression string that checks the length of a string. The pattern variable is
	// a regular expression string that will check if a string is between a certain length. The length variable is used to
	// set the length of the string that should be checked. The numbers variable is used to determine if the pattern string
	// should only check for numbers. If the numbers variable is true, the pattern string will only check for numbers,
	// otherwise it will check for any alphanumeric characters.
	pattern := fmt.Sprintf("^[a-zA-Z0-9!@#$%%^&*()_+]{%d}$", length)
	if numbers {
		pattern = fmt.Sprintf("^[0-9]{%d}$", length)
	}

	// The purpose of the variable 'result' is to store a pointer to a string. This allows the variable to point to a string
	// that is located in memory, instead of having to store the actual string itself. This can be useful when dealing with
	// large strings, since it can save memory.
	var result *string

	// This code is used to generate a random string that matches a pattern. It uses a loop to run until a random string is
	// generated that matches the given pattern. The "result" variable is initially set to nil, then the
	// regexp.MustCompile() function is used to check if the generated string matches the pattern. If it does not match, a
	// new random string is generated until a matching string is found.
	for result == nil || !regexp.MustCompile(pattern).MatchString(*result) {
		s := generateRandomString(length, allSymbols)
		result = &s
	}

	return *result
}

// generateRandomString - This function is used to generate a random string of a given length with given symbols. It uses the rand.Int31n
// function to generate the random string.
func generateRandomString(length int, symbols string) string {

	// The purpose of this line of code is to create a byte slice, called "result", of a specified length. The length of the
	// byte slice is specified by the value of the "length" variable.
	result := make([]byte, length)

	// This for loop is used to randomly assign a value from the symbols array to each element in the result array. It uses
	// the Int31n method from the rand package to generate the random index for the symbols array.
	for i := range result {
		result[i] = symbols[rand.Int31n(int32(len(symbols)))]
	}

	return string(result)
}
