package payment

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "time"
    
    "github.com/decred/dcrd/dcrec/secp256k1/v4"
    "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

type Payment struct {
    From      [33]byte
    To        [33]byte
    Amount    uint64
    Nonce     uint64
    Timestamp uint64
    Signature [64]byte
    Hash      [32]byte
}

func NewPayment(fromPrivKey *secp256k1.PrivateKey, toPubKey *secp256k1.PublicKey, amount uint64, nonce uint64) *Payment {
    fromPubKey := fromPrivKey.PubKey()
    
    var fromBytes [33]byte
    fromSlice := fromPubKey.SerializeCompressed()
    copy(fromBytes[:], fromSlice)
    
    var toBytes [33]byte
    toSlice := toPubKey.SerializeCompressed()
    copy(toBytes[:], toSlice)
    
    payment := &Payment{
        From:      fromBytes,
        To:        toBytes,
        Amount:    amount,
        Nonce:     nonce,
        Timestamp: uint64(time.Now().Unix()),
    }
    
    payment.Hash = payment.ComputeHash()
    
    signature := ecdsa.Sign(fromPrivKey, payment.Hash[:])
    copy(payment.Signature[:], signature.Serialize())
    
    return payment
}

func (p *Payment) ComputeHash() [32]byte {
    data := make([]byte, 33+33+8+8+8)
    offset := 0
    
    copy(data[offset:offset+33], p.From[:])
    offset += 33
    copy(data[offset:offset+33], p.To[:])
    offset += 33
    
    for i := 0; i < 8; i++ {
        data[offset+i] = byte(p.Amount >> (56 - uint(i)*8))
    }
    offset += 8
    
    for i := 0; i < 8; i++ {
        data[offset+i] = byte(p.Nonce >> (56 - uint(i)*8))
    }
    offset += 8
    
    for i := 0; i < 8; i++ {
        data[offset+i] = byte(p.Timestamp >> (56 - uint(i)*8))
    }
    
    return sha256.Sum256(data)
}

func (p *Payment) Verify() bool {
    fromPubKey, err := secp256k1.ParsePubKey(p.From[:])
    if err != nil {
        return false
    }
    
    signature, err := ecdsa.ParseDERSignature(p.Signature[:])
    if err != nil {
        return false
    }
    
    return signature.Verify(p.Hash[:], fromPubKey)
}

func (p *Payment) String() string {
    from := hex.EncodeToString(p.From[:])[:12]
    to := hex.EncodeToString(p.To[:])[:12]
    return fmt.Sprintf("Payment{from:%s..., to:%s..., amount:%d}", from, to, p.Amount)
}
