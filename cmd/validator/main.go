package main

import (
    "encoding/gob"
    "flag"
    "fmt"
    "log"
    "net"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/yourusername/nexus-l2/pkg/balance"
    "github.com/yourusername/nexus-l2/pkg/payment"
)

type Validator struct {
    port        string
    balances    *balance.Manager
    logger      *log.Logger
    txCount     int
    totalTime   time.Duration
}

func NewValidator(port string) *Validator {
    return &Validator{
        port:     port,
        balances: balance.NewManager(),
        logger:   log.New(os.Stdout, "[VALIDATOR] ", log.LstdFlags),
    }
}

func (v *Validator) Start() error {
    listener, err := net.Listen("tcp", "localhost:"+v.port)
    if err != nil {
        return fmt.Errorf("failed to start: %w", err)
    }
    defer listener.Close()
    
    v.logger.Printf("✅ Validator running on port %s", v.port)
    v.initTestBalances()
    
    go v.acceptConnections(listener)
    
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    <-sigChan
    
    v.logger.Println("Shutting down...")
    return nil
}

func (v *Validator) acceptConnections(listener net.Listener) {
    for {
        conn, err := listener.Accept()
        if err != nil {
            continue
        }
        go v.handleConnection(conn)
    }
}

func (v *Validator) handleConnection(conn net.Conn) {
    defer conn.Close()
    
    decoder := gob.NewDecoder(conn)
    encoder := gob.NewEncoder(conn)
    
    var batch payment.Batch
    if err := decoder.Decode(&batch); err != nil {
        v.logger.Printf("Failed to decode: %v", err)
        return
    }
    
    start := time.Now()
    v.logger.Printf("📦 Received batch %s with %d payments", batch.ID[:8], len(batch.Payments))
    
    validPayments := make([]payment.Payment, 0)
    
    for i, p := range batch.Payments {
        if !p.Verify() {
            v.logger.Printf("  ❌ Invalid signature for payment %d", i)
            continue
        }
        
        balance := v.balances.GetBalance(p.From)
        if balance < p.Amount {
            v.logger.Printf("  ❌ Insufficient balance for payment %d", i)
            continue
        }
        
        v.logger.Printf("  ✅ Payment %d is valid", i)
        validPayments = append(validPayments, p)
    }
    
    // Применяем валидные платежи
    for _, p := range validPayments {
        v.balances.Transfer(p.From, p.To, p.Amount)
    }
    
    elapsed := time.Since(start)
    
    response := payment.Batch{
        ID:       batch.ID,
        Payments: validPayments,
        Count:    len(validPayments),
    }
    
    encoder.Encode(response)
    v.logger.Printf("📤 Sent back %d valid payments in %v", len(validPayments), elapsed)
}

func (v *Validator) initTestBalances() {
    testAccounts := []struct {
        name    string
        balance uint64
    }{
        {"alice", 10000},
        {"bob", 5000},
        {"charlie", 8000},
        {"david", 3000},
    }
    
    for _, acc := range testAccounts {
        var addr [33]byte
        copy(addr[:], []byte(acc.name))
        v.balances.SetBalance(addr, acc.balance)
        v.logger.Printf("  Created account %s with balance %d", acc.name, acc.balance)
    }
    v.logger.Println("✅ Test balances initialized")
}

func main() {
    port := flag.String("port", "8001", "Port")
    flag.Parse()
    
    validator := NewValidator(*port)
    if err := validator.Start(); err != nil {
        log.Fatal(err)
    }
}