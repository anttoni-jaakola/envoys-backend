package help

import "encoding/json"

// Comparable - This function is used to compare an array of bytes to an index string and additional column strings. It first attempts
// to unmarshal the byte array into a string array. If it is successful, it adds the additional column strings to the
// string array. If the index string is found in the resulting string array, it returns true, otherwise it returns false.
func Comparable(bytea []byte, index string, addColumn ...string) bool {

	// The purpose of the following is to declare a variable called array, which is a slice of strings. This allows the
	// programmer to create an array of strings that can be accessed and manipulated.
	var (
		array []string
	)

	// This code checks for any errors that may have occurred when decoding a JSON string into an array. If an error is
	// encountered, the function will return false.
	if err := json.Unmarshal(bytea, &array); err != nil {
		return false
	}

	// This code is checking to see if the given index exists in the array. The first line appends the given additional
	// column to the existing array. The second line uses the IndexOf() function to check if the given index is present in
	// the array. If it is, it returns true.
	array = append(array, addColumn...)
	if IndexOf(array, index) {
		return true
	}

	return false
}

// IndexOf - This function is used to check if a given element is present in a given collection. It takes in two parameters, the
// collection (which is an array of values of type T) and the element to be searched for (which is also of type T). It
// then loops through the collection and returns true if the specified element is found, and false if it is not.
func IndexOf[T comparable](collection []T, el T) bool {

	for _, x := range collection {
		if x == el {
			return true
		}
	}

	return false
}
