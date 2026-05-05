package main

import (
    "bytes"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "time"
)

func main() {
    var (
        server   = flag.String("server", "http://localhost:8080", "Multiplexer HTTP API")
        amount   = flag.Uint64("amount", 0, "Amount")
        count    = flag.Int("count", 1, "Number of payments")
        fromAcc  = flag.Int("from", 1, "From account ID")
        toAcc    = flag.Int("to", 2, "To account ID")
    )
    flag.Parse()

    if *amount == 0 {
        fmt.Println("Usage: client -amount <n> [-count <n>] [-from <id>] [-to <id>]")
        os.Exit(1)
    }

    fmt.Printf("Sending %d payment(s) of %d tokens\n", *count, *amount)
    fmt.Println("--------------------------------------------------")

    totalTime := time.Duration(0)

    for i := 0; i < *count; i++ {
        start := time.Now()
        
        reqBody := map[string]interface{}{
            "batch_id":     1,
            "payment_id":   i + 1,
            "from_account": *fromAcc,
            "to_account":   *toAcc,
            "amount":       *amount,
            "hash":         fmt.Sprintf("tx-%d-%d", time.Now().UnixNano(), i),
        }

        jsonData, _ := json.Marshal(reqBody)

        resp, err := http.Post(*server+"/aggregate", "application/json", bytes.NewBuffer(jsonData))
        if err != nil {
            log.Printf("Failed to send payment %d: %v", i+1, err)
            continue
        }

        var result map[string]interface{}
        json.NewDecoder(resp.Body).Decode(&result)
        resp.Body.Close()

        elapsed := time.Since(start)
        totalTime += elapsed

        status := "OK"
        if result["status"] != "ok" {
            status = "FAIL"
        }
        fmt.Printf("  Payment %d: %v - %s\n", i+1, elapsed, status)
    }

    avgTime := totalTime / time.Duration(*count)
    fmt.Println("--------------------------------------------------")
    fmt.Printf("Average time: %v\n", avgTime)
    if avgTime > 0 {
        fmt.Printf("Throughput: %.0f TPS\n", float64(time.Second)/float64(avgTime))
    }
}
