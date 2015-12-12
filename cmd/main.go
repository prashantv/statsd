package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/prashantv/statsd"
)

var (
	host          = flag.String("host", "127.0.0.1", "Host to listen on")
	port          = flag.Int("port", 8125, "Port to listen on")
	flushDuration = flag.Duration("flushDuration", 10*time.Second, "Duration to flush stats")
)

const maxPacketSize = 65536

func main() {
	flag.Parse()
	hostPort := fmt.Sprintf("%s:%v", *host, *port)

	m, addr, err := statsd.Start(hostPort)
	if err != nil {
		log.Fatalf("Failed to start statsd server: %v", err)
	}
	fmt.Println("Started statsd server on", addr.String())

	for {
		time.Sleep(*flushDuration)
		printMetrics(m)
	}
}

func printMetrics(m *statsd.Metrics) {
	ss := m.FlushAndSnapshot()

	fmt.Println("Metrics:")
	fmt.Printf("  Counters (%v)\n", len(ss.Counters))
	for k, v := range ss.Counters {
		fmt.Printf("    %15v: %6d\n", k, v)
	}
	fmt.Printf("  Gauges   (%v)\n", len(ss.Gauges))
	for k, v := range ss.Gauges {
		fmt.Printf("    %15v: %6d\n", k, v)
	}

	fmt.Printf("  Timers   (%v)\n", len(ss.Timers))
	for k, v := range ss.Timers {
		fmt.Printf("    %15v: %v\n", k, v)
	}
}
