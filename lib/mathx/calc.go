package mathx

import (
	"errors"
	"fmt"
	"strconv"
)

func CalcRatioStr(use, total string) (string, error) {
	r, err := CalcRatioFloat64(use, total)
	if err != nil {
		return "", err
	}

	rs := fmt.Sprintf("%.2f", r)

	if rs == "+Inf" || rs == "-Inf" {
		return "", errors.New("divide by zero")
	}

	return rs, nil
}

func CalcRatioInt(use, total string) (int, error) {
	var r int
	u, err := strconv.Atoi(use)
	if err != nil {
		return r, err
	}
	t, err := strconv.Atoi(total)
	if err != nil {
		return r, err
	}
	if t == 0 {
		return 0, errors.New("divide by zero")
	}

	return u * 100 / t, nil
}

func CalcRatioInt64(use, total string) (int64, error) {
	var r int64
	u, err := strconv.ParseInt(use, 10, 64)
	if err != nil {
		return r, err
	}
	t, err := strconv.ParseInt(total, 10, 64)
	if err != nil {
		return r, err
	}
	if t == 0 {
		return 0, errors.New("divide by zero")
	}

	return u * 100 / t, nil
}

func CalcRatioFloat64(use, total string) (float64, error) {
	var r float64
	u, err := strconv.ParseFloat(use, 64)
	if err != nil {
		return r, err
	}
	t, err := strconv.ParseFloat(total, 64)
	if err != nil {
		return r, err
	}

	return u * 100 / t, nil
}

func CalcAvgStr(items ...string) string {
	return fmt.Sprintf("%.2f", CalcAvgFloat64(items...))
}

func CalcAvgFloat64(items ...string) float64 {
	var (
		cnt int
		sum float64
	)
	for _, item := range items {
		f, err := strconv.ParseFloat(item, 64)
		if err == nil {
			cnt++
			sum += f
		}
	}
	if cnt == 0 {
		return 0.00
	}

	return sum / float64(cnt)
}

func CalcDivStr(total, use string) (string, error) {
	r, err := CalcDiv(total, use)
	if err != nil {
		return "", err
	}
	rs := fmt.Sprintf("%.2f", r)

	if rs == "+Inf" || rs == "-Inf" {
		return "", errors.New("divide by zero")
	}

	return rs, nil
}

func CalcDiv(total, use string) (float64, error) {
	var r float64
	u, err := strconv.ParseFloat(use, 64)
	if err != nil {
		return r, err
	}
	t, err := strconv.ParseFloat(total, 64)
	if err != nil {
		return r, err
	}

	return t / u, nil
}

func CalcMinStr(items ...string) string {
	return fmt.Sprintf("%.2f", CalcMinFloat64(items...))
}

func CalcMaxStr(items ...string) string {
	return fmt.Sprintf("%.2f", CalcMaxFloat64(items...))
}

func CalcMinFloat64(items ...string) float64 {
	var min float64
	var af bool
	for _, item := range items {
		f, err := strconv.ParseFloat(item, 64)
		if err != nil {
			if af {
				if f < min {
					min = f
				}
			} else {
				min = f
				af = true
			}
		}
	}
	return min
}

func CalcMaxFloat64(items ...string) float64 {
	var max float64
	var af bool
	for _, item := range items {
		f, err := strconv.ParseFloat(item, 64)
		if err != nil {
			if af {
				if f > max {
					max = f
				}
			} else {
				max = f
				af = true
			}
		}
	}
	return max
}

func CalcIsInRange(i, a, b string) bool {
	f1, err := strconv.ParseFloat(i, 64)
	if err != nil {
		return false
	}
	f2, err := strconv.ParseFloat(a, 64)
	if err != nil {
		return false
	}
	f3, err := strconv.ParseFloat(b, 64)
	if err != nil {
		return false
	}

	if f1 >= f2 && f1 <= f3 {
		return true
	}

	return false
}
