package main

import (
    "context"
    "encoding/json"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/yourusername/nexus-l2/pkg/p4"
)

type Multiplexer struct {
    p4 *p4.P4Client
}

type Payment struct {
    BatchID    uint32 `json:"batch_id"`
    PaymentID  uint32 `json:"payment_id"`
    FromAcc    uint32 `json:"from_account"`
    ToAcc      uint32 `json:"to_account"`
    Amount     uint64 `json:"amount"`
    Hash       string `json:"hash"`
    IsValid    uint8  `json:"is_valid"`
}

func main() {
    addr := os.Getenv("P4_SWITCH_ADDR")
    if addr == "" {
        addr = "p4-switch:9090"
    }
    
    p4Client, err := p4.NewP4Aggregator(addr)
    if err != nil {
        log.Fatalf("P4 connection failed: %v", err)
    }
    defer p4Client.Close()
    
    m := &Multiplexer{p4: p4Client}
    
    http.HandleFunc("/aggregate", m.handleAggregate)
    http.HandleFunc("/stats", m.handleStats)
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
    })
    
    log.Printf("✅ Multiplexer started on :8080")
    log.Printf("   P4 real mode: %t, addr: %s", p4Client.IsReal(), addr)
    
    srv := &http.Server{Addr: ":8080"}
    
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed: %v", err)
        }
    }()
    
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan
    
    log.Println("Shutting down...")
    p4Client.PrintStats()
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    srv.Shutdown(ctx)
}

func (m *Multiplexer) handleAggregate(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "POST only", http.StatusMethodNotAllowed)
        return
    }
    
    var payment Payment
    if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
        log.Printf("❌ Failed to decode JSON: %v", err)
        http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
        return
    }
    
    log.Printf("📥 Received: batch=%d payment=%d amount=%d", payment.BatchID, payment.PaymentID, payment.Amount)
    
    start := time.Now()
    err := m.p4.AggregatePayment(
        payment.BatchID, payment.PaymentID,
        payment.FromAcc, payment.ToAcc,
        payment.Amount, payment.IsValid,
    )
    
    if err != nil {
        log.Printf("❌ P4 error: %v", err)
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    log.Printf("✅ Aggregated in %v", time.Since(start))
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "status":     "aggregated",
        "batch_id":   payment.BatchID,
        "payment_id": payment.PaymentID,
        "hash":       payment.Hash,
        "latency_ms": time.Since(start).Milliseconds(),
    })
}

func (m *Multiplexer) handleStats(w http.ResponseWriter, r *http.Request) {
    m.p4.PrintStats()
    
    sent, latency := m.p4.GetStats()
    stats := map[string]interface{}{
        "real_p4":      m.p4.IsReal(),
        "packets_sent": sent,
        "avg_latency":  latency.String(),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(stats)
}
