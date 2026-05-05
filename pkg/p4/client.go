package p4

import (
    "encoding/binary"
    "fmt"
    "log"
    "net"
    "sync"
    "time"
)

type P4Client struct {
    conn         net.Conn
    address      string
    connected    bool
    packetsSent  int
    totalLatency time.Duration
    batchSums    map[uint32]uint64
    batchCounts  map[uint32]uint32
    mu           sync.Mutex
}

func NewP4Client(addr string) (*P4Client, error) {
    client := &P4Client{
        address:     addr,
        batchSums:   make(map[uint32]uint64),
        batchCounts: make(map[uint32]uint32),
    }

    conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
    if err != nil {
        log.Printf("P4 switch not reachable at %s, using simulation", addr)
        return client, nil
    }

    client.conn = conn
    client.connected = true
    log.Printf("Connected to P4 bmv2 switch at %s", addr)
    return client, nil
}

func NewP4Aggregator(addr string) (*P4Client, error) {
    return NewP4Client(addr)
}

func (p *P4Client) AggregatePayment(batchID, paymentID, fromAcc, toAcc uint32, amount uint64, isValid uint8) error {
    start := time.Now()
    defer func() {
        p.mu.Lock()
        p.packetsSent++
        p.totalLatency += time.Since(start)
        p.mu.Unlock()
    }()

    if p.connected {
        return p.sendP4Packet(batchID, paymentID, fromAcc, toAcc, amount, isValid)
    }

    // Симуляция
    p.mu.Lock()
    p.batchSums[batchID] += amount
    p.batchCounts[batchID]++
    p.mu.Unlock()
    return nil
}

func (p *P4Client) sendP4Packet(batchID, paymentID, fromAcc, toAcc uint32, amount uint64, isValid uint8) error {
    // Формируем пакет точно как в P4 программе: payment_h
    packet := make([]byte, 25) // 4+4+4+4+8+1 = 25 байт
    
    binary.BigEndian.PutUint32(packet[0:4], batchID)
    binary.BigEndian.PutUint32(packet[4:8], paymentID)
    binary.BigEndian.PutUint32(packet[8:12], fromAcc)
    binary.BigEndian.PutUint32(packet[12:16], toAcc)
    binary.BigEndian.PutUint64(packet[16:24], amount)
    packet[24] = isValid

    _, err := p.conn.Write(packet)
    if err != nil {
        log.Printf("P4 write failed, switching to simulation: %v", err)
        p.conn.Close()
        p.connected = false
        p.mu.Lock()
        p.batchSums[batchID] += amount
        p.batchCounts[batchID]++
        p.mu.Unlock()
        return nil
    }

    // Читаем ответ (агрегированную сумму)
    resp := make([]byte, 8)
    p.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
    if n, _ := p.conn.Read(resp); n == 8 {
        sum := binary.BigEndian.Uint64(resp)
        p.mu.Lock()
        p.batchSums[batchID] = sum
        p.mu.Unlock()
    }

    return nil
}

func (p *P4Client) SendPacket(batchID, paymentID, fromAcc, toAcc uint32, amount uint64) error {
    return p.AggregatePayment(batchID, paymentID, fromAcc, toAcc, amount, 1)
}

func (p *P4Client) GetBatchResult(batchID uint32) (uint64, uint32) {
    p.mu.Lock()
    defer p.mu.Unlock()
    return p.batchSums[batchID], p.batchCounts[batchID]
}

func (p *P4Client) GetStats() (int, time.Duration) {
    p.mu.Lock()
    defer p.mu.Unlock()
    if p.packetsSent == 0 {
        return 0, 0
    }
    return p.packetsSent, p.totalLatency / time.Duration(p.packetsSent)
}

func (p *P4Client) IsReal() bool {
    return p.connected
}

func (p *P4Client) PrintStats() {
    p.mu.Lock()
    defer p.mu.Unlock()

    mode := "SIMULATION"
    if p.connected {
        mode = "P4 bmv2"
    }

    fmt.Printf("=== P4 Aggregator (%s) ===\n", mode)
    fmt.Printf("Packets: %d\n", p.packetsSent)
    if p.packetsSent > 0 {
        fmt.Printf("Avg latency: %v\n", p.totalLatency/time.Duration(p.packetsSent))
    }
}

func (p *P4Client) Close() error {
    if p.conn != nil {
        return p.conn.Close()
    }
    return nil
}
