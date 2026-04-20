package fibbench

import "testing"

const fibDepth = 20

var fibSink int

func fibonacciRecursive(n int) int {
	if n <= 1 {
		return n
	}
	return fibonacciRecursive(n-1) + fibonacciRecursive(n-2)
}

func fibonacciIterative(n int) int {
	if n <= 1 {
		return n
	}
	prev, cur := 0, 1
	for i := 2; i <= n; i++ {
		prev, cur = cur, prev+cur
	}
	return cur
}

func BenchmarkFibonacciDepth20Recursive(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		fibSink = fibonacciRecursive(fibDepth)
	}
}

func BenchmarkFibonacciDepth20Iterative(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		fibSink = fibonacciIterative(fibDepth)
	}
}
