package main

import (
	"math/big"
	"sync"
)

type NonceManager struct {
	lock          sync.Mutex
	nonceMemCache map[string]*big.Int
}

func NewNonceManager() *NonceManager {
	return &NonceManager{
		lock: sync.Mutex{},
	}
}

func (m *NonceManager) SetNonce(address string, nonce *big.Int) {
	if m.nonceMemCache == nil {
		m.nonceMemCache = make(map[string]*big.Int)
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	m.nonceMemCache[address] = nonce
}

func (m *NonceManager) GetNonce(address string) *big.Int {
	if m.nonceMemCache == nil {
		m.nonceMemCache = make(map[string]*big.Int)
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.nonceMemCache[address]
}

func (m *NonceManager) PlusNonce(address string) {
	if m.nonceMemCache == nil {
		m.nonceMemCache = make(map[string]*big.Int)
	}
	m.lock.Lock()
	defer m.lock.Unlock()
	oldNonce := m.nonceMemCache[address]
	newNonce := oldNonce.Add(oldNonce, big.NewInt(int64(1)))
	m.nonceMemCache[address] = newNonce
}
