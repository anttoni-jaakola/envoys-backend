package decimal

import (
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
)

type Float struct {
	decimal.Decimal
}

// New - This function creates a new Float object from an interface, which could be a float64, int64, int, or *big.Int. It
// takes the value from the interface and converts it to a decimal.NewFromFloat, decimal.NewFromInt, or
// decimal.NewFromString, and then stores it in the Float object.
func New(value interface{}) *Float {

	// This code is used to create a new decimal object with a value of 0.0. The decimal object can then be used to store
	// and manipulate decimal values in an efficient way.
	number := decimal.NewFromFloat(0)

	// This switch statement is used to convert different types of values into a decimal. The statement will check the type
	// of the value, then use the corresponding function to convert it into decimal. For example, if the value is a float64,
	// then decimal.NewFromFloat will be used.
	switch v := value.(type) {
	case float64:
		number = decimal.NewFromFloat(v)
	case int64:
		number = decimal.NewFromInt(v)
	case int:
		number = decimal.NewFromInt(int64(v))
	case *big.Int:
		number, _ = decimal.NewFromString(v.String())
	}

	return &Float{
		number,
	}
}

// Mul - This function is used to multiply a float value with another float value and return a new float value. The function
// takes a pointer to a float type and a float64 value as parameters and returns a pointer to a new float type containing
// the result of the multiplication.
func (p *Float) Mul(value float64) *Float {
	return &Float{p.Decimal.Mul(decimal.NewFromFloat(value))}
}

// Div - This function is used to divide a float value by another float value and return the result as a Float type. The
// 'Div()' function is a method of the decimal package, and it takes in a float64 value as an argument. The function
// returns the result as a Float type instead of a float64 type to ensure precision and accuracy of the computation.
func (p *Float) Div(value float64) *Float {
	return &Float{p.Decimal.Div(decimal.NewFromFloat(value))}
}

// Sub - This is a method of a Float type in Go. The purpose of this method is to subtract a given float64 value from the Float
// type value and then return the resulting Float type.
func (p *Float) Sub(value float64) *Float {
	return &Float{p.Decimal.Sub(decimal.NewFromFloat(value))}
}

// Add - This function adds a float64 value to a Float type which is a struct containing a decimal.Decimal struct. The function
// returns a pointer to a new Float struct with the new decimal.Decimal value.
func (p *Float) Add(value float64) *Float {
	return &Float{p.Decimal.Add(decimal.NewFromFloat(value))}
}

// Float - This is a method of the Float type. It is used to convert a Float to a float64 value by returning the float64
// representation of the Float type. It returns an error if the conversion fails.
func (p *Float) Float() float64 {
	float, _ := p.Float64()
	return float
}

// Value - This function is used to convert a value of type Float to a float64 type. It first attempts to convert the Float value
// to a string, and then uses the ParseFloat function to convert the string to a float64. If the conversion is
// successful, the float64 value is returned, otherwise 0 is returned.
func (p *Float) Value() float64 {
	if s, err := strconv.ParseFloat(p.String(), 64); err == nil {
		return s
	}
	return 0
}

// Round - This function is used to round a float number to the specified number of decimal places. It takes a pointer to a Float
// type and an integer as parameters and returns a pointer to a Float type with the rounded number.
func (p *Float) Round(number int32) *Float {
	return &Float{p.Decimal.Round(number)}
}

// Int64 - convert to int64.
func (p *Float) Int64() int64 {
	return p.Decimal.IntPart()
}

// Integer - This function is used to convert a float to a big integer. It uses the Decimal library to convert a float to a
// decimal, multiplies it by 10 to the specified number, and then returns a big integer.
func (p *Float) Integer(number int32) *big.Int {

	// This assignment statement is used to multiply two decimals together. The "result" variable is assigned the product of
	// the multiplication of the two decimals. The first decimal is "p.Decimal" and the other is the result of the
	// expression "decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(number)))" which is 10 raised to the
	// power of "number".
	result := p.Decimal.Mul(decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(number))))

	// The purpose of this code is to create a new big integer (big.Int) and set its value to the result of a previous
	// operation, represented as a string in base 10.
	impost := new(big.Int)
	impost.SetString(result.String(), 10)

	return impost
}

// Floating - This function is used to perform floating-point arithmetic operations on a given number. It takes an integer as an
// argument and divides it by 10 raised to the power of the given number. It then returns the result as a float64. It
// also checks to make sure that the result is positive before returning it.
func (p *Float) Floating(number int32) float64 {

	// This line of code is using the decimal package to convert a string to a decimal number. The underscore character is
	// used to ignore the error value returned by the NewFromString() function. The num variable will hold the converted
	// decimal number.
	num, _ := decimal.NewFromString(p.String())

	// This statement is using the Decimal library to divide the value of "num" by 10 to the power of the value of "number".
	// This is being stored in the variable "result".
	result := num.Div(decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(number))))

	// This code is checking if the variable "result" contains a float that is greater than 0. If it does, it returns the float value.
	if float, _ := result.Float64(); float > 0 {
		return float
	}

	return 0
}
