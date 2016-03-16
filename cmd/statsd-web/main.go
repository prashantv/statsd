package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prashantv/statsd"
)

var (
	host     = flag.String("host", "127.0.0.1", "Host to listen on")
	port     = flag.Int("port", 8125, "UDP port to listen on for statsd")
	httpAddr = flag.String("http", ":8080", "HTTP address for web UI")
	window   = flag.Duration("window", time.Second, "Window to aggregate metrics by")
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

	go startServer()
	for {
		time.Sleep(*window)
		recordMetrics(time.Now().Truncate(*window), m)
	}
}

func startServer() {
	http.HandleFunc("/json", jsonHandler)
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(*httpAddr, nil)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, indexTmpl)
}

func jsonHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	start := 0
	max := 100

	if parsed, err := strconv.Atoi(r.FormValue("from")); err == nil {
		start = parsed
	}
	if parsed, err := strconv.Atoi(r.FormValue("max")); err == nil {
		max = parsed
	}

	snapshotsLock.RLock()
	if start > len(snapshots) || start < 0 {
		log.Printf("Move start %v to be within range [0, %v]", start, len(snapshots))
		start = 0
	}
	toRender := snapshots[start:]
	if len(toRender) > max {
		toRender = toRender[len(toRender)-max:]
	}
	snapshotsLock.RUnlock()

	json.NewEncoder(w).Encode(toRender)
}
