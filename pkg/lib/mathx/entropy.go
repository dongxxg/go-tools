package mathx

import "math"

const epsilon = 1e-6

// 计算信息熵。变量的不确定性越大，熵也就越大。
func CalcEntropy(m map[interface{}]int) float64 {
	if len(m) == 0 || len(m) == 1 {
		return 1
	}

	var entropy float64
	var total int
	for _, v := range m {
		total += v
	}

	// H(x) = -∑P(xi)log(2,P(xi)) (i=1,2,..n)
	for _, v := range m {
		proba := float64(v) / float64(total)
		if proba < epsilon {
			proba = epsilon
		}
		entropy -= proba * math.Log2(proba)
	}

	// 换底
	return entropy / math.Log2(float64(len(m)))
}
