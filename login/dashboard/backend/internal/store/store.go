package store

import "sync"

// TokenStore 是一个线程安全的内存存储，用于保存 user -> slurm_token 的映射
type TokenStore struct {
	data map[string]string
	mu   *sync.RWMutex
}

func NewTokenStore() *TokenStore {
	return &TokenStore{
		data: make(map[string]string),
		mu:   &sync.RWMutex{},
	}
}

func (s *TokenStore) Set(username, slurmToken string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[username] = slurmToken
}

func (s *TokenStore) Get(username string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	token, ok := s.data[username]
	return token, ok
}
