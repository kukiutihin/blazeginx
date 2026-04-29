package storage

import (
	"sync"
	"time"
)

type Storage struct {
	mtex              sync.RWMutex
	defaultExpiration time.Duration
	cleanupInterval   time.Duration
	items             map[string]Item
}

type Item struct {
	Value      any
	Expiration int64
}

func New(
	defaultExpiration,
	cleanupInterval time.Duration,
) *Storage {
	items := make(map[string]Item)
	storage := Storage{
		defaultExpiration: defaultExpiration,
		cleanupInterval:   cleanupInterval,
		items:             items,
	}

	if cleanupInterval > 0 {
		go storage.startGC()
	}

	return &storage
}

func (s *Storage) startGC() {
	ticker := time.NewTicker(s.cleanupInterval)
	for {
		<-ticker.C
		s.deleteExpired()
	}
}

func (s *Storage) deleteExpired() {
	for key, value := range s.items {
		if value.Expiration <= time.Now().UnixNano() {
			s.Delete(key)
		}
	}
}

func (s *Storage) Delete(key string) {
	delete(s.items, key)
}

func (s *Storage) Add(key string, value any) {
	s.AddWithExpiration(key, value, s.defaultExpiration)
}

func (s *Storage) AddWithExpiration(key string, value any, exp time.Duration) {
	expiredAt := time.Now().Add(exp).UnixNano()
	s.mtex.Lock()
	s.items[key] = Item{
		Value:      value,
		Expiration: expiredAt,
	}
	s.mtex.Unlock()
}

func (s *Storage) Get(key string) (any, bool) {
	s.mtex.RLock()
	item, ok := s.items[key]
	if !ok {
		s.mtex.RUnlock()
		return nil, false
	}
	if item.Expiration > 0 {
		if time.Now().UnixNano() > item.Expiration {
			s.mtex.RUnlock()
			return nil, false
		}
	}
	s.mtex.RUnlock()
	return item.Value, true
}

func (s *Storage) Size() int {
	s.mtex.RLock()
	res := len(s.items)
	s.mtex.RUnlock()
	return res
}
