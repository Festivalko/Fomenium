package main

import (
    "bytes"
    "encoding/json"
    "flag"
    "fmt"
    "log"
    "net/http"
    "os"
    "strings"
    "time"
    
    "github.com/yourusername/nexus-l2/pkg/crypto"
    "github.com/yourusername/nexus-l2/pkg/payment"
)

func main() {
    var (
        fromHex = flag.String("from", "", "Private key")
        toHex   = flag.String("to", "", "Public key")
        amount  = flag.Uint64("amount", 0, "Amount")
        server  = flag.String("server", "http://localhost:8080", "Multiplexer HTTP API")
        generate = flag.Bool("generate", false, "Generate keys")
        count   = flag.Int("count", 1, "Number of payments")
    )
    flag.Parse()
    
    if *generate {
        generateKeys()
        return
    }
    
    if *fromHex == "" || *toHex == "" || *amount == 0 {
        fmt.Println("Usage:")
        fmt.Println("  Send: client -from <priv> -to <pub> -amount <n> [-count <n>]")
        fmt.Println("  Generate keys: client -generate")
        os.Exit(1)
    }
    
    privKey, _ := crypto.PrivateKeyFromHex(*fromHex)
    pubKey, _ := crypto.PublicKeyFromHex(*toHex)
    
    fmt.Printf("💰 Sending %d payment(s) of %d tokens\n", *count, *amount)
    fmt.Println(strings.Repeat("-", 50))
    
    totalTime := time.Duration(0)
    
    for i := 0; i < *count; i++ {
        nonce := uint64(time.Now().UnixNano())
        p := payment.NewPayment(privKey, pubKey, *amount, nonce)
        
        start := time.Now()
        sendPaymentViaHTTP(p, *server, i+1)
        elapsed := time.Since(start)
        totalTime += elapsed
        
        fmt.Printf("  Payment %d: %v\n", i+1, elapsed)
    }
    
    avgTime := totalTime / time.Duration(*count)
    fmt.Println(strings.Repeat("-", 50))
    fmt.Printf("📈 Average time: %v\n", avgTime)
    fmt.Printf("🚀 Theoretical TPS: %.0f\n", float64(time.Second)/float64(avgTime))
}

func sendPaymentViaHTTP(p *payment.Payment, serverAddr string, id int) {
    // Формируем запрос для мультиплексора
    reqBody := map[string]interface{}{
        "batch_id":    1,
        "payment_id":  id,
        "from_account": 1,
        "to_account":   2,
        "amount":      p.Amount,
        "hash":        fmt.Sprintf("%x", p.Hash[:8]),
    }
    
    jsonData, _ := json.Marshal(reqBody)
    
    resp, err := http.Post(serverAddr+"/aggregate", "application/json", bytes.NewBuffer(jsonData))
    if err != nil {
        log.Printf("Failed to send: %v", err)
        return
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != 200 {
        log.Printf("HTTP error: %d", resp.StatusCode)
    }
}

func generateKeys() {
    priv, pub, _ := crypto.GenerateKeyPair()
    fmt.Println("🔑 NEW KEY PAIR")
    fmt.Println("==================================================")
    fmt.Printf("Private: %x\n", priv.Serialize())
    fmt.Printf("Public:  %x\n", pub.SerializeCompressed())
    fmt.Printf("Address: %s\n", crypto.AddressFromPubKey(pub))
    fmt.Println("==================================================")
}
