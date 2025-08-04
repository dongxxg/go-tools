package tests

import "testing"

// @Author dx
// @Date 2025-08-04 22:27:00
// @Desc

func fib(n int) int {
	if n == 0 || n == 1 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func BenchmarkFib(b *testing.B) {
	for i := 0; i < b.N; i++ {
		fib(30)
	}
}

// PS E:\goproject\go-tools\tests> go test -bench .
//goos: windows
//goarch: amd64
//pkg: go-tools/tests
//cpu: Intel(R) Core(TM) i5-9600KF CPU @ 3.70GHz
//BenchmarkFib-6               342           3523009 ns/op
//PASS
//ok      go-tools/tests  1.738s

//-6 即 GOMAXPROCS，默认等于 CPU 核数
