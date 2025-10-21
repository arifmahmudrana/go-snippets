# Parallel Digits Counter

A Go program demonstrating **parallel text processing** using **goroutines**, **channels**, and **context cancellation**.

This example compares **sequential** vs **parallel** approaches for counting digit characters (`0–9`) in a text, showing how concurrency scales with CPU cores and how to manage context-aware worker pools efficiently.

---

## 🚀 Features

- ✅ Context-based worker cancellation (`context.Context`)
- ✅ Safe, memory-bounded worker pool pattern
- ✅ Graceful worker shutdown using `done` channel
- ✅ Merging partial results concurrently
- ✅ Unit tests & benchmarks (`go test -v -bench . -benchmem`)
- ✅ Supports dynamic CPU scaling (`runtime.GOMAXPROCS`)

---

## 🧠 How It Works

The program splits input text into words, then sends them to a pool of worker goroutines.  
Each worker counts digits in its assigned word and sends partial counts back via a results channel.  
Finally, all results are merged into a single aggregated map.

### Parallel Flow
```

words → tasks → [worker goroutines] → results → merger → final count

````

Each worker:
- Listens for tasks on the `tasks` channel.
- Counts digit characters.
- Sends back a `map[rune]int` result.
- Terminates gracefully when the context is cancelled or tasks are finished.

---

## 🧩 Example Output

```bash
$ go run parallel_digits/parallel_digits.go

Final counts:
'0' => 3
'1' => 4
'2' => 2
'3' => 2
'4' => 1
'5' => 1
'6' => 1
'7' => 2
'8' => 1
'9' => 1
````

---

## 🧪 Running Tests

```bash
# Run all tests
go test ./parallel_digits -v

# Run benchmarks with memory stats
go test ./parallel_digits -bench . -benchmem
```

To compare sequential vs concurrent performance:

```bash
go test -benchmem -run=^$ -bench ^BenchmarkSequential ./parallel_digits -race -count=6 -cpu=1,2 | tee seq.txt
go test -benchmem -run=^$ -bench ^BenchmarkParallel ./parallel_digits -race -count=6 -cpu=1,2 | tee parallel.txt
```

Then use [`benchstat`](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat):

```bash
benchstat seq.txt parallel.txt
```

---

## 📊 Performance Results

Benchmarks comparing sequential and parallel versions were run on an Apple M3 Pro (ARM64, macOS).

| Benchmark | sec/op | allocs/op | B/op | Notes |
|------------|--------|------------|------|-------|
| Sequential | 1.020 µs | 1 | 227 B | Simple single-threaded |
| Parallel | 33.99 µs | 55 | 3.3 KiB | Overhead from goroutines |
| SequentialLargeInput | 350.7 µs | 1 | 224 B | Linear scaling |
| ParallelLargeInput | 2.86 ms | 5033 | 424 KiB | Channel and map allocation overhead |

🧩 **Conclusion:**
- Sequential execution is faster for small datasets.
- Parallel execution only outperforms for **very large workloads** or **I/O-bound operations**.
- Use concurrency for scalability or responsiveness, not for trivial CPU-bound loops.

---

## ⚙️ Environment Variables

You can limit CPUs used by Go runtime:

```bash
GOMAXPROCS=1 go run parallel_digits/parallel_digits.go
```

This is useful to study how concurrency behaves under CPU constraints.

---

## 📂 Project Structure

```
parallel_digits/
├── parallel_digits.go          # main program
└── parallel_digits_test.go     # tests & benchmarks
```

---

## 🧭 Key Learnings

* Concurrency is **not always faster** on single-core systems.
* Proper use of channels prevents memory leaks.
* Context helps avoid deadlocks and resource starvation.
* Controlled worker pools provide safe and scalable parallelism.
