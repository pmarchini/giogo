// Fixture that creates a temporary directory containing a 5MB file
// Then the file is read in chunks of 1MB and the number of bytes read is printed

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// MeasureTime is similar to Node.js's console.time/console.timeEnd
func MeasureTime(label string, fn func()) {
	start := time.Now()
	fmt.Printf("Starting: %s\n", label)
	fn()
	duration := time.Since(start)
	fmt.Printf("Finished: %s in %v\n", label, duration)
}

func main() {
	// Create a temporary directory
	tempDir, err := ioutil.TempDir("", "io-reading")
	if err != nil {
		fmt.Printf("Failed to create temp directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(tempDir)

	MeasureTime("Read File", func() {
		// Create a 5MB file
		filePath := filepath.Join(tempDir, "file")
		if err := ioutil.WriteFile(filePath, make([]byte, 5*1024*1024), 0644); err != nil {
			fmt.Printf("Failed to create file: %v\n", err)
			os.Exit(1)
		}

		// Open the file for reading
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("Failed to open file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()

		// Read the whole file into a buffer
		buf, err := ioutil.ReadAll(file)
		if err != nil {
			fmt.Printf("Failed to read file: %v\n", err)
			os.Exit(1)
		}
		// Print the number of bytes read
		fmt.Printf("Read %d bytes\n", len(buf))
	})
}
