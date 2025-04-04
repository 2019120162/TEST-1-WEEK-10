# TEST-1-WEEK-10

# Purpose of this program

Purpose: This program demonstrates a concurrent TCP network port scanner that can:
- Scan a range of ports or specific ports on a target
- Grab banners for open ports
- Output results in both human-readable and JSON formats
- Use concurrent workers for efficient scanning

## Prerequisites

### Building the Tool

Clone the repository to your local machine:

git clone https://github.com/2019120162/TEST-1-WEEK-10.git

### Command-Line Flags:

-target: Specifies the target IP or hostname.

-start-port and -end-port: Define the range of ports to scan (default: 1 to 1024).

-workers: Number of concurrent workers for scanning (default: 100).

-timeout: Timeout for each connection in seconds (default: 5).

-banner: Enables banner grabbing (default: false).

-json: Outputs results in JSON format (default: false).

-ports: Allows scanning specific ports (optional).


# EXAMPLE USAGE

Scan a Range of Ports with Multiple Workers:
go run main.go -target=example.com -start-port=20 -end-port=80 -workers=50

Scan Specific Ports and Get JSON Output:
go run main.go -target=example.com -ports=22,80,443 -json=true

Enable Banner Grabbing:
go run main.go -target=example.com -banner=true
