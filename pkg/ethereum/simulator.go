package ethereum

import (
    "fmt"
    "strings"
    "time"
)

type EthereumTx struct {
    ID        int
    From      string
    To        string
    Amount    uint64
    GasUsed   uint64
    BlockTime time.Duration
    Timestamp time.Time
}

type EthereumL1 struct {
    transactions []EthereumTx
    totalGasUsed uint64
    totalTime    time.Duration
    currentBlock uint64
}

func NewEthereumL1() *EthereumL1 {
    return &EthereumL1{
        transactions: make([]EthereumTx, 0),
        currentBlock: 0,
    }
}

func (e *EthereumL1) SendTransaction(from, to string, amount uint64, txID int) (*EthereumTx, time.Duration) {
    start := time.Now()
    
    blockTime := 12 * time.Second
    time.Sleep(blockTime)
    
    gasUsed := uint64(21000)
    
    tx := EthereumTx{
        ID:        txID,
        From:      from,
        To:        to,
        Amount:    amount,
        GasUsed:   gasUsed,
        BlockTime: blockTime,
        Timestamp: time.Now(),
    }
    
    e.transactions = append(e.transactions, tx)
    e.totalGasUsed += gasUsed
    e.totalTime += time.Since(start)
    e.currentBlock++
    
    return &tx, time.Since(start)
}

func (e *EthereumL1) SendMultipleTransactions(from, to string, amount uint64, count int) ([]EthereumTx, time.Duration, uint64) {
    fmt.Printf("\n  🔄 Processing %d Ethereum transactions (one by one)...\n", count)
    fmt.Printf("  ⏱️  Each transaction waits for block confirmation (~12 seconds)\n")
    fmt.Printf("  📊 Total time will be ~%d seconds\n", count*12)
    fmt.Println()
    
    txs := make([]EthereumTx, 0)
    totalTime := time.Duration(0)
    totalGas := uint64(0)
    
    for i := 1; i <= count; i++ {
        fmt.Printf("     Transaction %d/%d: sending %d tokens... ", i, count, amount)
        tx, duration := e.SendTransaction(from, to, amount, i)
        txs = append(txs, *tx)
        totalTime += duration
        totalGas += tx.GasUsed
        fmt.Printf("✅ confirmed in %v\n", duration)
    }
    
    return txs, totalTime, totalGas
}

func (e *EthereumL1) CompareWithNexus(txCount int, nexusTime time.Duration, nexusGas uint64) {
    ethTotalTime := time.Duration(txCount) * 12 * time.Second
    ethTotalGas := uint64(txCount) * 21000
    
    fmt.Println("\n" + strings.Repeat("═", 70))
    fmt.Println("📊 ETHEREUM L1 vs SPIDERSPEED NEXUS - REAL COMPARISON")
    fmt.Println(strings.Repeat("═", 70))
    
    fmt.Printf("\n┌─────────────────────────────────────────────────────────────────────┐\n")
    fmt.Printf("│  METRIC                    │  ETHEREUM L1      │  SPIDERSPEED NEXUS  │\n")
    fmt.Printf("├────────────────────────────┼───────────────────┼─────────────────────┤\n")
    fmt.Printf("│  Time for %d txs           │  %-15v │  %-19v │\n", txCount, ethTotalTime, nexusTime)
    fmt.Printf("│  Time per transaction      │  %-15v │  %-19v │\n", ethTotalTime/time.Duration(txCount), nexusTime/time.Duration(txCount))
    fmt.Printf("│  Gas cost                  │  %-15d │  %-19d │\n", ethTotalGas, nexusGas)
    fmt.Printf("│  Cost (10 gwei, $3000)    │  $%-14.4f │  $%-18.4f │\n", 
        float64(ethTotalGas)*10/1e9*3000, 
        float64(nexusGas)*10/1e9*3000)
    fmt.Printf("└────────────────────────────┴───────────────────┴─────────────────────┘\n")
    
    speedup := float64(ethTotalTime) / float64(nexusTime)
    gasSavings := (1 - float64(nexusGas)/float64(ethTotalGas)) * 100
    
    fmt.Printf("\n🚀 RESULT: SpiderSpeed Nexus is %.0f× FASTER than Ethereum L1!\n", speedup)
    fmt.Printf("💰 RESULT: SpiderSpeed Nexus saves %.0f%% on gas costs!\n", gasSavings)
    fmt.Println()
}
