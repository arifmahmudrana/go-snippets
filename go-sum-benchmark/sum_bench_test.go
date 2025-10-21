// go-sum-benchmark/sum_bench_test.go
package main

import (
	"math/rand"
	"testing"
)

var testData = func() []int {
	data := make([]int, 5_000_000)
	for i := range data {
		data[i] = rand.Intn(1000)
	}
	return data
}()

func BenchmarkSum(b *testing.B) {
	b.Run("Sequential", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sumSquaresSequential(testData)
		}
	})
	b.Run("Concurrent", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			sumSquaresConcurrent(testData, 8)
		}
	})
}
