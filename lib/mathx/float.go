package mathx

import (
	"fmt"
	"math"
	"strconv"
)

const MIN = 0.000001

// IsEqual
// 判断 float 是否相等
func IsEqual(x, y float64) bool {
	return math.Abs(x-y) < MIN
}

func Max(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}
	first := arr[0]
	if len(arr) == 1 {
		return first
	}
	args := arr[1:]
	for _, v := range args {
		first = math.Max(first, v)
	}
	return first
}

func Min(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}
	first := arr[0]
	if len(arr) == 1 {
		return first
	}
	args := arr[1:]
	for _, v := range args {
		first = math.Min(first, v)
	}
	return first
}

func Avg(arr []float64) float64 {
	if len(arr) == 0 {
		return 0
	}
	first := arr[0]
	if len(arr) == 1 {
		return first
	}
	args := arr[1:]
	for _, v := range args {
		first += v
	}
	μ := first / float64(len(args)+1)
	return μ
}

func StdDev(arr []float64) float64 {
	if len(arr) <= 1 {
		return 0
	}
	μ := Avg(arr)
	var variance float64
	for _, v := range arr {
		variance += math.Pow((v - μ), 2)
	}
	σ := math.Sqrt(variance / float64(len(arr)))

	return σ
}

// 保留两位小数
func Float2Decimal(value float64) float64 {
	value, _ = strconv.ParseFloat(fmt.Sprintf("%.2f", value), 64)
	return value
}
