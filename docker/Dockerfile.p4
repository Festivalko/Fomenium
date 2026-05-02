FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY <<'GOCODE' main.go
package main

import (
    "encoding/binary"
    "log"
    "net"
    "sync"
    "time"
)

type BatchStats struct {
    Sum     uint64
    Count   uint32
    LastTx  time.Time
}

type P4Simulator struct {
    batches map[uint32]*BatchStats
    mu      sync.RWMutex
}

func NewP4Simulator() *P4Simulator {
    return &P4Simulator{
        batches: make(map[uint32]*BatchStats),
    }
}

func (p *P4Simulator) processPacket(conn net.Conn) {
    defer conn.Close()
    
    for {
        packet := make([]byte, 29)
        n, err := conn.Read(packet)
        if err != nil || n < 29 {
            return
        }
        
        batchID := binary.BigEndian.Uint32(packet[0:4])
        paymentID := binary.BigEndian.Uint32(packet[4:8])
        fromAcc := binary.BigEndian.Uint32(packet[8:12])
        toAcc := binary.BigEndian.Uint32(packet[12:16])
        amount := binary.BigEndian.Uint64(packet[16:24])
        status := packet[24]
        
        p.mu.Lock()
        if p.batches[batchID] == nil {
            p.batches[batchID] = &BatchStats{
                LastTx: time.Now(),
            }
        }
        batch := p.batches[batchID]
        if status == 1 {
            batch.Sum += amount
            batch.Count++
        }
        batch.LastTx = time.Now()
        sum := batch.Sum
        p.mu.Unlock()
        
        log.Printf("[P4] Batch=%d PID=%d From=%d To=%d Amount=%d Valid=%d Sum=%d", 
            batchID, paymentID, fromAcc, toAcc, amount, status, sum)
        
        resp := make([]byte, 8)
        binary.BigEndian.PutUint64(resp, sum)
        conn.Write(resp)
        
        time.Sleep(500 * time.Microsecond)
    }
}

func (p *P4Simulator) statsHandler() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        p.mu.RLock()
        log.Printf("\n=== P4 Switch Statistics ===")
        log.Printf("Active batches: %d", len(p.batches))
        for id, stats := range p.batches {
            log.Printf("  Batch %d: %d txns, sum=%d", id, stats.Count, stats.Sum)
        }
        p.mu.RUnlock()
    }
}

func main() {
    p4 := NewP4Simulator()
    
    go p4.statsHandler()
    
    ln, err := net.Listen("tcp", "0.0.0.0:9090")
    if err != nil {
        log.Fatalf("Failed to bind: %v", err)
    }
    defer ln.Close()
    
    log.Println("=== P4 Switch Emulator ===")
    log.Println("Port: 9090")
    log.Println("Processing packets...")
    
    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Printf("Accept error: %v", err)
            continue
        }
        go p4.processPacket(conn)
    }
}
GOCODE

RUN go build -o p4switch main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates netcat-openbsd
WORKDIR /app
COPY --from=builder /app/p4switch .
EXPOSE 9090
HEALTHCHECK --interval=5s --timeout=3s --retries=3 \
    CMD nc -z localhost 9090 || exit 1
CMD ["./p4switch"]
