package balance

import (
    "fmt"
    "sync"
)

type Account struct {
    Address [33]byte
    Balance uint64
    Nonce   uint64
}

type Manager struct {
    mu       sync.RWMutex
    accounts map[string]*Account
}

func NewManager() *Manager {
    return &Manager{
        accounts: make(map[string]*Account),
    }
}

func (m *Manager) SetBalance(addr [33]byte, balance uint64) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    key := fmt.Sprintf("%x", addr)
    if acc, exists := m.accounts[key]; exists {
        acc.Balance = balance
    } else {
        m.accounts[key] = &Account{
            Address: addr,
            Balance: balance,
            Nonce:   0,
        }
    }
}

func (m *Manager) GetBalance(addr [33]byte) uint64 {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    key := fmt.Sprintf("%x", addr)
    if acc, exists := m.accounts[key]; exists {
        return acc.Balance
    }
    return 0
}

func (m *Manager) Transfer(from, to [33]byte, amount uint64) error {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    fromKey := fmt.Sprintf("%x", from)
    toKey := fmt.Sprintf("%x", to)
    
    fromAcc, exists := m.accounts[fromKey]
    if !exists {
        return fmt.Errorf("sender account not found")
    }
    
    if fromAcc.Balance < amount {
        return fmt.Errorf("insufficient balance: %d < %d", fromAcc.Balance, amount)
    }
    
    fromAcc.Balance -= amount
    
    if _, exists := m.accounts[toKey]; !exists {
        m.accounts[toKey] = &Account{
            Address: to,
            Balance: 0,
            Nonce:   0,
        }
    }
    m.accounts[toKey].Balance += amount
    
    return nil
}

func (m *Manager) GetNonce(addr [33]byte) uint64 {
    m.mu.RLock()
    defer m.mu.RUnlock()
    
    key := fmt.Sprintf("%x", addr)
    if acc, exists := m.accounts[key]; exists {
        return acc.Nonce
    }
    return 0
}

func (m *Manager) IncrementNonce(addr [33]byte) {
    m.mu.Lock()
    defer m.mu.Unlock()
    
    key := fmt.Sprintf("%x", addr)
    if acc, exists := m.accounts[key]; exists {
        acc.Nonce++
    }
}