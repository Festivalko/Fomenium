package p4

import (
    "encoding/binary"
    "fmt"
    "log"
    "net"
    "time"
)

type P4Client struct {
    conn         net.Conn
    address      string
    connected    bool
    simMode      bool
    packetsSent  int
    totalLatency time.Duration
    batchSums    map[uint32]uint64
    batchCounts  map[uint32]uint32
}

func NewP4Client(addr string) (*P4Client, error) {
    client := &P4Client{
        address:     addr,
        batchSums:   make(map[uint32]uint64),
        batchCounts: make(map[uint32]uint32),
    }
    
    conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
    if err != nil {
        log.Printf("⚠️ P4 switch not reachable at %s, using SIM mode", addr)
        client.simMode = true
        return client, nil
    }
    
    client.conn = conn
    client.connected = true
    log.Printf("✅ Connected to P4 switch at %s", addr)
    return client, nil
}

func NewP4Aggregator(addr string) (*P4Client, error) {
    return NewP4Client(addr)
}

func (p *P4Client) AggregatePayment(batchID, paymentID, fromAcc, toAcc uint32, amount uint64, isValid uint8) error {
    return p.SendPacket(batchID, paymentID, fromAcc, toAcc, amount)
}

func (p *P4Client) PrintStats() {
    fmt.Println("=== P4 Aggregator Stats ===")
    fmt.Printf("Real P4: %t | Packets: %d\n", p.IsReal(), p.packetsSent)
    
    sent, latency := p.GetStats()
    if sent > 0 {
        fmt.Printf("Avg latency: %v\n", latency)
    }
    
    fmt.Println("Batch sums:")
    for id, sum := range p.batchSums {
        count := p.batchCounts[id]
        fmt.Printf("  Batch %d: sum=%d, count=%d\n", id, sum, count)
    }
}

func (p *P4Client) SendPacket(batchID, paymentID, fromAcc, toAcc uint32, amount uint64) error {
    start := time.Now()
    
    packet := make([]byte, 14+25)
    
    copy(packet[0:6], []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x01})
    copy(packet[6:12], []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x02})
    binary.BigEndian.PutUint16(packet[12:14], 0x0800)
    
    offset := 14
    binary.BigEndian.PutUint32(packet[offset:offset+4], batchID); offset += 4
    binary.BigEndian.PutUint32(packet[offset:offset+4], paymentID); offset += 4
    binary.BigEndian.PutUint32(packet[offset:offset+4], fromAcc); offset += 4
    binary.BigEndian.PutUint32(packet[offset:offset+4], toAcc); offset += 4
    binary.BigEndian.PutUint64(packet[offset:offset+8], amount); offset += 8
    packet[offset] = 1
    
    if p.connected {
        _, err := p.conn.Write(packet)
        if err != nil {
            p.connected = false
            p.simMode = true
            return err
        }
        resp := make([]byte, 8)
        p.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
        if n, err := p.conn.Read(resp); err == nil && n == 8 {
            sum := binary.BigEndian.Uint64(resp)
            p.batchSums[batchID] = sum
        }
    } else {
        p.batchSums[batchID] += amount
        p.batchCounts[batchID]++
    }
    
    p.packetsSent++
    p.totalLatency += time.Since(start)
    return nil
}

func (p *P4Client) GetBatchResult(batchID uint32) (uint64, uint32) {
    return p.batchSums[batchID], p.batchCounts[batchID]
}

func (p *P4Client) GetStats() (int, time.Duration) {
    if p.packetsSent == 0 {
        return 0, 0
    }
    return p.packetsSent, p.totalLatency / time.Duration(p.packetsSent)
}

func (p *P4Client) IsReal() bool {
    return p.connected && !p.simMode
}

func (p *P4Client) Close() error {
    if p.conn != nil {
        return p.conn.Close()
    }
    return nil
}
