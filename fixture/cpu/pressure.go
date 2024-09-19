package main

import (
	"fmt"
	"math"
	"os"
	"runtime"
	"time"
)

// IsPrime checks if a number is prime
func IsPrime(n int) bool {
	if n <= 1 {
		return false
	}
	for i := 2; i <= int(math.Sqrt(float64(n))); i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// CPUIntensiveTask runs a CPU-intensive task (calculating primes)
func CPUIntensiveTask(limit int) int {
	primeCount := 0
	for i := 2; i <= limit; i++ {
		if IsPrime(i) {
			primeCount++
		}
	}
	return primeCount
}

// MeasureTime is similar to Node.js's console.time/console.timeEnd
func MeasureTime(label string, fn func()) {
	start := time.Now()
	fmt.Printf("Starting: %s\n", label)
	fn()
	duration := time.Since(start)
	fmt.Printf("Finished: %s in %v\n", label, duration)
}

func main() {
	// Set maximum CPUs to use
	numCPU := runtime.NumCPU()
	runtime.GOMAXPROCS(numCPU)
	fmt.Printf("Using %d CPUs\n", numCPU)

	// Set a high limit for prime number calculation to simulate high CPU usage
	limit := 1000000
	if len(os.Args) > 1 {
		fmt.Sscanf(os.Args[1], "%d", &limit)
	}

	// Run the CPU-intensive task and measure its time
	MeasureTime("Prime Calculation", func() {
		primeCount := CPUIntensiveTask(limit)
		fmt.Printf("Found %d prime numbers up to %d\n", primeCount, limit)
	})
}
