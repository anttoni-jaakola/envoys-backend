package help

import "encoding/json"

func Comparable(bytea []byte, index string, addColumn ...string) bool {

	var (
		array []string
	)

	if err := json.Unmarshal(bytea, &array); err != nil {
		return false
	}

	array = append(array, addColumn...)
	if IndexOf(array, index) {
		return true
	}

	return false
}

func IndexOf[T comparable](collection []T, el T) bool {
	for _, x := range collection {
		if x == el {
			return true
		}
	}
	return false
}
