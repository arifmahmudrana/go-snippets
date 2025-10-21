// parallel_digits/parallel_digits.go
package main

import (
	"context"
	"fmt"
	"runtime"
	"slices"
	"strings"
	"time"
)

// worker processes words from tasks channel and sends digit counts to results.
// It respects context cancellation and signals completion via done channel.
func worker(ctx context.Context, tasks <-chan string, results chan<- map[rune]int, done chan<- struct{}) {
	defer func() { done <- struct{}{} }() // signal when this worker exits

	for {
		select {
		case <-ctx.Done():
			return
		case w, ok := <-tasks:
			if !ok {
				// tasks closed -> normal exit
				return
			}
			// process word: count digits
			m := make(map[rune]int)
			for _, r := range w {
				if r >= '0' && r <= '9' {
					m[r]++
				}
			}
			// non-blocking send: respect ctx cancellation
			select {
			case <-ctx.Done():
				return
			case results <- m:
			}
		}
	}
}

// mergeResults collects all digit count maps from results channel and
// merges them into a single map. Returns when results channel is closed
// or context is cancelled.
func mergeResults(ctx context.Context, results <-chan map[rune]int) map[rune]int {
	final := make(map[rune]int)
	for {
		select {
		case <-ctx.Done():
			return final
		case m, ok := <-results:
			if !ok {
				return final
			}
			for k, v := range m {
				final[k] += v
			}
		}
	}
}

// countDigitsParallel counts digit occurrences in words using worker pool pattern.
// Returns a map of digit rune to count.
//
// Memory efficiency:
// - tasks channel: small buffer (numWorkers) for bounded memory
// - results channel: matches worker count for optimal throughput
// - words are streamed, not all loaded into channel at once
func countDigitsParallel(ctx context.Context, words []string, numWorkers int) map[rune]int {
	// Small buffers: memory-efficient, stream-based processing
	tasks := make(chan string, numWorkers)         // only buffer what workers can handle
	results := make(chan map[rune]int, numWorkers) // one slot per worker
	done := make(chan struct{}, numWorkers)        // buffered, workers never block

	// start workers
	for range numWorkers {
		go worker(ctx, tasks, results, done)
	}

	// producer: stream tasks (non-blocking with context)
	go func() {
		defer close(tasks) // signal no more work when done
		for _, w := range words {
			select {
			case <-ctx.Done():
				return // stop producing if cancelled
			case tasks <- w:
				// send task (blocks if buffer full, that's OK - backpressure)
			}
		}
	}()

	// coordinator: wait for all workers, then close results
	go func() {
		for range numWorkers {
			<-done
		}
		close(results) // signal merger that no more results coming
	}()

	return mergeResults(ctx, results)
}

// printSortedCounts prints digit counts in sorted order (0-9) for consistent output.
func printSortedCounts(counts map[rune]int) {
	// extract keys and sort
	keys := make([]rune, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	// print in sorted order
	fmt.Println("Final counts:")
	for _, k := range keys {
		fmt.Printf("%q => %d\n", k, counts[k])
	}
}

func main() {
	text := "1I12 1l0v3 Y!!07 something 123 45 67 890"
	words := strings.Fields(text)

	// Use GOMAXPROCS instead of NumCPU for container-aware parallelism
	// GOMAXPROCS respects cgroup CPU limits (e.g., docker --cpus=2)
	// NumCPU returns physical cores (ignores container quotas)
	maxWorkers := max(runtime.GOMAXPROCS(0), 1)

	// context with timeout to prevent hangs
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// count digits in parallel
	final := countDigitsParallel(ctx, words, maxWorkers)

	// print sorted results (deterministic output)
	printSortedCounts(final)
}
