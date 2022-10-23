package decimal

import (
	"github.com/shopspring/decimal"
	"math/big"
)

const (
	persistentDecimals = 32
)

type Decimal struct {
	decimal.Decimal
}

func FromFloat(source float64) Decimal {
	return Decimal{decimal.NewFromFloat(source).Round(persistentDecimals)}
}

func FromInt(source int64) Decimal {
	return Decimal{decimal.NewFromInt(source).Round(persistentDecimals)}
}

func (a Decimal) Add(b Decimal) Decimal {
	return Decimal{a.Decimal.Add(b.Decimal)}
}

func (a Decimal) Sub(b Decimal) Decimal {
	return Decimal{a.Decimal.Sub(b.Decimal)}
}

func (a Decimal) Div(b Decimal) Decimal {
	return Decimal{a.Decimal.DivRound(b.Decimal, persistentDecimals)}
}

func (a Decimal) Mul(b Decimal) Decimal {
	return Decimal{a.Decimal.Mul(b.Decimal).Round(persistentDecimals)}
}

func (a Decimal) Neg() Decimal {
	return Decimal{a.Decimal.Neg()}
}

func (a Decimal) Cmp(b Decimal) int {
	return a.Decimal.Cmp(b.Decimal)
}

func (a Decimal) Floor() Decimal {
	return Decimal{a.Decimal.Floor()}
}

func (a Decimal) Ceil() Decimal {
	return Decimal{a.Decimal.Ceil()}
}

func (a Decimal) Round(decimals int32) Decimal {
	return Decimal{a.Decimal.Round(decimals)}
}

func (a Decimal) Equal(b Decimal) bool {
	return a.Decimal.Equal(b.Decimal)
}

func (a Decimal) Persist() string {
	return a.Decimal.String()
}

func (a Decimal) Float64() float64 {
	f, _ := a.Decimal.Float64()
	return f
}

func (a Decimal) Int64() int64 {
	return a.Decimal.IntPart()
}

// ValueInt - float to big int.
func ValueInt(value interface{}, decimals int32) *big.Int {

	amount := decimal.NewFromFloat(0)
	switch v := value.(type) {
	case string:
		amount, _ = decimal.NewFromString(v)
	case float64:
		amount = decimal.NewFromFloat(v)
	case int64:
		amount = decimal.NewFromFloat(float64(v))
	case decimal.Decimal:
		amount = v
	case *decimal.Decimal:
		amount = *v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	result := amount.Mul(mul)

	impost := new(big.Int)
	impost.SetString(result.String(), 10)

	return impost
}

// ValueFloat - big int to float.
func ValueFloat(value interface{}, decimals int32) float64 {

	impost := new(big.Int)
	switch v := value.(type) {
	case string:
		impost.SetString(v, 10)
	case *big.Int:
		impost = v
	}

	mul := decimal.NewFromFloat(float64(10)).Pow(decimal.NewFromFloat(float64(decimals)))
	num, _ := decimal.NewFromString(impost.String())
	result := num.Div(mul)

	if float, _ := result.Float64(); float > 0 {
		return float
	}

	return 0
}
