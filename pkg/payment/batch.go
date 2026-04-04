package payment

import (
    "crypto/sha256"
    "fmt"
    "time"
    
    "github.com/google/uuid"
)

type Batch struct {
    ID          string
    Payments    []Payment
    TotalAmount uint64
    Count       int
    Timestamp   uint64
    Hash        [32]byte
}

func NewBatch(payments []Payment) *Batch {
    batch := &Batch{
        ID:         uuid.New().String(),
        Payments:   payments,
        TotalAmount: 0,
        Count:      len(payments),
        Timestamp:  uint64(time.Now().Unix()),
    }
    
    for _, p := range payments {
        batch.TotalAmount += p.Amount
    }
    
    batch.Hash = batch.ComputeHash()
    
    return batch
}

func (b *Batch) ComputeHash() [32]byte {
    data := make([]byte, 0)
    data = append(data, []byte(b.ID)...)
    
    for _, p := range b.Payments {
        data = append(data, p.Hash[:]...)
    }
    
    return sha256.Sum256(data)
}

func (b *Batch) String() string {
    return fmt.Sprintf("Batch{id:%s, payments:%d, total:%d}", 
        b.ID[:8], b.Count, b.TotalAmount)
}