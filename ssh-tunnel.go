package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

// Server struct to hold server details
type Server struct {
	Host  string `yaml:"host"`
	Port  string `yaml:"port"`
	User  string `yaml:"user"`
	Proxy string `yaml:"proxy"` // socks5 or http
}

// Config struct to hold the configuration
type Config struct {
	Servers []Server `yaml:"servers"`
}

// TestResult holds latency test results
type TestResult struct {
	Server  Server
	Latency time.Duration
}

// testServer pings a server and measures latency
func testServer(server Server) (TestResult, error) {
	var totalLatency time.Duration
	var successfulPings int
	numPings := 3 // Number of pings to calculate average latency

	for i := 0; i < numPings; i++ {
		var cmd *exec.Cmd
		var re *regexp.Regexp

		if runtime.GOOS == "windows" {
			cmd = exec.Command("ping", "-n", "1", server.Host)
			re = regexp.MustCompile(`time[=<]?(\d+)ms`)
		} else {
			cmd = exec.Command("ping", "-c", "1", server.Host)
			re = regexp.MustCompile(`time[=<]([\d.]+) ms`)
		}

		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("Ping failed for %s: %v, output: %s", server.Host, err, string(output))
			continue
		}

		matches := re.FindStringSubmatch(string(output))
		if len(matches) < 2 {
			log.Printf("Failed to parse latency for %s: %s", server.Host, string(output))
			continue
		}
		latencyFloat, err := strconv.ParseFloat(matches[1], 64)
		if err != nil {
			log.Printf("Invalid latency value for %s: %v", server.Host, err)
			continue
		}

		totalLatency += time.Duration(latencyFloat * float64(time.Millisecond))
		successfulPings++
	}

	if successfulPings == 0 {
		return TestResult{}, fmt.Errorf("failed to ping server %s", server.Host)
	}

	averageLatency := totalLatency / time.Duration(successfulPings)
	return TestResult{Server: server, Latency: averageLatency}, nil
}

// startTunnel starts either SOCKS5 or HTTP proxy via SSH
func startTunnel(server Server) {
	for {
		var sshArgs []string

		if server.Proxy == "http" {
			sshArgs = []string{
				"-L", "8888:0.0.0.0:8888", fmt.Sprintf("%s@%s", server.User, server.Host), "-p", server.Port,
			}
		} else {
			sshArgs = []string{
				"-N", "-D", "0.0.0.0:8080", fmt.Sprintf("%s@%s", server.User, server.Host), "-p", server.Port,
			}
		}

		cmd := exec.Command("ssh", sshArgs...)
		log.Printf("Starting %s proxy on %s...", server.Proxy, server.Host)
		if err := cmd.Start(); err != nil {
			log.Printf("Failed to start SSH tunnel for %s: %v", server.Host, err)
			time.Sleep(5 * time.Second)
			continue
		}

		err := cmd.Wait()
		if err != nil {
			log.Printf("SSH tunnel to %s exited with error: %v", server.Host, err)
		} else {
			log.Printf("SSH tunnel to %s closed gracefully.", server.Host)
		}

		time.Sleep(5 * time.Second)
		log.Printf("Restarting tunnel to %s...", server.Host)
	}
}

func main() {
	currentDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory: %v", err)
	}

	configPath := filepath.Join(currentDir, "config.yaml")
	log.Printf("Using config file: %s", configPath)

	configFile, err := os.Open(configPath)
	if err != nil {
		log.Fatalf("Error opening config file: %v", err)
	}
	defer configFile.Close()

	var config Config
	decoder := yaml.NewDecoder(configFile)
	if err := decoder.Decode(&config); err != nil {
		log.Fatalf("Error decoding config file: %v", err)
	}

	if len(config.Servers) == 0 {
		log.Fatalf("No servers found in the configuration")
	}

	log.Println("Starting latency tests...")

	var results []TestResult
	for _, server := range config.Servers {
		result, err := testServer(server)
		if err != nil {
			log.Printf("Failed to test server %s: %v", server.Host, err)
			continue
		}
		log.Printf("Server %s responded in %v", server.Host, result.Latency)
		results = append(results, result)
	}

	if len(results) == 0 {
		log.Fatalf("No servers available")
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Latency < results[j].Latency
	})

	bestServer := results[0].Server
	log.Printf("Selected best server: %s with latency %v", bestServer.Host, results[0].Latency)

	go startTunnel(bestServer)
	select {}
}
