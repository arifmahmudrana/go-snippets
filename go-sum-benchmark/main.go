// go-sum-benchmark/main.go
package main

import (
	"fmt"
	"math/rand"
	"runtime"
	"time"
)

func generateData(n int) []int {
	data := make([]int, n)
	for i := range data {
		data[i] = rand.Intn(1000)
	}
	return data
}

// Sequential: sum of squares
func sumSquaresSequential(data []int) int {
	sum := 0
	for _, v := range data {
		sum += v * v
	}
	return sum
}

// Concurrent: divide work into chunks and sum in goroutines
func sumSquaresConcurrent(data []int, workers int) int {
	chunkSize := (len(data) + workers - 1) / workers
	results := make(chan int, workers)

	for i := range workers {
		start := i * chunkSize
		end := min(start+chunkSize, len(data))

		go func(chunk []int) {
			sum := 0
			for _, v := range chunk {
				sum += v * v
			}
			results <- sum
		}(data[start:end])
	}

	total := 0
	for range workers {
		total += <-results
	}
	close(results)
	return total
}

func main() {
	runtime.GOMAXPROCS(1) // force single-CPU execution

	data := generateData(5_000_000)

	start := time.Now()
	s1 := sumSquaresSequential(data)
	fmt.Printf("Sequential: sum=%d, took=%v\n", s1, time.Since(start))

	start = time.Now()
	s2 := sumSquaresConcurrent(data, 8)
	fmt.Printf("Concurrent: sum=%d, took=%v\n", s2, time.Since(start))
}
