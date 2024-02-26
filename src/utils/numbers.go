package utils

import "math"

func Percent(n int64, p float64) int64 {
	return int64(float64(n) * p)
}

func PowInt64(x, y int64) int64 {
	return int64(math.Pow(float64(x), float64(y)))
}
