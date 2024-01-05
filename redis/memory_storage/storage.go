package memory_storage

import "sync"

type storage struct {
	storage sync.Map
}

func New() *storage {
	return &storage{
		storage: sync.Map{},
	}
}

func (s *storage) Set(key string, value any) {
	s.storage.Store(key, value)
}
func (s *storage) Get(key string) (any, bool) {
	value, ok := s.storage.Load(key)
	return value, ok
}
