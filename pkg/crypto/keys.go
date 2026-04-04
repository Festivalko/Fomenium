package crypto

import (
    "encoding/hex"
    
    "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

func GenerateKeyPair() (*secp256k1.PrivateKey, *secp256k1.PublicKey, error) {
    privKey, err := secp256k1.GeneratePrivateKey()
    if err != nil {
        return nil, nil, err
    }
    return privKey, privKey.PubKey(), nil
}

func PrivateKeyFromHex(hexStr string) (*secp256k1.PrivateKey, error) {
    bytes, err := hex.DecodeString(hexStr)
    if err != nil {
        return nil, err
    }
    return secp256k1.PrivKeyFromBytes(bytes), nil
}

func PublicKeyFromHex(hexStr string) (*secp256k1.PublicKey, error) {
    bytes, err := hex.DecodeString(hexStr)
    if err != nil {
        return nil, err
    }
    return secp256k1.ParsePubKey(bytes)
}

func PublicKeyToString(pubKey *secp256k1.PublicKey) string {
    return hex.EncodeToString(pubKey.SerializeCompressed())
}

func AddressFromPubKey(pubKey *secp256k1.PublicKey) string {
    hash := pubKey.SerializeCompressed()
    return hex.EncodeToString(hash[:20])
}