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
    tablesSet    bool
}

func NewP4Client(addr string) (*P4Client, error) {
    client := &P4Client{
        address:     addr,
        batchSums:   make(map[uint32]uint64),
        batchCounts: make(map[uint32]uint32),
    }
    
    conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
    if err != nil {
        log.Printf(" P4 switch not reachable at %s, using SIM mode", addr)
        client.simMode = true
        return client, nil
    }
    
    client.conn = conn
    client.connected = true
    
    // Инициализируем таблицы через Thrift API
    if err := client.setupTables(); err != nil {
        log.Printf(" Failed to setup tables: %v, using SIM mode", err)
        client.connected = false
        client.simMode = true
    }
    
    if client.connected {
        log.Printf(" Connected to P4 switch at %s", addr)
    }
    
    return client, nil
}

func NewP4Aggregator(addr string) (*P4Client, error) {
    return NewP4Client(addr)
}

func (p *P4Client) setupTables() error {
    // Отправляем команду настройки таблиц через Thrift
    setupPacket := []byte{
        0x00, 0x00, 0x00, 0x01, // table_id = 1
        0x00, 0x00, 0x00, 0x01, // action_id = 1 (aggregate)
    }
    
    if p.conn != nil {
        _, err := p.conn.Write(setupPacket)
        if err != nil {
            return err
        }
        
        // Ждем подтверждения
        resp := make([]byte, 4)
        p.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
        if _, err := p.conn.Read(resp); err != nil {
            return err
        }
        
        p.tablesSet = true
    }
    
    return nil
}

func (p *P4Client) AggregatePayment(batchID, paymentID, fromAcc, toAcc uint32, amount uint64, isValid uint8) error {
    start := time.Now()
    defer func() {
        p.packetsSent++
        p.totalLatency += time.Since(start)
    }()
    
    if p.connected && p.tablesSet {
        return p.sendRealPacket(batchID, paymentID, fromAcc, toAcc, amount, isValid)
    }
    
    // Fallback mode
    p.batchSums[batchID] += amount
    p.batchCounts[batchID]++
    time.Sleep(500 * time.Microsecond) // Эмулируем задержку P4
    return nil
}

func (p *P4Client) sendRealPacket(batchID, paymentID, fromAcc, toAcc uint32, amount uint64, isValid uint8) error {
    // Формируем пакет для P4 simple_switch
    packet := make([]byte, 29)
    offset := 0
    
    binary.BigEndian.PutUint32(packet[offset:], batchID); offset += 4
    binary.BigEndian.PutUint32(packet[offset:], paymentID); offset += 4
    binary.BigEndian.PutUint32(packet[offset:], fromAcc); offset += 4
    binary.BigEndian.PutUint32(packet[offset:], toAcc); offset += 4
    binary.BigEndian.PutUint64(packet[offset:], amount); offset += 8
    packet[offset] = isValid
    
    if _, err := p.conn.Write(packet); err != nil {
        log.Printf(" P4 write failed: %v, falling back to SIM", err)
        p.connected = false
        p.simMode = true
        p.batchSums[batchID] += amount
        p.batchCounts[batchID]++
        return nil
    }
    
    // Читаем ответ
    resp := make([]byte, 8)
    p.conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
    if n, err := p.conn.Read(resp); err == nil && n == 8 {
        sum := binary.BigEndian.Uint64(resp)
        p.batchSums[batchID] = sum
    } else {
        // Если не получили ответ, считаем сами
        p.batchSums[batchID] += amount
        p.batchCounts[batchID]++
    }
    
    return nil
}

func (p *P4Client) SendPacket(batchID, paymentID, fromAcc, toAcc uint32, amount uint64) error {
    return p.AggregatePayment(batchID, paymentID, fromAcc, toAcc, amount, 1)
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
    return p.connected && p.tablesSet && !p.simMode
}

func (p *P4Client) PrintStats() {
    fmt.Println("=== P4 Aggregator Stats ===")
    fmt.Printf("Connected: %t | Packets: %d\n", p.connected, p.packetsSent)
    
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

func (p *P4Client) Close() error {
    if p.conn != nil {
        return p.conn.Close()
    }
    return nil
}
