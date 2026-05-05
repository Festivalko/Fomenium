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
)

type BatchRequest struct {
    ID     string
    Count  int
    Amount uint64
}

type BatchResponse struct {
    ID         string
    Valid      bool
    NewBalance uint64
}

type Validator struct {
    port     string
    balances *balance.Manager
    logger   *log.Logger
}

func NewValidator(port string) *Validator {
    return &Validator{
        port:     port,
        balances: balance.NewManager(),
        logger:   log.New(os.Stdout, "[VALIDATOR] ", log.LstdFlags),
    }
}

func (v *Validator) Start() error {
    listener, err := net.Listen("tcp", "0.0.0.0:"+v.port)
    if err != nil {
        return fmt.Errorf("failed to start: %w", err)
    }
    defer listener.Close()

    v.logger.Printf("Validator running on port %s", v.port)
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

    var req BatchRequest
    if err := decoder.Decode(&req); err != nil {
        v.logger.Printf("Failed to decode: %v", err)
        return
    }

    start := time.Now()
    v.logger.Printf("Received batch %s: %d payments, total %d", req.ID, req.Count, req.Amount)

    // Проверяем баланс отправителя (account 1 = alice)
    var fromAddr [33]byte
    copy(fromAddr[:], []byte("alice"))
    balance := v.balances.GetBalance(fromAddr)

    valid := balance >= req.Amount
    newBalance := balance

    if valid {
        var toAddr [33]byte
        copy(toAddr[:], []byte("bob"))
        v.balances.Transfer(fromAddr, toAddr, req.Amount)
        newBalance = v.balances.GetBalance(fromAddr)
        v.logger.Printf("Payment valid: %d tokens, new balance: %d", req.Amount, newBalance)
    } else {
        v.logger.Printf("Insufficient balance: %d < %d", balance, req.Amount)
    }

    elapsed := time.Since(start)

    resp := BatchResponse{
        ID:         req.ID,
        Valid:      valid,
        NewBalance: newBalance,
    }

    encoder.Encode(resp)
    v.logger.Printf("Validated in %v: valid=%t, balance=%d", elapsed, valid, newBalance)
}

func (v *Validator) initTestBalances() {
    testAccounts := []struct {
        name    string
        balance uint64
    }{
        {"alice", 1000000},
        {"bob", 500000},
        {"charlie", 800000},
        {"david", 300000},
    }

    for _, acc := range testAccounts {
        var addr [33]byte
        copy(addr[:], []byte(acc.name))
        v.balances.SetBalance(addr, acc.balance)
        v.logger.Printf("Created account %s with balance %d", acc.name, acc.balance)
    }
    v.logger.Println("Test balances initialized")
}

func main() {
    port := flag.String("port", "8001", "Port")
    flag.Parse()

    validator := NewValidator(*port)
    if err := validator.Start(); err != nil {
        log.Fatal(err)
    }
}
