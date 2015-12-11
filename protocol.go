package statsd

import (
	"bytes"
	"fmt"
	"log"
	"net"
)

// Start starts a statsd server on the given hostPort.
func Start(hostPort string) (*Metrics, *net.UDPAddr, error) {
	addr, err := net.ResolveUDPAddr("udp", hostPort)
	if err != nil {
		return nil, nil, err
	}

	udpConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, nil, err
	}

	m := newMetrics()
	go m.handleConn(udpConn)
	return m, udpConn.LocalAddr().(*net.UDPAddr), nil
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
		m.callOnUpdate()
	}
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

func (m *Metrics) process(name, value, stype []byte) error {
	if len(name) == 0 {
		return fmt.Errorf("missimg name")
	}
	if len(value) == 0 {
		return fmt.Errorf("missimg value")
	}

	switch string(stype) {
	case "c":
		return m.processCounter(name, value)
	case "g":
		return m.processGauge(name, value)
	case "ms":
		return m.processTimer(name, value)
	default:
		return fmt.Errorf("unknown metric type: %v", string(stype))
	}
}
func packetTokenizer(packet []byte, end byte) ([]byte, []byte, error) {
	endIndex := bytes.IndexByte(packet, end)
	if endIndex < 0 {
		return nil, packet, fmt.Errorf("cannot find '%c' in packet: %s", end, packet)
	}
	return packet[endIndex+1:], packet[:endIndex], nil
}

// func dumpPacket(packet []byte) {
// 	fmt.Printf("Packet (length: %v)\n", len(packet))
// 	dumper := hex.Dumper(os.Stdout)
// 	dumper.Write(packet)
// 	dumper.Close()
// }
