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
	suppressEmpty = flag.Bool("suppressEmpty", true, "Suppress printing empty metrics")
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

	if skipEmpty(ss) {
		return
	}

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
	fmt.Println()
}

type emptyState int

const (
	nothingPrinted emptyState = iota
	emptyPrinted
	statsPrinted
)

var state emptyState

func skipEmpty(ss *statsd.Snapshot) bool {
	if len(ss.Counters)+len(ss.Gauges)+len(ss.Timers) == 0 {
		if state != emptyPrinted {
			fmt.Printf("Metrics: no new metrics.")
		} else {
			fmt.Printf(".")
		}
		state = emptyPrinted
		return true
	}

	if state == emptyPrinted {
		fmt.Println()
		fmt.Println()
	}
	state = statsPrinted
	return false
}
