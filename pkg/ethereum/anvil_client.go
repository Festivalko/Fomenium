package ethereum

import (
    "context"
    "crypto/ecdsa"
    "math/big"
    "time"

    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/core/types"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
)

type AnvilClient struct {
    client      *ethclient.Client
    chainID     *big.Int
    privateKey  *ecdsa.PrivateKey
    fromAddress common.Address
}

type TransactionReceipt struct {
    TxHash      common.Hash
    GasUsed     uint64
    BlockNumber uint64
    Status      uint64
    Duration    time.Duration
}

func NewAnvilClient(rpcURL, privateKeyHex string) (*AnvilClient, error) {
    client, err := ethclient.Dial(rpcURL)
    if err != nil {
        return nil, err
    }

    chainID, err := client.ChainID(context.Background())
    if err != nil {
        client.Close()
        return nil, err
    }

    privateKey, err := crypto.HexToECDSA(privateKeyHex)
    if err != nil {
        client.Close()
        return nil, err
    }

    fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

    return &AnvilClient{
        client:      client,
        chainID:     chainID,
        privateKey:  privateKey,
        fromAddress: fromAddress,
    }, nil
}

func (a *AnvilClient) SendTransaction(to common.Address, amount *big.Int) (*TransactionReceipt, error) {
    start := time.Now()
    ctx := context.Background()

    nonce, err := a.client.PendingNonceAt(ctx, a.fromAddress)
    if err != nil {
        return nil, err
    }

    gasPrice, err := a.client.SuggestGasPrice(ctx)
    if err != nil {
        return nil, err
    }

    tx := types.NewTransaction(nonce, to, amount, 21000, gasPrice, nil)

    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(a.chainID), a.privateKey)
    if err != nil {
        return nil, err
    }

    err = a.client.SendTransaction(ctx, signedTx)
    if err != nil {
        return nil, err
    }

    time.Sleep(2 * time.Second)

    receipt, err := a.client.TransactionReceipt(ctx, signedTx.Hash())
    if err != nil {
        return nil, err
    }

    return &TransactionReceipt{
        TxHash:      signedTx.Hash(),
        GasUsed:     receipt.GasUsed,
        BlockNumber: receipt.BlockNumber.Uint64(),
        Status:      receipt.Status,
        Duration:    time.Since(start),
    }, nil
}

func (a *AnvilClient) Close() {
    if a.client != nil {
        a.client.Close()
    }
}
