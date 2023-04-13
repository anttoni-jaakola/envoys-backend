package index

import (
	"github.com/cryptogateway/backend-envoys/assets"
)

// Service - The Service struct is used to create a structure that holds a pointer to an assets.Context. This allows the Service
// struct to access the assets.Context and all of its data, and to use that data to carry out tasks.
type Service struct {
	Context *assets.Context
}

// queryPrice - This function is used to calculate the price ratio between two different units of a currency. It takes the base and
// quote units as arguments and returns the ratio as a float64 and an error (if any). It uses the database to query the
// price of the two units, then calculates the ratio based on the prices.
func (i *Service) queryPrice(base, quote string) (ratio float64, err error) {

	// The purpose of this code is to create a variable called "scales" that is an empty slice of type float64. This
	// variable can then be used to store floats in a list.
	var (
		scales []float64
	)

	// This code is querying a database for two values: the price from trades for two provided units, and ordering the
	// results by their ID in descending order. It is then limiting the result set to the two most recent results. Finally,
	// it is deferring the closing of the rows until the function ends.
	rows, err := i.Context.Db.Query("select price from ohlcv where base_unit = $1 and quote_unit = $2 order by id desc limit 2", base, quote)
	if err != nil {
		return ratio, err
	}
	defer rows.Close()

	// The for loop is used to iterate over a collection of rows from the database. The rows.Next() function is used to get
	// the next row from the result set. This allows the program to loop through all the returned rows in order to process them.
	for rows.Next() {

		// The purpose of the following is to declare a variable named current of type float64. This variable will be used to
		// store a number with decimal places.
		var (
			current float64
		)

		// This statement is used to check for errors while scanning the rows of data. If an error is found, it will return the
		// ratio and the error, allowing the program to handle the error accordingly.
		if err := rows.Scan(&current); err != nil {
			return ratio, err
		}

		// The purpose of the line of code above is to add the variable "current" to the end of the existing list of values
		// stored in the variable "scales".
		scales = append(scales, current)
	}

	// This code is used to calculate the ratio between two scales. The ratio is calculated by subtracting the two scales
	// and then dividing by the second scale, and then multiplying by 100.
	if len(scales) == 2 {
		ratio = ((scales[0] - scales[1]) / scales[1]) * 100
	}

	return ratio, nil
}
