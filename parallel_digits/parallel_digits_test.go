// parallel_digits/parallel_digits_test.go
package main

import (
	"context"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"
)

// TestCountDigitsParallel_Basic tests basic functionality
func TestCountDigitsParallel_Basic(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[rune]int
	}{
		{
			name:  "simple text",
			input: "1I12 1l0v3 Y!!07",
			want:  map[rune]int{'0': 2, '1': 3, '2': 1, '3': 1, '7': 1},
		},
		{
			name:  "no digits",
			input: "hello world",
			want:  map[rune]int{},
		},
		{
			name:  "only digits",
			input: "123 456 789",
			want:  map[rune]int{'1': 1, '2': 1, '3': 1, '4': 1, '5': 1, '6': 1, '7': 1, '8': 1, '9': 1},
		},
		{
			name:  "repeated digits",
			input: "111 222 333",
			want:  map[rune]int{'1': 3, '2': 3, '3': 3},
		},
		{
			name:  "empty string",
			input: "",
			want:  map[rune]int{},
		},
		{
			name:  "all digits 0-9",
			input: "0123456789",
			want:  map[rune]int{'0': 1, '1': 1, '2': 1, '3': 1, '4': 1, '5': 1, '6': 1, '7': 1, '8': 1, '9': 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			words := strings.Fields(tt.input)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			got := countDigitsParallel(ctx, words, runtime.NumCPU())

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("countDigitsParallel() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCountDigitsParallel_LargeInput tests with larger input
func TestCountDigitsParallel_LargeInput(t *testing.T) {
	// generate large input: repeat "word123" 1000 times
	words := make([]string, 1000)
	for i := range words {
		words[i] = "word123"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	got := countDigitsParallel(ctx, words, runtime.NumCPU())
	want := map[rune]int{'1': 1000, '2': 1000, '3': 1000}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("countDigitsParallel() = %v, want %v", got, want)
	}
}

// TestCountDigitsParallel_ContextCancellation tests cancellation behavior
func TestCountDigitsParallel_ContextCancellation(t *testing.T) {
	// create a lot of work
	words := make([]string, 10000)
	for i := range words {
		words[i] = "test123456789"
	}

	// context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	got := countDigitsParallel(ctx, words, runtime.NumCPU())

	// result should be partial or empty (workers exit early)
	// just verify it doesn't hang or panic
	t.Logf("Got %d digit types with cancelled context", len(got))
}

// TestCountDigitsParallel_DifferentWorkerCounts tests various worker counts
func TestCountDigitsParallel_DifferentWorkerCounts(t *testing.T) {
	input := "1I12 1l0v3 Y!!07 something 123 45 67 890"
	words := strings.Fields(input)
	want := map[rune]int{'0': 3, '1': 4, '2': 2, '3': 2, '4': 1, '5': 1, '6': 1, '7': 2, '8': 1, '9': 1}

	workerCounts := []int{1, 2, 4, 8, 16}

	for _, numWorkers := range workerCounts {
		t.Run(string(rune(numWorkers+'0'))+" workers", func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			got := countDigitsParallel(ctx, words, numWorkers)

			if !reflect.DeepEqual(got, want) {
				t.Errorf("with %d workers: got %v, want %v", numWorkers, got, want)
			}
		})
	}
}

// TestCountDigitsParallel_Unicode tests that non-digit unicode is ignored
func TestCountDigitsParallel_Unicode(t *testing.T) {
	input := "hello世界123مرحبا456שלום789"
	words := strings.Fields(input)
	want := map[rune]int{'1': 1, '2': 1, '3': 1, '4': 1, '5': 1, '6': 1, '7': 1, '8': 1, '9': 1}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	got := countDigitsParallel(ctx, words, runtime.NumCPU())

	if !reflect.DeepEqual(got, want) {
		t.Errorf("countDigitsParallel() = %v, want %v", got, want)
	}
}

// BenchmarkSequential benchmarks sequential processing
func BenchmarkSequential(b *testing.B) {
	text := "1I12 1l0v3 Y!!07 something 123 45 67 890"
	words := strings.Fields(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		final := make(map[rune]int)
		for _, w := range words {
			for _, r := range w {
				if r >= '0' && r <= '9' {
					final[r]++
				}
			}
		}
	}
}

// BenchmarkParallel benchmarks parallel processing
func BenchmarkParallel(b *testing.B) {
	text := "1I12 1l0v3 Y!!07 something 123 45 67 890"
	words := strings.Fields(text)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = countDigitsParallel(ctx, words, runtime.NumCPU())
		cancel()
	}
}

// BenchmarkParallelLargeInput benchmarks with larger input
func BenchmarkParallelLargeInput(b *testing.B) {
	// generate 1000 words
	words := make([]string, 1000)
	for i := range words {
		words[i] = "test123456789word"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = countDigitsParallel(ctx, words, runtime.NumCPU())
		cancel()
	}
}

// BenchmarkSequentialLargeInput benchmarks sequential with larger input
func BenchmarkSequentialLargeInput(b *testing.B) {
	words := make([]string, 1000)
	for i := range words {
		words[i] = "test123456789word"
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		final := make(map[rune]int)
		for _, w := range words {
			for _, r := range w {
				if r >= '0' && r <= '9' {
					final[r]++
				}
			}
		}
	}
}

// BenchmarkParallelVaryingWorkers benchmarks different worker counts
func BenchmarkParallelVaryingWorkers(b *testing.B) {
	words := make([]string, 1000)
	for i := range words {
		words[i] = "test123456789word"
	}

	workerCounts := []int{1, 2, 4, 8, 16}

	for _, numWorkers := range workerCounts {
		b.Run(string(rune(numWorkers+'0'))+" workers", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_ = countDigitsParallel(ctx, words, numWorkers)
				cancel()
			}
		})
	}
}
