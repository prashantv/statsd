package main

import (
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	host = flag.String("host", "127.0.0.1", "Host to listen on")
	port = flag.Int("port", 8125, "Port to listen on")
)

const maxPacketSize = 65536

func main() {
	flag.Parse()
	hostPort := fmt.Sprintf("%s:%v", *host, *port)

	if err := startStatsd(hostPort); err != nil {
		log.Fatalf("Failed to start statsd server: %v", err)
	}

	fmt.Println("Started statsd server on", hostPort)
	select {}
}

// Metrics represents all the metrics collected by this statsd server.
type Metrics struct {
	sync.RWMutex

	counters map[string]int64
	gauges   map[string]int64
	timers   map[string][]timer
}

type timer struct {
	time  time.Time
	value float64
}

func (t timer) String() string {
	return fmt.Sprint(t.value)
}

func startStatsd(hostPort string) error {
	addr, err := net.ResolveUDPAddr("udp", hostPort)
	if err != nil {
		return err
	}

	udpConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}

	m := newMetrics()
	go m.handleConn(udpConn)
	go m.printEvery(time.Second)
	return nil
}

func newMetrics() *Metrics {
	return &Metrics{
		counters: make(map[string]int64),
		gauges:   make(map[string]int64),
		timers:   make(map[string][]timer),
	}
}

func (m *Metrics) print() {
	m.RLock()
	defer m.RUnlock()

	fmt.Printf("Metrics:\n  counters: %v\n  gauges: %v\n  timers: %+v\n",
		m.counters, m.gauges, m.timers)
}

func (m *Metrics) printEvery(d time.Duration) {
	for {
		time.Sleep(d)
		m.print()
	}
}

func (m *Metrics) handleConn(conn *net.UDPConn) {
	buf := make([]byte, maxPacketSize)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Fatalf("read failed: %v", err)
		}

		if err := m.processPacket(buf[:n]); err != nil {
			log.Printf("Failed to process packet %s: %v", string(buf[:n]), err)
		}
	}
}

func (m *Metrics) processCounter(name, value []byte) error {
	valueInt, err := strconv.ParseInt(string(value), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid counter value %s: %v", value, err)
	}

	m.Lock()
	m.counters[string(name)] += valueInt
	m.Unlock()
	return nil
}

func (m *Metrics) processGauge(name, value []byte) error {
	valueInt, err := strconv.ParseInt(string(value), 10, 64)
	if err != nil {
		return fmt.Errorf("invalid counter value %s: %v", value, err)
	}

	m.Lock()
	m.gauges[string(name)] = valueInt
	m.Unlock()
	return nil
}

func (m *Metrics) processTimer(name, value []byte) error {
	valueFloat, err := strconv.ParseFloat(string(value), 64)
	if err != nil {
		return fmt.Errorf("invalid counter value %s: %v", value, err)
	}

	now := time.Now()
	sName := string(name)

	m.Lock()
	m.timers[sName] = append(m.timers[sName], timer{
		time:  now,
		value: valueFloat,
	})
	m.Unlock()
	return nil
}

func (m *Metrics) process(name, value, stype []byte) error {
	switch string(stype) {
	case "c":
		return m.processCounter(name, value)
	case "g":
		return m.processGauge(name, value)
	case "ms":
		return m.processTimer(name, value)
	default:
		return fmt.Errorf("Unknown metric type: %v", string(stype))
	}
}

func packetTokenizer(packet []byte, end byte) ([]byte, []byte, error) {
	endIndex := bytes.IndexByte(packet, end)
	if endIndex < 0 {
		return nil, packet, fmt.Errorf("cannot find '%c' in packet: %s", end, packet)
	}
	return packet[endIndex+1:], packet[:endIndex], nil
}

func dumpPacket(packet []byte) {
	fmt.Printf("Packet (length: %v)\n", len(packet))
	dumper := hex.Dumper(os.Stdout)
	dumper.Write(packet)
	dumper.Close()
}

func (m *Metrics) processPacket(packet []byte) error {
	for len(packet) > 0 {
		var key, value, statType []byte
		var err error

		if packet, key, err = packetTokenizer(packet, ':'); err != nil {
			return err
		}
		if packet, value, err = packetTokenizer(packet, '|'); err != nil {
			return err
		}

		// The new line at the end is not always present, so we can ignore errors.
		packet, statType, err = packetTokenizer(packet, '\n')

		if err := m.process(key, value, statType); err != nil {
			return err
		}
	}

	return nil
}
