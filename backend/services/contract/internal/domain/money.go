package domain

import (
	"fmt"
	"math"
)

const moneyScale = 100.0
const moneyEpsilon = 1e-6

func MoneyToMinorUnits(v float64, field string) (int64, error) {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("%s must be a valid number", field)
	}
	scaled := v * moneyScale
	rounded := math.Round(scaled)
	if math.Abs(scaled-rounded) > moneyEpsilon {
		return 0, fmt.Errorf("%s must have at most 2 decimal places", field)
	}
	if rounded > math.MaxInt64 || rounded < math.MinInt64 {
		return 0, fmt.Errorf("%s is out of range", field)
	}
	return int64(rounded), nil
}

func MinorUnitsToMoney(v int64) float64 {
	return float64(v) / moneyScale
}
