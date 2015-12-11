package statsd

import "testing"

func benchmarkForPacket(b *testing.B, packet []byte) {
	m := newMetrics()
	for i := 0; i < b.N; i++ {
		m.processPacket(packet)
	}
}

func BenchmarkProcessPacketOneMetric(b *testing.B) {
	benchmarkForPacket(b, []byte("counter1:1|c"))
}

func BenchmarkProcessPacketMultiMetric(b *testing.B) {
	benchmarkForPacket(b, []byte("counter1:1|c\ncounter2:2|c"))
}
