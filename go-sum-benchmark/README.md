# 🧮 Go Sum Benchmark

A simple Go benchmarking project comparing **sequential** vs **concurrent** computation of the **sum of squares** of a large dataset — designed to analyze performance impact of **goroutines and channels** even on a **single CPU**.

---

## 📂 Project Structure

```
go-sum-benchmark/
├── main.go              # Sequential and concurrent implementations
└── sum_bench_test.go    # Benchmark tests for both versions
```

---

## 🚀 Overview

This project demonstrates how concurrency in Go (using goroutines and channels) can be applied to a computational problem — summing the squares of integers — and how it performs compared to a plain sequential implementation.

Even when restricted to a **single CPU** (`runtime.GOMAXPROCS(1)`), the concurrent version may still show different behavior due to Go’s **scheduler overhead**, **context switching**, and **channel communication cost**.

---

## ⚙️ How It Works

### 1. **Sequential Version**

A simple loop computes the sum of squares:

```go
func sumSquaresSequential(data []int) int {
    sum := 0
    for _, v := range data {
        sum += v * v
    }
    return sum
}
```

### 2. **Concurrent Version**

The data is divided into chunks, and each chunk is processed by a separate goroutine:

```go
func sumSquaresConcurrent(data []int, workers int) int {
    chunkSize := (len(data) + workers - 1) / workers
    results := make(chan int, workers)

    for i := 0; i < workers; i++ {
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
    for i := 0; i < workers; i++ {
        total += <-results
    }
    close(results)
    return total
}
```

---

## 🧪 Running the Benchmark

### 1. **Run the main program**

```bash
go run ./go-sum-benchmark/main.go
```

Example output:

```
Sequential: sum=1666087121, took=16.58ms
Concurrent: sum=1666087121, took=8.18ms
```

### 2. **Run the benchmarks**

```bash
go test -bench=. -benchmem
```

### 3. **Compare results using benchstat**

```bash
go test -benchmem -run=^$ -bench=. ./go-sum-benchmark -race -count=6 -cpu=1,2 | tee go-sum-benchmark.txt
benchstat -col .name go-sum-benchmark.txt
```

---

## 📊 Benchmark Results

From `benchstat` (Darwin/arm64, single CPU):

| Benchmark        | sec/op   | B/op   | allocs/op |
|------------------|----------|--------|-----------|
| Sequential       | 16.24m   | 0      | 0         |
| Sequential-2     | 16.59m   | 0      | 0         |
| Concurrent       | 42.01m   | 672    | 17        |
| Concurrent-2     | 22.42m   | 672.5  | 17        |

**Key Insight:**  
Sequential execution is faster and allocation-free in this workload. Concurrency introduces overhead (allocations, memory, scheduling) and only pays off under certain conditions (larger workloads, multiple CPUs, or I/O-bound tasks).

---

## 📊 Key Takeaways

* The **concurrent version** may or may not outperform the sequential one, depending on:

  * CPU count (`GOMAXPROCS`)
  * Task size (granularity)
  * Communication and synchronization overhead

* Useful for understanding:

  * Goroutine scheduling
  * Channel-based communication costs
  * Benchmarking in Go (`testing.B`)
