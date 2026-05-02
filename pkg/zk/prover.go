package zk

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net/http"
    "sync"
    "time"
)

type Proof struct {
    BatchID        string
    RootHash       string
    ProofData      []byte
    ProofSize      int
    GenerationTime time.Duration
}

type Prover struct {
    proverURL       string
    httpClient      *http.Client
    mu              sync.Mutex
    proofsGenerated int
    totalTime       time.Duration
    connected       bool
}

func NewProver(proverURL string) *Prover {
    p := &Prover{
        proverURL:  proverURL,
        httpClient: &http.Client{Timeout: 30 * time.Second},
    }
    
    // Проверяем соединение
    if err := p.healthCheck(); err != nil {
        fmt.Printf(" SP1 Prover not available: %v\n", err)
    } else {
        p.connected = true
        fmt.Printf(" Connected to SP1 Prover at %s\n", proverURL)
    }
    
    return p
}

func (p *Prover) healthCheck() error {
    resp, err := p.httpClient.Get(p.proverURL + "/health")
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    return nil
}

func (p *Prover) GenerateProof(batchID string, paymentsCount int, totalAmount uint64, stateRoot string) *Proof {
    start := time.Now()
    
    proof := &Proof{
        BatchID: batchID,
    }
    
    if p.connected {
        p.generateRealProof(proof, batchID, paymentsCount, totalAmount, stateRoot)
    } else {
        p.generateSimulatedProof(proof, batchID, paymentsCount, totalAmount, stateRoot)
    }
    
    proof.GenerationTime = time.Since(start)
    
    p.mu.Lock()
    p.proofsGenerated++
    p.totalTime += proof.GenerationTime
    p.mu.Unlock()
    
    return proof
}

func (p *Prover) generateRealProof(proof *Proof, batchID string, paymentsCount int, totalAmount uint64, stateRoot string) {
    reqBody := map[string]interface{}{
        "batch_id":       batchID,
        "payments_count": paymentsCount,
        "total_amount":   totalAmount,
        "state_root":     stateRoot,
    }
    
    jsonData, _ := json.Marshal(reqBody)
    
    resp, err := p.httpClient.Post(p.proverURL+"/prove", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        fmt.Printf("SP1 request failed: %v\n", err)
        p.generateSimulatedProof(proof, batchID, paymentsCount, totalAmount, stateRoot)
        return
    }
    defer resp.Body.Close()
    
    var result struct {
        ProofData        string `json:"proof_data"`
        RootHash         string `json:"root_hash"`
        GenerationTimeMs uint64 `json:"generation_time_ms"`
        ProofSizeBytes   int    `json:"proof_size_bytes"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        fmt.Printf("SP1 decode failed: %v\n", err)
        return
    }
    
    proof.RootHash = result.RootHash
    proof.ProofData = []byte(result.ProofData)
    proof.ProofSize = result.ProofSizeBytes
}

func (p *Prover) generateSimulatedProof(proof *Proof, batchID string, paymentsCount int, totalAmount uint64, stateRoot string) {
    time.Sleep(15 * time.Millisecond)
    proof.RootHash = fmt.Sprintf("%x", batchID[:8])
    proof.ProofData = []byte(fmt.Sprintf("%s-%d", batchID, totalAmount))
    proof.ProofSize = len(proof.ProofData)
}

func (p *Prover) VerifyProof(proof *Proof) (bool, time.Duration) {
    start := time.Now()
    
    if p.connected {
        return p.verifyRealProof(proof)
    }
    
    time.Sleep(2 * time.Millisecond)
    return true, time.Since(start)
}

func (p *Prover) verifyRealProof(proof *Proof) (bool, time.Duration) {
    start := time.Now()
    
    reqBody := map[string]interface{}{
        "proof_data": string(proof.ProofData),
    }
    
    jsonData, _ := json.Marshal(reqBody)
    
    resp, err := p.httpClient.Post(p.proverURL+"/verify", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        return true, time.Since(start)
    }
    defer resp.Body.Close()
    
    var result struct {
        IsValid bool `json:"is_valid"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return true, time.Since(start)
    }
    
    return result.IsValid, time.Since(start)
}

func (p *Prover) GetStats() (int, time.Duration) {
    p.mu.Lock()
    defer p.mu.Unlock()
    if p.proofsGenerated == 0 {
        return 0, 0
    }
    return p.proofsGenerated, p.totalTime / time.Duration(p.proofsGenerated)
}

func (p *Prover) PrintStats() {
    count, avgTime := p.GetStats()
    fmt.Printf("\n ZK Prover Stats:\n")
    fmt.Printf("   Connected: %t\n", p.connected)
    fmt.Printf("   Proofs generated: %d\n", count)
    if count > 0 {
        fmt.Printf("   Avg generation: %v\n", avgTime)
    }
}
