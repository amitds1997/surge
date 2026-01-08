// grab_bench.go - Simple benchmark helper using the grab library
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/cavaliergopher/grab/v3"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: grab_bench <url> <output_dir>\n")
		os.Exit(1)
	}

	url := os.Args[1]
	outputDir := os.Args[2]

	client := grab.NewClient()
	client.UserAgent = "SurgeBenchmark/1.0"
	req, err := grab.NewRequest(outputDir, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		os.Exit(1)
	}
	// Disable resume to avoid potential HEAD/Range request issues on some servers
	req.NoResume = true

	start := time.Now()
	resp := client.Do(req)

	// Wait for download to complete
	<-resp.Done

	elapsed := time.Since(start)

	if err := resp.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
		os.Exit(1)
	}

	// Output stats in a parseable format
	fmt.Printf("file=%s\n", resp.Filename)
	fmt.Printf("size=%d\n", resp.Size())
	fmt.Printf("elapsed_ms=%d\n", elapsed.Milliseconds())
	fmt.Printf("speed_mbps=%.2f\n", float64(resp.Size())/(1024*1024)/elapsed.Seconds())
}
