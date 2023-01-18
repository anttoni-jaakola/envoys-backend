package decimal

import (
	"github.com/shopspring/decimal"
	"math/big"
	"strconv"
)

type Float struct {
	decimal.Decimal
}

// New - new big float.
func New(value interface{}) *Float {

	number := decimal.NewFromFloat(0)
	switch v := value.(type) {
	case float64:
		number = decimal.NewFromFloat(v)
	case int:
		number = decimal.NewFromFloat(float64(v))
	case *big.Int:
		number, _ = decimal.NewFromString(v.String())
	}

	return &Float{
		number,
	}
}

// Mul - multiply the sum from the value.
func (p *Float) Mul(value float64) *Float {
	return &Float{p.Decimal.Mul(decimal.NewFromFloat(value))}
}

// Div - divide sum from value.
func (p *Float) Div(value float64) *Float {
	return &Float{p.Decimal.Div(decimal.NewFromFloat(value))}
}

// Sub - minus value from float.
func (p *Float) Sub(value float64) *Float {
	return &Float{p.Decimal.Sub(decimal.NewFromFloat(value))}
}

// Add - add value to float.
func (p *Float) Add(value float64) *Float {
	return &Float{p.Decimal.Add(decimal.NewFromFloat(value))}
}

// Float - convert big float to float64.
func (p *Float) Float() float64 {
	float, _ := p.Float64()
	return float
}

// Value - convert big float to float64.
func (p *Float) Value() float64 {
	if s, err := strconv.ParseFloat(p.String(), 32); err == nil {
		return s
	}
	return 0
}

// Round - rounding decimal value.
func (p *Float) Round(number int32) *Float {
	return &Float{p.Decimal.Round(number)}
}

// Int64 - convert to int64.
func (p *Float) Int64() int64 {
	return p.Decimal.IntPart()
}

// Integer - big int decimal.
func (p *Float) Integer(number int32) *big.Int {
	result := p.Decimal.Mul(decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(number))))

	impost := new(big.Int)
	impost.SetString(result.String(), 10)

	return impost
}

// Floating - big float decimal.
func (p *Float) Floating(number int32) float64 {
	num, _ := decimal.NewFromString(p.String())
	result := num.Div(decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(number))))

	if float, _ := result.Float64(); float > 0 {
		return float
	}

	return 0
}
