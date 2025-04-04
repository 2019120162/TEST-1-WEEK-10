// Filename: main.go
// Purpose: This program demonstrates how to create a TCP network connection using Go and added enhancements like configurable flags, concurrency, banner grabbing, etc.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Struct to hold scan results for output
type ScanResult struct {
	Target string `json:"target"`
	Port   int    `json:"port"`
	IsOpen bool   `json:"is_open"`
	Banner string `json:"banner,omitempty"`
}

func worker(wg *sync.WaitGroup, tasks chan string, dialer net.Dialer, timeout time.Duration, results chan<- ScanResult, bannerGrab bool) {
	defer wg.Done()

	for addr := range tasks {
		conn, err := dialer.Dial("tcp", addr)
		if err == nil {
			// Immediately close the connection
			defer conn.Close()

			result := ScanResult{
				Target: addr,
				Port:   extractPortFromAddr(addr),
				IsOpen: true,
			}

			// If banner grabbing is enabled, try to read the banner from the server
			if bannerGrab {
				conn.SetReadDeadline(time.Now().Add(timeout)) // Set read deadline based on timeout flag
				banner := make([]byte, 1024)                  // Buffer to store the banner
				n, err := conn.Read(banner)
				if err == nil || err.Error() == "EOF" {
					result.Banner = string(banner[:n]) // Store the banner if successfully read
				}
			}

			// Send the result to the results channel
			results <- result
		}
	}
}

func extractPortFromAddr(addr string) int {
	_, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)
	return port
}

func main() {
	// Define flags for configurable options
	target := flag.String("target", "scanme.nmap.org", "Target IP or hostname")
	startPort := flag.Int("start-port", 1, "Start port (default: 1)")
	endPort := flag.Int("end-port", 1024, "End port (default: 1024)")
	workers := flag.Int("workers", 100, "Number of concurrent workers (default: 100)")
	timeout := flag.Int("timeout", 5, "Timeout for each connection in seconds (default: 5)")
	bannerGrab := flag.Bool("banner", false, "Enable banner grabbing (default: false)")
	jsonOutput := flag.Bool("json", false, "Output in JSON format (default: false)")
	ports := flag.String("ports", "", "Comma-separated list of specific ports to scan (optional)")

	// Parse the flags
	flag.Parse()

	// Start measuring time
	startTime := time.Now()

	// Prepare the list of ports to scan
	var scanPorts []int
	if *ports != "" {
		// Parse specific ports from the comma-separated list
		for _, portStr := range strings.Split(*ports, ",") {
			port, err := strconv.Atoi(portStr)
			if err != nil {
				fmt.Printf("Invalid port %s\n", portStr)
				os.Exit(1)
			}
			scanPorts = append(scanPorts, port)
		}
	} else {
		// Default port range if no specific ports are provided
		for p := *startPort; p <= *endPort; p++ {
			scanPorts = append(scanPorts, p)
		}
	}

	// Dialer with custom timeout
	dialer := net.Dialer{
		Timeout: time.Duration(*timeout) * time.Second,
	}

	// Channel for tasks and results
	tasks := make(chan string, 100)
	results := make(chan ScanResult, len(scanPorts))

	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 1; i <= *workers; i++ {
		wg.Add(1)
		go worker(&wg, tasks, dialer, time.Duration(*timeout)*time.Second, results, *bannerGrab)
	}

	// Scan the target and ports
	for _, port := range scanPorts {
		portStr := strconv.Itoa(port)
		address := net.JoinHostPort(*target, portStr)
		tasks <- address
	}

	// Close tasks channel once all tasks are enqueued
	close(tasks)

	// Wait for all workers to finish
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	var openPorts []ScanResult
	for result := range results {
		if result.IsOpen {
			openPorts = append(openPorts, result)
		}
	}

	// Calculate the time taken for the scan
	duration := time.Since(startTime)

	// Output the scan summary
	if *jsonOutput {
		// JSON output format
		jsonData, err := json.MarshalIndent(openPorts, "", "  ")
		if err != nil {
			fmt.Println("Error generating JSON output:", err)
			return
		}
		fmt.Println(string(jsonData))
	} else {
		// Regular output format
		fmt.Printf("Scan Summary:\n")
		fmt.Printf("Target: %s\n", *target)
		fmt.Printf("Total ports scanned: %d\n", len(scanPorts))
		fmt.Printf("Open ports (%d):\n", len(openPorts))
		for _, result := range openPorts {
			fmt.Printf("Port: %d is open\n", result.Port)
			if result.Banner != "" {
				fmt.Printf("Banner: %s\n", result.Banner)
			}
		}
	}

	// Print the time taken for the scan
	fmt.Printf("\nTime taken: %v\n", duration)
}
